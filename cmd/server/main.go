package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
  "log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bhutch29/sclipi/internal/utils"
)

var version = "unknown"
var instCache = newInstrumentCache()
var config *Config
var preferences *Preferences

type scpiResponse struct {
	Response    string   `json:"response"`
	Errors      []string `json:"errors"`
	ServerError string   `json:"serverError"`
}

type healthResponse struct {
  Healthy        bool   `json:"healthy"`
  ConnectionMode string `json:"connectionMode"`
  Version        string `json:"version"`
}

func main() {
  opts := &slog.HandlerOptions{
    Level: slog.LevelDebug,
  }
  logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
  slog.SetDefault(logger);

	var err error
	config, err = loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	preferences, err = loadPreferences()
	if err != nil {
		log.Fatalf("Failed to load preferences: %v", err)
	} else if preferences == nil {
	  var prefs Preferences
    preferences = &prefs;
		log.Printf("No preferences found, loaded default preferences: %+v", preferences)
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
	http.HandleFunc("/commands", handleCommandsRequest)
	http.HandleFunc("/preferences", handlePreferences)
	http.HandleFunc("/isConnected", handleIsConnected)
	http.HandleFunc("/dumpInstCache", handleDumpInstCache)

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
	slog.Debug("Handling request", "route", "/health")
  healthResponse := healthResponse{Healthy: true, Version: version, ConnectionMode: config.ConnectionMode}
	slog.Debug("Request info", "route", "/health", "response", healthResponse)
	w.WriteHeader(http.StatusOK)
	responseData, _ := json.Marshal(healthResponse)
	fmt.Fprintf(w, "%s\n", responseData)
}

func handleAddress(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/scpiAddress")
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", preferences.ScpiAddress)
	} else if r.Method == http.MethodPost {
		bodyData, err := io.ReadAll(r.Body)
		if err != nil {
	    slog.Error("Failed to read body from post", "route", "/scpiAddress", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		address := string(bodyData)
		preferences.ScpiAddress = address
		if err := preferences.save(); err != nil {
			slog.Warn("Failed to save preferences", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ScpiAddress updated to %v\n", address)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	  slog.Error("Received request with unsupported method", "route", "/scpiAddress", "method", r.Method)
		fmt.Fprintf(w, "/scpiAddress supports GET and POST methods\n")
	}
}

func handlePort(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/scpiPort")
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", preferences.ScpiPort)
	} else if r.Method == http.MethodPost {
		bodyData, err := io.ReadAll(r.Body)
		if err != nil {
	    slog.Error("Failed to read body from post", "route", "/scpiPort", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var port int
		_, err = fmt.Sscanf(string(bodyData), "%d", &port)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
      slog.Error("Invalid port number: must be an integer", "route", "/scpiPort", "error", err)
			fmt.Fprintf(w, "Invalid port number: must be an integer\n")
			return
		}

		if port < 1 || port > 65535 {
			w.WriteHeader(http.StatusBadRequest)
      slog.Error("Invalid port number: must be an between 1 and 65535", "route", "/scpiPort", "error", err, "port", port)
			fmt.Fprintf(w, "Invalid port number: must be between 1 and 65535\n")
			return
		}

		preferences.ScpiPort = port
		if err := preferences.save(); err != nil {
			slog.Warn("Failed to save preferences", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Port updated to %d\n", port)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	  slog.Error("Received request with unsupported method", "route", "/scpiPort", "method", r.Method)
		fmt.Fprintf(w, "/scpiPort supports GET and POST methods\n")
	}
}

func handlePreferences(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/preferences")
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
	  slog.Error("Received request with unsupported method", "route", "/preferences", "method", r.Method)
		fmt.Fprintf(w, "/preferences only supports DELETE method\n")
	}

	if err := preferences.delete(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to delete preferences: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Preferences cleared")
}

func executeWithRetry(address string, port int, timeout time.Duration, operation func(utils.Instrument) error) error {
	inst, err := instCache.get(address, port, timeout, nil)
	if err != nil {
		return err
	}

	inst.SetTimeout(timeout)
	err = operation(inst)

	if err != nil && errors.Is(err, utils.ErrConnectionClosed) {
		slog.Warn("Connection closed, attempting reconnect", "error", err)
		instCache.invalidate(address, port)
		inst, err = instCache.get(address, port, timeout, nil)
		if err != nil {
			return err
		}
		inst.SetTimeout(timeout)
		return operation(inst)
	}

	return err
}

func handleCommandsRequest(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/commands")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	  slog.Error("Received request with unsupported method", "route", "/commands", "method", r.Method)
		fmt.Fprintln(w, "/commands only supports GET")
		return
	}

	address := r.URL.Query().Get("address")
	portString := r.URL.Query().Get("port")

	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Missing required parameter: address", "route", "/commands")
		fmt.Fprintln(w, "Missing required parameter: address")
		return
	}

	if portString == "" {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Missing required parameter: port", "route", "/commands")
		fmt.Fprintln(w, "Missing required parameter: port")
		return
	}

  port, err := strconv.Atoi(portString);
  if err != nil {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Required parameter port must be a number", "route", "/commands")
		fmt.Fprintln(w, "Required parameter port must be a number")
		return
  }

	slog.Debug("Request info", "route", "/commands", "address", address, "port", port)

	inst, err := instCache.get(address, port, 10 * time.Second, nil)
	if err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
    slog.Error("Failed to get instrument", "route", "/commands", "error", err)
    fmt.Fprintf(w, "Failed to get instrument: %v", err)
    return
	}

  starTree, colonTree, err := inst.GetSupportedCommandsTree()
  if err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
    slog.Error("Failed to get commands", "route", "/commands", "error", err)
    fmt.Fprintf(w, "Failed to get commands: %v", err)
    return
  }

  type result struct {
    StarTree utils.ScpiNode `json:"starTree"`
    ColonTree utils.ScpiNode `json:"colonTree"`
  }

  responseData, err := json.Marshal(result{StarTree: starTree, ColonTree: colonTree})
  if err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
    slog.Error("Failed to read commands", "route", "/commands", "error", err)
    fmt.Fprintf(w, "Failed to read commands: %v", err)
    return
  }

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()

	_, err = gzipWriter.Write(responseData)
	if err != nil {
    slog.Error("Error writing gzipped response", "route", "/commands", "error", err)
	}
}

func handleScpiRequest(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/scpi")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	  slog.Error("Received request with unsupported method", "route", "/scpi", "method", r.Method)
		fmt.Fprintln(w, "/scpi only supports POST")
		return
	}

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Failed to read request body", "route", "/scpi", "error", err)
		fmt.Fprintln(w, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	scpi := string(bodyData)
	if scpi == "" {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Request body (SCPI command) cannot be empty", "route", "/scpi")
		fmt.Fprintln(w, "Request body (SCPI command) cannot be empty")
		return
	}

	address := r.URL.Query().Get("address")
	portString := r.URL.Query().Get("port")
	simulatedString := r.URL.Query().Get("simulated")
	autoSystErrorString := r.URL.Query().Get("autoSystErr")
	timeoutSecondsString := r.URL.Query().Get("timeoutSeconds")

	slog.Debug("Request info", "route", "/scpi", "scpi", scpi, "address", address, "port", portString, "simulated", simulatedString, "autoSystErr", autoSystErrorString, "timeoutSeconds", timeoutSecondsString)

	if address == "" {
		address = preferences.ScpiAddress
	}

	port := preferences.ScpiPort
	if portString != "" {
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
      slog.Error("Parameter port must be a number", "route", "/scpi", "port", port)
			fmt.Fprintln(w, "Parameter port must be a number")
			return
		}
		if port < 1 || port > 65535 {
			w.WriteHeader(http.StatusBadRequest)
      slog.Error("Port must be between 1 and 65535", "route", "/scpi", "port", port)
			fmt.Fprintln(w, "Port must be between 1 and 65535")
			return
		}
	}

  simulated := simulatedString == "true"
	autoSystError := autoSystErrorString == "true"

	timeoutSeconds := 10
	if timeoutSecondsString != "" {
		var err error
		timeoutSeconds, err = strconv.Atoi(timeoutSecondsString)
		if err != nil || timeoutSeconds < 0 {
			w.WriteHeader(http.StatusBadRequest)
      slog.Error("Parameter timeoutSeconds must be a positive number", "route", "/scpi", "timeoutSeconds", timeoutSeconds)
			fmt.Fprintln(w, "Parameter timeoutSeconds must be a positive number")
			return
		}
	}

	if simulated {
		address = "simulated"
	}

	timeout := time.Duration(timeoutSeconds) * time.Second

  var executeError error
	scpiResponse := scpiResponse{}

  if (scpi == ":_ERR") {
    scpiResponse.ServerError = fmt.Sprint("This is a fake server error for testing purposes.")
    var errors []string
    if (autoSystError) {
      errors = append(errors, "First fake :SYST:ERR? response")
      errors = append(errors, "Another fake :SYST:ERR? response for testing")
    }
    scpiResponse.Errors = errors;

	  w.WriteHeader(http.StatusOK)
	  responseData, _ := json.Marshal(scpiResponse)
	  fmt.Fprintf(w, "%s\n", responseData)
    return
  }

  if (scpi == ":_SLOW?") {
    time.Sleep(8 * time.Second)
    scpiResponse.Response = "Testing command :_SLOW? returned after 8 seconds, ignoring whatever timeout was set";

	  w.WriteHeader(http.StatusOK)
	  responseData, _ := json.Marshal(scpiResponse)
	  fmt.Fprintf(w, "%s\n", responseData)
    return
  }

	if strings.Contains(scpi, "?") {
		var queryResponse string
		executeError = executeWithRetry(address, port, timeout, func(inst utils.Instrument) error {
			var err error
			queryResponse, err = inst.Query(scpi)
			return err
		})
		if executeError != nil {
			slog.Error("Error sending query", "route", "/scpi", "error", executeError)
			scpiResponse.ServerError = fmt.Sprintf("%v", executeError)
		} else {
			scpiResponse.Response = queryResponse
		}
	} else {
		executeError = executeWithRetry(address, port, timeout, func(inst utils.Instrument) error {
			return inst.Command(scpi)
		})
		if executeError != nil {
			slog.Error("Error sending command", "route", "/scpi", "error", executeError)
			scpiResponse.ServerError = fmt.Sprintf("%v", executeError)
		}
	}

	if autoSystError && !errors.Is(executeError, utils.ErrConnectionClosed) {
		var systErrors []string
		err := executeWithRetry(address, port, timeout, func(inst utils.Instrument) error {
			var err error
			systErrors, err = inst.QueryError([]string{})
			return err
		})
		if err != nil {
      slog.Error("Error doing auto :syst:err?", "route", "/scpi", "error", err)
			scpiResponse.ServerError = fmt.Sprintf("Error querying system errors: %v", err)
		} else {
			scpiResponse.Errors = systErrors
		}
	}

	w.WriteHeader(http.StatusOK)
	responseData, _ := json.Marshal(scpiResponse)
	fmt.Fprintf(w, "%s\n", responseData)
}

