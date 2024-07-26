package main

import (
	"hitachienergy/scalability-test-client/examples/httppool"
	"hitachienergy/scalability-test-client/examples/thingsboard/thingsboard"
	"hitachienergy/scalability-test-client/templates"

	"golang.org/x/xerrors"
)

// ------------------------------------ Client Factory -----------------------------------

var TBHTTPDefaultFactory HTTPDefaultClientFactory

type HTTPDefaultClientFactory struct {
	templates.DeviceFactory
	Config ThingsBoardConfig
}

func (d HTTPDefaultClientFactory) ParseConfig(data []byte) (templates.DeviceFactory, error) {
	configs, err := ParseConfig(data)
	if err != nil {
		return nil, xerrors.Errorf("Invalid config structure for devices simulation")
	}
	d.Config = *configs

	err = httpCheckAndSetInputs(d.Config.Client.Args)
	if err != nil {
		return nil, err
	}

	return d, nil
}
func (d HTTPDefaultClientFactory) NewDevice(controller templates.Controller) (client templates.Device, err error) {
	params := d.Config.Client.Args
	address := d.Config.Server.DevicesEndpoint

	httpPoolSize := params["httpPoolSize"].(int)
	if httpPoolSize > 0 {
		httppool.Pool.Init(httpPoolSize)
	}

	c := thingsboard.NewHTTPClient(controller, address, params["pollDelay"].(int), httpPoolSize > 0)
	// c.UpdateModule = &defaultHTTPUM{UpdateManager: thingsboard.NewUpdateManager(c.HTTPService), ctr: controller}

	return c, nil
}

func httpCheckAndSetInputs(params map[string]interface{}) (err error) {
	if _, ok := params["pollDelay"]; !ok {
		params["pollDelay"] = 30 // default poll delay: 30s
	} else if _, ok := params["pollDelay"].(int); !ok {
		return xerrors.Errorf("Invalid input (pollDelay). Expected: int")
	}
	if _, ok := params["httpPoolSize"].(int); !ok {
		params["httpPoolSize"] = 0 // default: not use HTTPPool (no limit on HTTP Connections)
	}

	return nil
}
