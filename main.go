package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ulrichheringer/consistency-lab/sim"
)

//go:embed index.html
var page embed.FS

func main() {
	addr := flag.String("addr", ":8090", "HTTP listen address")
	delay := flag.Duration("delay", 3*time.Second, "replica propagation delay")
	flag.Parse()
	model := sim.New(*delay)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, _ := page.ReadFile("index.html")
		_, _ = w.Write(data)
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, _ *http.Request) {
		model.Advance(time.Now())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Snapshot())
	})
	mux.HandleFunc("POST /api/write", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Replica string
			Key     string
			Value   string
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		if err := model.Write(input.Replica, input.Key, input.Value, time.Now()); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	log.Printf("consistency lab listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