func handleIsConnected(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/isConnected")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	  slog.Error("Received request with unsupported method", "route", "/isConnected", "method", r.Method)
		fmt.Fprintln(w, "/isConnected only supports GET")
		return
	}

	address := r.URL.Query().Get("address")
	portString := r.URL.Query().Get("port")
	timeoutSecondsString := r.URL.Query().Get("timeoutSeconds")
	simulatedString := r.URL.Query().Get("simulated")

	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Missing required parameter: address", "route", "/isConnected")
		fmt.Fprintln(w, "Missing required parameter: address")
		return
	}

	if portString == "" {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Missing required parameter: port", "route", "/isConnected")
		fmt.Fprintln(w, "Missing required parameter: port")
		return
	}

  port, err := strconv.Atoi(portString);
  if err != nil {
		w.WriteHeader(http.StatusBadRequest)
    slog.Error("Required parameter port must be a number", "route", "/isConnected")
		fmt.Fprintln(w, "Required parameter port must be a number")
		return
  }

	timeoutSeconds := 10
	if timeoutSecondsString != "" {
		var err error
		timeoutSeconds, err = strconv.Atoi(timeoutSecondsString)
		if err != nil || timeoutSeconds < 0 {
			w.WriteHeader(http.StatusBadRequest)
      slog.Error("Parameter timeoutSeconds must be a positive number", "route", "/isConnected", "timeoutSeconds", timeoutSeconds)
			fmt.Fprintln(w, "Parameter timeoutSeconds must be a positive number")
			return
		}
	}

  simulated := simulatedString == "true"
	if simulated {
		address = "simulated"
	}

	timeout := time.Duration(timeoutSeconds) * time.Second

  executeError := executeWithRetry(address, port, timeout, func(inst utils.Instrument) error {
    _, err := inst.Query("*IDN?")
		return err
	})

	w.WriteHeader(http.StatusOK)

	if executeError != nil {
		slog.Error("Error sending test *IDN?", "route", "/isConnected", "error", executeError)
	  fmt.Fprintln(w, "false")
	} else {
	  fmt.Fprintln(w, "true")
	}
}

func handleDumpInstCache(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "route", "/dumpInstCache")
  slog.Info("Dumping instCache", "cache", instCache.cache)
  fmt.Fprintf(w, "%v\n", instCache.cache)
	w.WriteHeader(http.StatusOK)
}
