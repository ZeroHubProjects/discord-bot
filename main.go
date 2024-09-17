package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ZeroOnyxProjects/discord-bot/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.WebhookRequestHandler)
	fmt.Println("Running server...")
	http.ListenAndServe(":14081", recoverMiddleware(mux))
}

func recoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Handler panicked: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(handlers.InternalErrorResponse)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
