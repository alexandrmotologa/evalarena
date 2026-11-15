package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexandrmotologa/evalarena/internal/tournament"
)

// HandleSSE streams real-time tournament progress events to web clients.
func HandleSSE(bus *tournament.EventBus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		eventChan := bus.Subscribe()
		defer bus.Unsubscribe(eventChan)

		// Send initial keep-alive
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-eventChan:
				if !ok {
					return
				}
				data, err := json.Marshal(ev.Payload)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(data))
				flusher.Flush()
			}
		}
	}
}
