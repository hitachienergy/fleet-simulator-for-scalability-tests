package templates

import (
	"context"

	"github.com/panjf2000/ants"
	"github.com/rs/zerolog"
)

// Controller records and generates the simulation scenario. Clients can interacti with it to do different simulation
type Controller interface {

	// interactions with the Simulator
	NewDevice(ctx context.Context, clientFactory DeviceFactory) error
	StartDevice() error
	StopDevice() error

	// interactions with the Device
	Connect(success bool)
	StartTask()
	CompleteTask(success bool)

	// getters and utils
	GetIdentifier() string
	GetLogger() *zerolog.Logger
	GetScheduler() *ants.Pool
}
