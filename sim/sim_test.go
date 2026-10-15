package sim

import (
	"testing"
	"time"
)

func TestConcurrentWritesDivergeThenConverge(t *testing.T) {
	now := time.Unix(100, 0)
	s := New(3 * time.Second)
	if err := s.Write("replica-a", "color", "blue", now); err != nil {
		t.Fatal(err)
	}
	if err := s.Write("replica-b", "color", "green", now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	s.Advance(now.Add(2 * time.Second))
	before := s.Snapshot()
	if before.Replicas[0].Values["color"].Value == before.Replicas[1].Values["color"].Value {
		t.Fatal("replicas should diverge before propagation")
	}
	s.Advance(now.Add(5 * time.Second))
	after := s.Snapshot()
	for _, replica := range after.Replicas {
		if replica.Values["color"].Value != "green" {
			t.Fatalf("%s has %q", replica.Name, replica.Values["color"].Value)
		}
	}
	foundConflict := false
	for _, event := range after.Events {
		foundConflict = foundConflict || event.Conflict
	}
	if !foundConflict {
		t.Fatal("expected conflict event")
	}
}
