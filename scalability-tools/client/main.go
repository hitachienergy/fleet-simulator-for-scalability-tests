package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"hitachienergy/scalability-test-client/config"
	"hitachienergy/scalability-test-client/httpserver"
	"hitachienergy/scalability-test-client/simulation"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

const ANALYSIS_FILENAME = "simulator_analysis.txt"
const DEFAULT_OUTPUT_FOLDER = "device-simulator"

func main() {

	/* ------ parse input parameters ------ */

	var configData string
	flag.StringVar(&configData, "config", "", "config path (required)")
	var templatePath string
	flag.StringVar(&templatePath, "template", "", "overwritten path to client template (optional)")
	var opath string
	flag.StringVar(&opath, "opath", "", "overwritten output path (optional)")
	var clientNum int
	flag.IntVar(&clientNum, "num", 0, "overwritten number of clients (optional)")
	var idxOffset int
	flag.IntVar(&idxOffset, "offset", 0, "offset the device idx should start from (optional)")
	var influence string
	flag.StringVar(&influence, "influence", "{}", "simulation influnce range (optional)")
	var serverPort int
	flag.IntVar(&serverPort, "serverport", 8086, "overwritten port of the simulator's status server (optional)")
	flag.Parse()

	/* ------ setup simulation variables ------ */

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-5s|", i))
	}
	log := zerolog.New(output).With().Timestamp().Logger()
	mainlog := log.With().Str("object", "main").Logger()

	data := []byte(configData)
	simulationConfig, err := config.ParseConfig(data)
	if err != nil {
		mainlog.Fatal().Msgf("Fail to read configuration: %s", err)
	}

	switch simulationConfig.LogLevel {
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if len(templatePath) > 0 {
		mainlog.Info().Msgf("Overwrite template path. Original: %s, Now: %s", simulationConfig.Client.Template, templatePath)
		simulationConfig.Client.Template = templatePath
	}

	if clientNum > 0 {
		mainlog.Info().Msgf("Overwrite number of clients. Original: %d, Now: %d", simulationConfig.Client.Number, clientNum)
		simulationConfig.Client.Number = clientNum
	}

	if len(opath) > 0 {
		mainlog.Info().Msgf("Overwrite output path to: %s", opath)
		simulationConfig.Output.Path = opath
	} else {
		simulationConfig.Output.Path = filepath.Join(simulationConfig.Output.Path, DEFAULT_OUTPUT_FOLDER)
	}

	var influnceRange config.SimulationInfluenceCount
	err = json.Unmarshal([]byte(influence), &influnceRange)
	if err != nil {
		mainlog.Fatal().Err(err)
	}
	mainlog.Info().Msgf("Set Simulation Influence Constraints: {dummyWork: %d, crash: %d}", influnceRange.DummyWork, influnceRange.Crash)

	/* ------ start simulation ------ */

	mainlog.Info().Msg("Start Simulation...")

	simulator := simulation.NewSimulator(simulationConfig, data, log)
	connectChan := make(chan struct{}, 1)
	stopChann := make(chan struct{}, 1)

	// start HTTP server
	mainlog.Info().Msgf("Starting Status Server on port: %d", serverPort)
	server := httpserver.MainHttp(uint(serverPort), &mainlog, simulator, connectChan, stopChann)

	// create all the elements for devices simulation
	finishChann := make(chan struct{}, 1)
	err = simulator.SetupDevices(idxOffset, influnceRange, &log, finishChann)
	if err != nil {
		mainlog.Error().Msgf("Fail to initialize Simulator: %s", err)
		os.Exit(1)
	}
	mainlog.Info().Msg("Simulator succesfully initialized .")

	<-connectChan // wait for connection start event sent by the FIST Simulator Manger via HTTP endpoint

	err = simulator.StartDevices()
	if err != nil {
		mainlog.Error().Msgf("Fail to start and connect devices: %s", err)
		err = simulator.StopDevices()
		if err != nil {
			mainlog.Error().Msgf("Fail to stop devices: %s", err)
		}
		os.Exit(1)
	}

	mainlog.Info().Msg("All devices registered successfully.")

	/* ------ wait simulation to end ------ */

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stopSignal:
		mainlog.Info().Msg("System signal detected.")
	case <-finishChann:
		mainlog.Info().Msg("Simulation is finished. Saving results to disk...")
		err := simulator.SaveResult(filepath.Join(simulationConfig.Output.Path, ANALYSIS_FILENAME))
		if err != nil {
			mainlog.Error().Msgf("Fail to save analysis to disk: %s", err)
		}
	case <-stopChann:
		mainlog.Info().Msg("Simulation stop event detected.")
	}

	mainlog.Info().Msg("Stopping simulation...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	errServer := server.Shutdown(ctxShutDown)
	if errServer != nil {
		mainlog.Error().Msgf("Fail to stop http server: %s", errServer)
	}

	err = simulator.StopDevices()
	if err != nil {
		mainlog.Error().Msgf("Fail to stop devices: %s", err)
	}
	if errServer != nil || err != nil {
		os.Exit(1)
	}

	<-time.After(time.Second)

	mainlog.Info().Msg("Bye bye.")
}
