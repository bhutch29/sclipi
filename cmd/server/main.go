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
	"strings"
	"syscall"
	"time"
)

var version = "undefined"
var instCache = newInstrumentCache()

type scpiRequestBody struct {
    Type string `json:"type"`
    Scpi string `json:"scpi"`
    Port string `json:"port"`
    Simulated bool `json:"simulated"`
    AutoSystError bool `json:"autoSystErr"`
    TimeoutSeconds int `json:"timeoutSeconds"`
}

type scpiResponse struct {
    Response string `json:"response"`
    Errors []string `json:"errors"`
    ServerError string `json:"serverError"`
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
        fmt.Fprintln(w, "/scpi only supports POST");
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
    inst, err := instCache.get(address, body.Port, time.Duration(body.TimeoutSeconds) * time.Second, nil)
    if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, "%s\n", err.Error())
	return
    }

    inst.SetTimeout(time.Duration(body.TimeoutSeconds) * time.Second)

    scpiResponse := scpiResponse{}
    if body.Type == "smart" {
        if (strings.Contains(body.Scpi, "?")) {
            queryResponse, err := inst.Query(body.Scpi)
            if (err != nil) {
                log.Printf("Error sending query: %v", err)
                scpiResponse.ServerError = fmt.Sprintf("Error sending query: %v", err)
            } else {
                scpiResponse.Response = queryResponse
            }
        } else {
            err := inst.Command(body.Scpi)
            if err != nil {
                log.Printf("Error sending command: %v", err)
                scpiResponse.ServerError = fmt.Sprintf("Error sending query: %v", err)
            }
        }

        if body.AutoSystError {
            errors, err := inst.QueryError([]string{})
            if err != nil {
                log.Printf("Error doing auto :syst:err?: %v", err)
                scpiResponse.ServerError = fmt.Sprintf("Error sending query: %v", err)
            } else {
                scpiResponse.Errors = errors
            }
        }

        w.WriteHeader(http.StatusOK)
        responseData , _ := json.Marshal(scpiResponse)
        fmt.Fprintf(w, "%s\n", responseData)

        return
    }

    if body.Type == "sendOnly" {
        err := inst.Command(body.Scpi)
        if err != nil {
            log.Printf("Error sending command: %v", err)
            fmt.Fprintf(w, "Error sending command: %v", err)
	    w.WriteHeader(http.StatusInternalServerError)
            return
        }
    }
}

func validateScpiRequestBody(bodyData []byte) (scpiRequestBody, error) {
    body := scpiRequestBody{}
    if err := json.Unmarshal(bodyData, &body); err != nil {
        fmt.Printf("Failed to parse as JSON: %v\n", string(bodyData))
        return body, errors.New(fmt.Sprintf("invalid JSON: %v", err))
    }

    if len(body.Scpi) == 0 {
        fmt.Printf("Received empty SCPI field: %v\n", string(bodyData))
        return body, errors.New("scpi field cannot be empty")
    }

    if len(body.Type) == 0 {
        body.Type = "smart"
    }
    if body.Type != "smart" && body.Type != "sendOnly" {
        return body, errors.New("type must be 'smart' or 'sendOnly'")
    }

    if len(body.Port) == 0 {
        body.Port = "8100" // TODO
    }

    if body.TimeoutSeconds < 0 {
        fmt.Printf("Received negative timeoutSeconds: %v\n", string(bodyData))
        return body, errors.New("timeoutSeconds must be a positive integer")
    }
    if body.TimeoutSeconds == 0 {
        body.TimeoutSeconds = 10
    }
    return body, nil
}
