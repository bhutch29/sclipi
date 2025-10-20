package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bhutch29/sclipi/internal/utils"
)

var version = "undefined"
var instCache = newInstrumentCache()
var config *Config
var preferences *Preferences

type scpiResponse struct {
	Response    string   `json:"response"`
	Errors      []string `json:"errors"`
	ServerError string   `json:"serverError"`
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
	http.HandleFunc("/commands", handleCommandsRequest)
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
		log.Printf("Connection closed, attempting reconnect")
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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "/commands only supports GET")
		return
	}

	address := r.URL.Query().Get("address")
	portString := r.URL.Query().Get("port")

	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing required parameter: address")
		return
	}

	if portString == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing required parameter: port")
		return
	}

  port, err := strconv.Atoi(portString);
  if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Required parameter port must be a number")
		return
  }

	log.Printf("Handling /commands request for address=%s, port=%d", address, port)

	inst, err := instCache.get(address, port, 10 * time.Second, nil)
	if err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Failed to get instrument: %v", err)
    return
	}

  starTree, colonTree, err := inst.GetSupportedCommandsTree()
  if err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
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
		log.Printf("Error writing gzipped response: %v", err)
	}
}

func handleScpiRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /scpi")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "/scpi only supports POST")
		return
	}

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	scpi := string(bodyData)
	if scpi == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Request body (SCPI command) cannot be empty")
		return
	}

	address := r.URL.Query().Get("address")
	portString := r.URL.Query().Get("port")
	simulatedString := r.URL.Query().Get("simulated")
	autoSystErrorString := r.URL.Query().Get("autoSystErr")
	timeoutSecondsString := r.URL.Query().Get("timeoutSeconds")

	if address == "" {
		address = preferences.ScpiAddress
	}

	port := preferences.ScpiPort
	if portString != "" {
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Parameter port must be a number")
			return
		}
		if port < 1 || port > 65535 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Port must be between 1 and 65535")
			return
		}
	}

	simulated := false
	if simulatedString != "" {
		simulated = simulatedString == "true"
	}

	autoSystError := false
	if autoSystErrorString != "" {
		autoSystError = autoSystErrorString == "true"
	}

	timeoutSeconds := 10
	if timeoutSecondsString != "" {
		var err error
		timeoutSeconds, err = strconv.Atoi(timeoutSecondsString)
		if err != nil || timeoutSeconds < 0 {
			w.WriteHeader(http.StatusBadRequest)
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
	if strings.Contains(scpi, "?") {
		var queryResponse string
		executeError = executeWithRetry(address, port, timeout, func(inst utils.Instrument) error {
			var err error
			queryResponse, err = inst.Query(scpi)
			return err
		})
		if executeError != nil {
			log.Printf("Error sending query: %v", executeError)
			scpiResponse.ServerError = fmt.Sprintf("Error sending query: %v", executeError)
		} else {
			scpiResponse.Response = queryResponse
		}
	} else {
		executeError = executeWithRetry(address, port, timeout, func(inst utils.Instrument) error {
			return inst.Command(scpi)
		})
		if executeError != nil {
			log.Printf("Error sending command: %v", executeError)
			scpiResponse.ServerError = fmt.Sprintf("Error sending command: %v", executeError)
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
			log.Printf("Error doing auto :syst:err?: %v", err)
			scpiResponse.ServerError = fmt.Sprintf("Error querying system errors: %v", err)
		} else {
			scpiResponse.Errors = systErrors
		}
	}

	w.WriteHeader(http.StatusOK)
	responseData, _ := json.Marshal(scpiResponse)
	fmt.Fprintf(w, "%s\n", responseData)
}
