package main

import (
	"hitachienergy/scalability-test-client/examples/hawkbit/hawkbit"
	"hitachienergy/scalability-test-client/examples/httppool"
	"hitachienergy/scalability-test-client/templates"

	"golang.org/x/xerrors"
)

// ------------------------------------ Client Factory -----------------------------------

var HawkbitDDIDefaultFactory DDIDefaultClientFactory

type DDIDefaultClientFactory struct {
	templates.DeviceFactory
	Config HawkbitConfig
}

func (d DDIDefaultClientFactory) ParseConfig(data []byte) (templates.DeviceFactory, error) {
	configs, err := ParseConfig(data)
	if err != nil {
		return nil, xerrors.Errorf("Invalid config structure for devices simulation")
	}
	d.Config = *configs

	err = ddiCheckAndSetInputs(d.Config.Client.Args)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d DDIDefaultClientFactory) NewDevice(controller templates.Controller) (client templates.Device, err error) {
	params := d.Config.Client.Args
	baseEndpoint := d.Config.Server.DevicesEndpoint

	httpPoolSize := params["httpPoolSize"].(int)
	if httpPoolSize > 0 {
		httppool.Pool.Init(httpPoolSize)
	}

	client = hawkbit.NewDDIClient(controller,
		params["tenant"].(string),
		params["pollDelay"].(int),
		baseEndpoint,
		params["gatewayToken"].(string),
		httpPoolSize > 0,
	)

	return client, nil
}

func ddiCheckAndSetInputs(params map[string]interface{}) (err error) {
	if _, ok := params["tenant"]; !ok {
		return xerrors.Errorf("Missing mandatory input (tenant)")
	}

	if _, ok := params["pollDelay"]; !ok {
		params["pollDelay"] = 30 // default poll delay: 30s
	}
	if _, ok := params["pollDelay"].(int); !ok {
		return xerrors.Errorf("Invalid input (pollDelay). Expected: int")
	}
	if _, ok := params["httpPoolSize"].(int); !ok {
		params["httpPoolSize"] = 0 // default: not use HTTPPool (no limit on HTTP Connections)
	}

	return nil
}
