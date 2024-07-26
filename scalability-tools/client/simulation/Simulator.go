package simulation

import (
	"context"
	"hitachienergy/scalability-test-client/config"
	"hitachienergy/scalability-test-client/device"
	"hitachienergy/scalability-test-client/templates"
	"plugin"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

// Simulator is the central manager of the device-side simulation
type Simulator struct {
	config    config.Config
	rawConfig []byte

	controllers   []*device.DeviceController
	clientFactory templates.DeviceFactory

	cancel context.CancelFunc

	waitgroup   *waitGroup
	taskStats   *dataStore
	finishChann chan struct{}

	isReady     atomic.Bool
	isConnected atomic.Bool

	log zerolog.Logger
}

// NewSimulator creates a new instance of simulator
func NewSimulator(simulationConfig config.Config, rawConfig []byte, log zerolog.Logger) *Simulator {
	simLog := log.With().Str("object", "simulator manager").Logger()
	return &Simulator{
		config:    simulationConfig,
		rawConfig: rawConfig,
		log:       simLog,
	}
}

// SetupDevices will create and starts all devices according to the simulation configuration
func (s *Simulator) SetupDevices(indexOffset int, influnceRange config.SimulationInfluenceCount, logger *zerolog.Logger, finishChann chan struct{}) (err error) {
	if s.config.Client.Number <= 0 {
		return xerrors.Errorf("Non-positive client number. Got %d", s.config.Client.Number)
	}

	s.finishChann = finishChann
	s.taskStats = newDataStore(s.config.Client.Number)
	s.waitgroup = newWaitGroup(s.config.Client.Number)

	// load platform specific devices creation scripts
	clientFactory, err := loadClientFactory(s.config.Client.Template, s.config.Client.Factory, s.rawConfig, logger)
	if err != nil {
		return err
	}
	s.clientFactory = clientFactory

	// precompute devices controllers
	controllers, err := device.CalculateAndSetController(s.config, indexOffset, influnceRange, logger, s.waitgroup.add, s.finishDevice)
	if err != nil {
		return err
	}
	s.controllers = controllers

	s.isReady.Store(true)

	return nil
}

func (s *Simulator) StartDevices() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	failureCount := 0
	switch s.config.Client.DevicesRegisterMode {
	case "parallel":
		failureCount = s.parallelRegister(ctx, s.clientFactory)
	case "sequential":
		failureCount = s.sequentialRegister(ctx, s.clientFactory)
	default:
		return xerrors.Errorf("Unrecognized devices registration mode %s", s.config.Client.DevicesRegisterMode)
	}
	if failureCount > 0 {
		return xerrors.Errorf("Fail to create and connect all devices. Fail: %d, Total: %d", failureCount, s.config.Client.Number)
	}
	s.isConnected.Store(true)
	return nil
}

// StopDevices stops and clean all the devices
func (s *Simulator) StopDevices() (err error) {
	if s.cancel != nil {
		s.cancel()
	}

	for _, controller := range s.controllers {
		if controller == nil {
			continue
		}
		err = controller.StopDevice()
	}

	<-time.After(time.Second)

	s.cancel = nil

	return err
}

// IsReady shows if all devices are registered to the devices
func (s *Simulator) IsReady() bool {
	return s.isReady.Load()
}

// IsReady shows if all devices are registered to the devices
func (s *Simulator) IsConnected() bool {
	return s.isConnected.Load()
}

// GetProcess returns the in-time statistics of the simulation
func (s *Simulator) GetProcess() SimulationStats {
	return s.taskStats.getStatistics()
}

// SaveResult saves the simulation results to the target path
func (s *Simulator) SaveResult(opth string) error {
	return s.taskStats.saveToDisk(opth)
}

// finishDevice respresents the logic that need to be done when each device finishes it simulation
// It is passed to the controller to be triggered for each device
func (s *Simulator) finishDevice(id string, start time.Time, duration time.Duration, success bool) {
	finish := s.taskStats.storeState(id, start, duration, success)
	if finish && s.finishChann != nil {
		s.finishChann <- struct{}{}
	}
}

// sequentialRegister connects all the devices to the server sequentially
func (s *Simulator) sequentialRegister(ctx context.Context, clientFactory templates.DeviceFactory) (failureCount int) {
	s.log.Info().Msg("Starting devices registration in sequential mode ")
	for _, controller := range s.controllers {
		err := controller.NewDevice(ctx, clientFactory)
		if err != nil {
			log.Err(err).Send()
			failureCount += 1
			break
		}
		err = controller.StartDevice()
		if err != nil {
			log.Err(err).Send()
			failureCount += 1
			break
		}
	}
	return failureCount
}

// paralleRegister connects all the devices to the server in parallel
func (s *Simulator) parallelRegister(ctx context.Context, clientFactory templates.DeviceFactory) (failureCount int) {
	s.log.Info().Msg("Starting devices registration in parallel mode ")
	for idx, controller := range s.controllers {
		go func(idx int, controller *device.DeviceController) {
			err := controller.NewDevice(ctx, clientFactory)
			if err != nil {
				log.Err(err).Send()
				return
			}
			err = controller.StartDevice()
			if err != nil {
				log.Err(err).Send()
				return
			}
		}(idx, controller)
	}
	failureCount = <-s.waitgroup.readyChan
	return failureCount
}

func loadClientFactory(path string, factoryName string, data []byte, logger *zerolog.Logger) (factory templates.DeviceFactory, err error) {
	p, err := plugin.Open(path)
	if err != nil {
		return factory, err
	}

	object, err := p.Lookup(factoryName)
	if err != nil {
		return factory, err
	}

	factory, ok := object.(templates.DeviceFactory)
	if !ok {
		return factory, xerrors.Errorf("Invalid client factory interface %s: %s", factoryName, path)
	}

	factory, err = factory.ParseConfig(data)
	if err != nil {
		return nil, xerrors.Errorf("Invalid YAML config structure.")
	}

	return factory, nil
}
