package templates

import (
	"context"
)

// Device is a abstract of the simulated device
type Device interface {
	Start(context.Context) error
	Stop() error
}

// DeviceFactory is a factory to create device instances
type DeviceFactory interface {
	NewDevice(ctr Controller) (Device, error)
	ParseConfig(data []byte) (DeviceFactory, error)
}
