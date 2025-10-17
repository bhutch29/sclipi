package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/bhutch29/sclipi/internal/utils"
)

var version = "undefined"
var instCache = newInstrumentCache()

type scpiRequestBody struct {
    Type string `json:"type"`
    Scpi string `json:"scpi"`
    Port string `json:"port"`
    Simulated bool `json:"simulated"`
}

func main() {
    server := &http.Server{
        Addr: ":8080",
    }

    http.HandleFunc("/health", handleHealth)
    http.HandleFunc("/scpi", handleScpiRequest)

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

func handleScpiRequest(w http.ResponseWriter, r *http.Request) {
    log.Println("Handling /scpi")
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    bodyData, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    body, err := validateScpiRequestBody(bodyData)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, "%s\n", err.Error())
        return
    }

    address := "localhost"
    if body.Simulated {
        address = "simulated"
    }
    inst, err := instCache.get(address, body.Port, 10 * time.Second, nil)
    if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, "%s\n", err.Error())
	return
    }

    if body.Type == "command" {
        sendCommand(w, inst, body.Scpi)
        return
    }

    if body.Type == "query" {
        sendQuery(w, inst, body.Scpi)
        return
    }
}

func sendCommand(w http.ResponseWriter, inst utils.Instrument, scpi string) {
    err := inst.Command(scpi)
    if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
	return
    }
    w.WriteHeader(http.StatusOK)
}

func sendQuery(w http.ResponseWriter, inst utils.Instrument, scpi string) {
    response, err := inst.Query(scpi)
    if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
	return
    }
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "%s\n", response)
}

func validateScpiRequestBody(bodyData []byte) (scpiRequestBody, error) {
    body := scpiRequestBody{}
    if err := json.Unmarshal(bodyData, &body); err != nil {
        return body, errors.New(fmt.Sprintf("invalid JSON: %v", err))
    }

    if len(body.Scpi) == 0 {
        return body, errors.New("scpi field cannot be empty")
    }
    if len(body.Type) == 0 {
        return body, errors.New("type field cannot be empty")
    }

    if body.Type != "command" && body.Type != "query" {
        return body, errors.New("type must be 'command' or 'query'")
    }

    if len(body.Port) == 0 {
        body.Port = "5025"
    }
    return body, nil
}

