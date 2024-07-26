package device

import (
	"context"
	"fmt"
	"hitachienergy/scalability-test-client/config"
	"hitachienergy/scalability-test-client/templates"
	"math/rand"
	"sync"
	"time"

	"github.com/panjf2000/ants"
	"github.com/rs/zerolog"
)

/* MEMO

ants tasks cannot be externally stopped during execution.
This means that, if a crash happens, the scheduler will complete the current task anyway and then stop.
Using sync.Once ensure that device lifecycle behaviors are not repeated multiple times but only once.

*/

type ConnectCallback func(success bool)
type FinishCallback func(id string, start time.Time, duration time.Duration, success bool)

// DeviceController is a unit hold by a device.
// It contains all simulation information related to the device, and help to report device's status to the simulator
type DeviceController struct {
	templates.Controller

	Device templates.Device

	logger *zerolog.Logger

	cancel    context.CancelFunc
	deviceCtx context.Context

	id              string
	mainTask        string
	connectCallback ConnectCallback
	finishCallback  FinishCallback

	willCrash         bool
	crashWithin       time.Duration
	dummyTaskDuration time.Duration
	dummyTaskTimeout  time.Duration
	scheduler         *ants.Pool

	taskStartTime time.Time

	crashOnce        sync.Once
	connectOnce      sync.Once
	startTaskOnce    sync.Once
	completeTaskOnce sync.Once
}

// NewDeviceController creates an instance of the controller
func NewDeviceController(logger *zerolog.Logger, id string, mainTask string, connectFunc ConnectCallback, finishFunc FinishCallback) *DeviceController {
	localLogger := logger.With().Str("object", fmt.Sprintf("device-%s", id)).Logger()

	// scheduler for device tasks execution (real and simulated tasks)
	s, _ := ants.NewPool(1)

	return &DeviceController{
		logger:          &localLogger,
		connectCallback: connectFunc,
		finishCallback:  finishFunc,
		id:              id,
		mainTask:        mainTask,
		scheduler:       s,
	}
}

// NewDevice creates a new Device instance. This instance is internally used and managed by the DeviceController
func (c *DeviceController) NewDevice(sharedCtx context.Context, clientFactory templates.DeviceFactory) error {
	c.deviceCtx, c.cancel = context.WithCancel(sharedCtx)
	device, err := clientFactory.NewDevice(c)
	if err != nil {
		return err
	}
	c.Device = device
	return nil
}

// StartDevice tells the DeviceController to start the device
func (c *DeviceController) StartDevice() error {
	return c.Device.Start(c.deviceCtx)
}

// StartDevice tells DeviceController to stop the internal Device instance. If the task has already been completed, this function should have no effect.
func (c *DeviceController) StopDevice() error {
	if c.Device != nil {
		return c.Device.Stop()
	}
	return nil
}

// GetIdentifier returns the unique Identifier of this controller / device
func (c *DeviceController) GetIdentifier() string {
	return c.id
}

// GetLogger returns a device-specific logger, which generates formatted logs with the device identifier
func (c *DeviceController) GetLogger() *zerolog.Logger {
	return c.logger
}

// GetIdentifier returns the unique Identifier of this controller / device
func (c *DeviceController) GetScheduler() *ants.Pool {
	return c.scheduler
}

// Connect reports to the simulator that the device is connected successfully to the remote platform
func (c *DeviceController) Connect(success bool) {
	c.connectOnce.Do(func() {
		if success {
			c.logger.Debug().Msgf("%s connects successfully to the server", c.id)
		} else {
			c.logger.Debug().Msgf("%s failed connection to the server", c.id)
		}
		c.connectCallback(success)
	})
}

// StartTask logs that the target task, i.e. OTA Update, is started
func (c *DeviceController) StartTask() {
	c.startTaskOnce.Do(func() {

		c.taskStartTime = time.Now()

		if c.dummyTaskDuration > 0 {
			c.logger.Debug().Msg("Setup dummy work task...")
			c.dummyWork() // ensures that the scheduler will always start with a dummy work
			go func() {
				for {
					time.Sleep(c.dummyTaskTimeout)
					c.dummyWork()
				}
			}()
		}

		if c.willCrash {
			c.logger.Debug().Msg("Setup crash task...")
			// crash task is not part of the scheduler
			go func() {
				select {
				case <-time.After(c.crashWithin):
					c.logger.Debug().Msg("Crash event occurred")
					c.crashOnce.Do(func() {
						c.cancel()
						c.CompleteTask(false)
					})
				case <-c.deviceCtx.Done():
					c.logger.Debug().Msg("Disabled crash event")
				}
			}()
		}
	})
}

