package main

import (
	"hitachienergy/scalability-test-client/examples/hawkbit/hawkbit"
	"hitachienergy/scalability-test-client/examples/httppool"
	"hitachienergy/scalability-test-client/templates"

	"golang.org/x/xerrors"
)

// ------------------------------------ Client Factory -----------------------------------
var HawkbitDMFDefaultFactory DMFDefaultClientFactory

type DMFDefaultClientFactory struct {
	templates.DeviceFactory
	Config HawkbitConfig
}

func (d DMFDefaultClientFactory) ParseConfig(data []byte) (templates.DeviceFactory, error) {
	configs, err := ParseConfig(data)
	if err != nil {
		return nil, xerrors.Errorf("Invalid config structure for devices simulation")
	}
	d.Config = *configs

	params := d.Config.Client.Args
	err = dmfCheckAndSetInputs(params)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d DMFDefaultClientFactory) NewDevice(controller templates.Controller) (client templates.Device, err error) {
	params := d.Config.Client.Args
	baseEndpoint := d.Config.Server.DevicesEndpoint

	httpPoolSize := params["httpPoolSize"].(int)
	if httpPoolSize > 0 {
		httppool.Pool.Init(httpPoolSize)
	}

	client = hawkbit.NewDMFClient(controller,
		params["tenant"].(string),
		baseEndpoint,
		params["virtualHost"].(string),
		params["replyExchange"].(string),
		httpPoolSize > 0,
		params["pingPeriod"].(uint),
		map[string]string{
			"targetType": "DMF",
		},
	)
	// c.UpdateManagerDMF = &defaultDMFUM{DMFUpdateManager: hawkbit.NewDMFUpdateManager(c.DMFAmqpService), ctr: controller}

	return client, nil
}

func dmfCheckAndSetInputs(params map[string]interface{}) (err error) {
	if _, ok := params["tenant"]; !ok {
		return xerrors.Errorf("Missing mandatory input (tenant)")
	}

	if _, ok := params["virtualHost"]; !ok {
		params["virtualHost"] = "/" // default virtualHost: "/"
	}

	if _, ok := params["replyExchange"]; !ok {
		params["replyExchange"] = "simulator.replyTo" // default rabbit exchange for sending messages: "simulator.replyTo"
	}

	if _, ok := params["pingPeriod"]; !ok {
		params["pingPeriod"] = uint(0) // default ping period: only ping once
	}
	if _, ok := params["pingPeriod"].(uint); !ok {
		return xerrors.Errorf("Invalid input (pingPeriod). Expected: uint")
	}
	if _, ok := params["httpPoolSize"].(int); !ok {
		params["httpPoolSize"] = 0 // default: not use HTTPPool (no limit on HTTP Connections)
	}

	return nil
}
