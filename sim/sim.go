package sim

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type VersionedValue struct {
	Value   string `json:"value"`
	Version uint64 `json:"version"`
}

type Replica struct {
	Name   string                    `json:"name"`
	Values map[string]VersionedValue `json:"values"`
}

type Event struct {
	At       time.Time `json:"at"`
	Message  string    `json:"message"`
	Conflict bool      `json:"conflict"`
}

type Snapshot struct {
	Replicas []Replica `json:"replicas"`
	Pending  int       `json:"pending"`
	Events   []Event   `json:"events"`
}

type delivery struct {
	from, to, key string
	value         VersionedValue
	due           time.Time
}

type Simulator struct {
	mu         sync.Mutex
	delay      time.Duration
	next       uint64
	replicas   map[string]map[string]VersionedValue
	deliveries []delivery
	events     []Event
}

func New(delay time.Duration) *Simulator {
	if delay < 0 {
		delay = 0
	}
	return &Simulator{delay: delay, replicas: map[string]map[string]VersionedValue{
		"replica-a": {}, "replica-b": {}, "replica-c": {},
	}}
}

func (s *Simulator) Write(replica, key, value string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	values, ok := s.replicas[replica]
	if !ok {
		return errors.New("unknown replica")
	}
	if key == "" {
		return errors.New("key is required")
	}
	s.next++
	v := VersionedValue{Value: value, Version: s.next}
	values[key] = v
	s.events = append(s.events, Event{At: now, Message: replica + " wrote " + key + "=" + value})
	for name := range s.replicas {
		if name != replica {
			s.deliveries = append(s.deliveries, delivery{from: replica, to: name, key: key, value: v, due: now.Add(s.delay)})
		}
	}
	return nil
}

func (s *Simulator) Advance(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sort.SliceStable(s.deliveries, func(i, j int) bool { return s.deliveries[i].due.Before(s.deliveries[j].due) })
	remaining := s.deliveries[:0]
	for _, d := range s.deliveries {
		if d.due.After(now) {
			remaining = append(remaining, d)
			continue
		}
		local, exists := s.replicas[d.to][d.key]
		conflict := exists && local.Value != d.value.Value
		message := d.from + " -> " + d.to + " delivered " + d.key + "=" + d.value.Value
		if conflict {
			message = d.to + " had " + local.Value + "; compared with " + d.value.Value
		}
		if !exists || d.value.Version > local.Version {
			s.replicas[d.to][d.key] = d.value
			if conflict {
				message += " (newer version kept)"
			}
		} else if conflict {
			message += " (newer local version kept)"
		}
		s.events = append(s.events, Event{At: d.due, Message: message, Conflict: conflict})
	}
	s.deliveries = remaining
}

func (s *Simulator) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := Snapshot{Pending: len(s.deliveries), Events: append([]Event(nil), s.events...)}
	for name, values := range s.replicas {
		copyValues := make(map[string]VersionedValue, len(values))
		for key, value := range values {
			copyValues[key] = value
		}
		out.Replicas = append(out.Replicas, Replica{Name: name, Values: copyValues})
	}
	sort.Slice(out.Replicas, func(i, j int) bool { return out.Replicas[i].Name < out.Replicas[j].Name })
	return out
}
