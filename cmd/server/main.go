package main

import (
		"fmt"
    "context"
    "errors"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

var version = "undefined"

func main() {
    server := &http.Server{
        Addr: ":8080",
    }

    http.Handle("/", http.FileServer(http.Dir("./web/dist/sclipi-web/browser")))
		http.HandleFunc("/health", handleHealth)

    go func() {
				log.Println("Serving on port 8080")
        if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
            log.Fatalf("HTTP server error: %v", err)
        }
        log.Println("Stopped serving new connections.")
    }()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownRelease()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("HTTP shutdown error: %v", err)
    }
    log.Println("Shutdown complete.")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /health")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}

