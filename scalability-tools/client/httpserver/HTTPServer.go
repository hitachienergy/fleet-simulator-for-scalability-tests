package httpserver

import (
	"encoding/json"
	"fmt"
	"hitachienergy/scalability-test-client/simulation"
	"io"

	"net/http"

	"github.com/rs/zerolog"
)

// MainHttp starts the HTTP Server
func MainHttp(port uint, logger *zerolog.Logger, simulator *simulation.Simulator, connectChan chan struct{}, stopChan chan struct{}) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "Device Simulator\n")
		},
	))
	mux.Handle("/ready", getReadyStateHandler(simulator))
	mux.Handle("/start", startDevicesHandler(connectChan))
	mux.Handle("/connected", getConnectedStateHandler(simulator))
	mux.Handle("/stats", getSimulationProcessHandler(simulator))
	mux.Handle("/stop", stopHandler(stopChan))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Msgf("Fail to start HTTP server: %s", err)
		}
	}()

	return server
}

// getSimulationProcessHandler is a url handler that returns the in-time statistics of the simulation process
func getSimulationProcessHandler(simulator *simulation.Simulator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := simulator.GetProcess()
		json, _ := json.Marshal(stats)
		w.Write(json)
	}
}

// getReadyStateHandler is a url handler that will only return '200 OK' after all devices are reigstered to the server
func getReadyStateHandler(simulator *simulation.Simulator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ready := simulator.IsReady()
		if !ready {
			w.WriteHeader(503)
		}
	}
}

// getConnectedStateHandler is a url handler that will only return '200 OK' after all devices are reigstered to the server
func getConnectedStateHandler(simulator *simulation.Simulator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connected := simulator.IsConnected()
		if !connected {
			w.WriteHeader(503)
		}
	}
}

// stopHandler is a url handler that stops the simulation
func stopHandler(stopChan chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stopChan <- struct{}{}
		io.WriteString(w, "Stopping Simulation\n")
	}
}

// stopHandler is a url handler that stops the simulation
func startDevicesHandler(connectChan chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectChan <- struct{}{}
		io.WriteString(w, "Connecting Devices\n")
	}
}