func (c *DeviceController) dummyWork() {
	c.scheduler.Submit(func() {
		c.logger.Debug().Msg("Dummy work in progress...")
		select {
		case <-time.After(c.dummyTaskDuration):
			c.logger.Debug().Msg("Dummy work completed")
		case <-c.deviceCtx.Done():
			c.logger.Debug().Msg("Dummy work interrupted")
		}
	})
}

// Complete Task logs the target task is completed and reports the details to the simulator
func (c *DeviceController) CompleteTask(success bool) {
	c.completeTaskOnce.Do(func() {
		c.scheduler.Release()

		duration := time.Since(c.taskStartTime)
		result := "fail"
		if success {
			result = "success"
		}
		c.logger.Debug().Msgf("%s completes (%s).", c.mainTask, result)

		c.finishCallback(c.id, c.taskStartTime, duration, success)
	})
}

// configureDeviceCtx is a private function. It will make the shared context per-device base so that the crash of a single device does not affect others
func (c *DeviceController) configureDeviceCtx(sharedCtx context.Context) context.Context {
	c.deviceCtx, c.cancel = context.WithCancel(sharedCtx)
	return c.deviceCtx
}

// CalculateAndSetController takes the simulation configuration and generates the random dummywork and crash for each device.
// It returns a list of pre-configured controllers. Each controller should be assigned to a distinct device
func CalculateAndSetController(config config.Config, offset int, influnceRange config.SimulationInfluenceCount, logger *zerolog.Logger, connectDevice ConnectCallback, finishDevice FinishCallback) (controllers []*DeviceController, err error) {
	var r *rand.Rand
	if config.Simulation.Seed != 0 {
		r = rand.New(rand.NewSource(config.Simulation.Seed))
	} else {
		r = rand.New(rand.NewSource(time.Now().Unix()))
	}

	for i := offset; i < offset+config.Client.Number; i++ {
		controllers = append(controllers, NewDeviceController(logger, fmt.Sprintf("%s%d", config.Client.NamePrefix, i), config.Simulation.Task, connectDevice, finishDevice))
	}

	if influnceRange.DummyWork > 0 {
		dummyWorkDetail := config.Simulation.DummyWork
		mask := SelectRandom(r, config.Client.Number, influnceRange.DummyWork)
		dummyTaskDuration := GenerateVariation(r, time.Duration(dummyWorkDetail.Duration), time.Duration(dummyWorkDetail.Variation), len(mask))
		for i, idx := range mask {
			logger.Info().Msgf("Device %d will run a dummy task. Duration: %d Period: %d ", idx, time.Duration(dummyTaskDuration[i]), time.Duration(dummyWorkDetail.Period))
			controllers[idx].dummyTaskDuration = time.Duration(dummyTaskDuration[i])
			controllers[idx].dummyTaskTimeout = time.Duration(dummyWorkDetail.Period)
		}
	}

	crashDetail := config.Simulation.Crash
	if influnceRange.Crash > 0 {
		mask := SelectRandom(r, config.Client.Number, influnceRange.Crash)
		dummyCrashDelay := GenerateDuration(r, time.Duration(crashDetail.Within), len(mask))
		for i, idx := range mask {
			logger.Info().Msgf("Device %d will crash. Within: %d ", idx, time.Duration(dummyCrashDelay[i]))
			controllers[idx].willCrash = true
			controllers[idx].crashWithin = time.Duration(dummyCrashDelay[i])
		}
	}

	return controllers, nil
}

// SelectRandom randomly selects k devices from n devices
func SelectRandom(r *rand.Rand, n int, k int) []int {
	arr := makeRange(0, n-1)
	r.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
	return arr[:k]
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

// GenerateVariation randomly generates k number with a fixed average. All k numbers fall into the predefined variation range
func GenerateVariation(r *rand.Rand, target time.Duration, variation time.Duration, k int) []float64 {
	if k == 0 {
		return nil
	}

	sum := float64(k) * float64(target)
	rangeSize := float64(2 * variation)
	offset := float64(target - variation)
	numbers := make([]float64, k)
	for i := 0; i < k-1; i++ {
		numbers[i] = r.Float64()*float64(rangeSize) + offset
		sum -= numbers[i]
	}
	numbers[k-1] = sum
	return numbers
}

// GenerateVariation randomly generates k number with a fixed average. All k numbers fall into the predefined variation range
func GenerateDuration(r *rand.Rand, within time.Duration, k int) []float64 {
	if k == 0 {
		return nil
	}

	numbers := make([]float64, k)
	for i := 0; i < k; i++ {
		numbers[i] = r.Float64() * float64(within)
	}
	return numbers
}
