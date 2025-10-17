package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/bhutch29/sclipi/internal/utils"
    "io"
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

    http.HandleFunc("/health", handleHealth)
    http.HandleFunc("/idn", handleIdn)

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

func handleIdn(w http.ResponseWriter, r *http.Request) {
    log.Println("Handling /idn")

    body, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    query := string(body)

    if len(query) == 0 {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, "body cannot be empty")
    }

    inst, err := buildAndConnectInstrument("simulated", "", 10 * time.Second, nil)
    if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
	return
    }

    response, err := inst.Query(query)
    if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
	return
    }
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "%s\n", response)
}

func buildAndConnectInstrument(address string, port string, timeout time.Duration, progressFn func(int)) (utils.Instrument, error) {
    var inst utils.Instrument
    if address == "simulated" {
	inst = utils.NewSimInstrument(timeout)
    } else {
	inst = utils.NewScpiInstrument(timeout)
    }

    if err := inst.Connect(address+":"+port, progressFn); err != nil {
	return inst, err
    }

    return inst, nil
}
