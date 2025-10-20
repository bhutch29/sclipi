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

	"github.com/bhutch29/sclipi/internal/utils"
)

var version = "undefined"
var instCache = newInstrumentCache()
var config *Config
var preferences *Preferences

type scpiRequestBody struct {
    Type string `json:"type"`
    Scpi string `json:"scpi"`
    Port int `json:"port"`
    Address string `json:"address"`
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
    var err error
    config, err = loadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    preferences, err = loadPreferences()
    if err != nil {
        log.Fatalf("Failed to load preferences: %v", err)
    } else {
	log.Printf("Loaded preferences: %+v", preferences)
    }

    addr := fmt.Sprintf(":%d", config.ServerPort)
    server := &http.Server{
        Addr: addr,
    }

    http.HandleFunc("/health", handleHealth)
    http.HandleFunc("/scpiPort", handlePort)
    http.HandleFunc("/scpiAddress", handleAddress)
    http.HandleFunc("/scpi", handleScpiRequest)
    http.HandleFunc("/preferences", handlePreferences)

    go func() {
        log.Printf("Serving on port %d", config.ServerPort)
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

func handleAddress(w http.ResponseWriter, r *http.Request) {
    log.Println("Handling /scpiAddress")
    if r.Method == http.MethodGet {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "%v", preferences.ScpiAddress)
    } else if r.Method == http.MethodPost {
        bodyData, err := io.ReadAll(r.Body)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        defer r.Body.Close()

        address := string(bodyData)
        preferences.ScpiAddress = address
        if err := preferences.save(); err != nil {
            log.Printf("Warning: failed to save preferences: %v", err)
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "ScpiAddress updated to %v\n", address)
    } else {
        w.WriteHeader(http.StatusMethodNotAllowed)
        fmt.Fprintf(w, "/scpiAddress supports GET and POST methods\n")
    }
}

func handlePort(w http.ResponseWriter, r *http.Request) {
    log.Println("Handling /scpiPort")
    if r.Method == http.MethodGet {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "%d", preferences.ScpiPort)
    } else if r.Method == http.MethodPost {
        bodyData, err := io.ReadAll(r.Body)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        defer r.Body.Close()

        var port int
        _, err = fmt.Sscanf(string(bodyData), "%d", &port)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "Invalid port number: must be an integer\n")
            return
        }

        if port < 1 || port > 65535 {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "Invalid port number: must be between 1 and 65535\n")
            return
        }

        preferences.ScpiPort = port
        if err := preferences.save(); err != nil {
            log.Printf("Warning: failed to save preferences: %v", err)
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Port updated to %d\n", port)
    } else {
        w.WriteHeader(http.StatusMethodNotAllowed)
        fmt.Fprintf(w, "/scpiPort supports GET and POST methods\n")
    }
}

func handlePreferences(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        w.WriteHeader(http.StatusMethodNotAllowed)
        fmt.Fprintf(w, "/preferences only supports DELETE method\n")
    }

    if err := preferences.delete(); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, "Failed to delete preferences: %v\n", err)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "Preferences cleared");
}

func executeWithRetry(address string, port int, timeout time.Duration, operation func(utils.Instrument) error) error {
    inst, err := instCache.get(address, port, timeout, nil)
    if err != nil {
        return err
    }

    inst.SetTimeout(timeout)
    err = operation(inst)

    if err != nil && errors.Is(err, utils.ErrConnectionClosed) {
        log.Printf("Connection closed, attempting reconnect")
        instCache.invalidate(address, port)
        inst, err = instCache.get(address, port, timeout, nil)
        if err != nil {
            return fmt.Errorf("connection lost and reconnection failed: %w", err)
        }
        inst.SetTimeout(timeout)
        return operation(inst)
    }

    return err
}

func handleScpiRequest(w http.ResponseWriter, r *http.Request) {
    log.Println("Handling /scpi")
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
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

    address := body.Address
    if body.Simulated {
        address = "simulated"
    }

    timeout := time.Duration(body.TimeoutSeconds) * time.Second

    scpiResponse := scpiResponse{}
    if body.Type == "smart" {
        if (strings.Contains(body.Scpi, "?")) {
            var queryResponse string
            err := executeWithRetry(address, body.Port, timeout, func(inst utils.Instrument) error {
                var err error
                queryResponse, err = inst.Query(body.Scpi)
                return err
            })
            if err != nil {
                log.Printf("Error sending query: %v", err)
                scpiResponse.ServerError = fmt.Sprintf("Error sending query: %v", err)
            } else {
                scpiResponse.Response = queryResponse
            }
        } else {
            err := executeWithRetry(address, body.Port, timeout, func(inst utils.Instrument) error {
                return inst.Command(body.Scpi)
            })
            if err != nil {
                log.Printf("Error sending command: %v", err)
                scpiResponse.ServerError = fmt.Sprintf("Error sending command: %v", err)
            }
        }

        if body.AutoSystError && scpiResponse.ServerError == "" {
            var systErrors []string
            err := executeWithRetry(address, body.Port, timeout, func(inst utils.Instrument) error {
                var err error
                systErrors, err = inst.QueryError([]string{})
                return err
            })
            if err != nil {
                log.Printf("Error doing auto :syst:err?: %v", err)
                scpiResponse.ServerError = fmt.Sprintf("Error querying system errors: %v", err)
            } else {
                scpiResponse.Errors = systErrors
            }
        }

        w.WriteHeader(http.StatusOK)
        responseData , _ := json.Marshal(scpiResponse)
        fmt.Fprintf(w, "%s\n", responseData)

        return
    }

    if body.Type == "sendOnly" {
        err := executeWithRetry(address, body.Port, timeout, func(inst utils.Instrument) error {
            return inst.Command(body.Scpi)
        })
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

    if body.Port == 0 {
        body.Port = config.DefaultScpiSocketPort
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
