package thingsboard

import (
	"context"
	"hitachienergy/scalability-test-client/templates"
	"time"
)

// func RunHTTP() {
// 	log.Info().Msg("Start Thingsboard HTTP Client...")
// 	token := "http0"
// 	if len(os.Args) > 1 {
// 		token = os.Args[1]
// 	}
// 	client := NewHTTPClient(device.NewSimulationController(&log.Logger, token, "test",
// 		func(success bool) {
// 			fmt.Println("Connected!")
// 		}, func(id string, start time.Time, duration time.Duration, success bool) {
// 			fmt.Println("Finished")
// 		}),
// 		"localhost:8080", 3, false)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	err := client.Start(ctx)
// 	if err != nil {
// 		log.Err(err).Send()
// 	}

// 	stopSignal := make(chan os.Signal, 1)
// 	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

// 	<-stopSignal
// 	cancel()
// 	log.Info().Msg("Stop Thingsboard HTTP Client...")
// }

type HTTPClient struct {
	*HTTPService
	controller templates.Controller
	UpdateModule

	pollDelay time.Duration
}

// NewHTTPClient creates an instance of HTTP Client
func NewHTTPClient(controller templates.Controller, endpoint string, pollDelay int, useHTTPPool bool) *HTTPClient {
	api := newHTTPService(endpoint, controller.GetIdentifier(), useHTTPPool)
	c := &HTTPClient{
		controller:  controller,
		HTTPService: api,
		pollDelay:   time.Duration(pollDelay),
	}
	c.UpdateModule = newUpdateManager(c)
	return c
}

// Start implements Device.Start
func (c *HTTPClient) Start(ctx context.Context) (err error) {
	err = c.ReportCurrState()
	c.controller.Connect(err == nil)
	if err != nil {
		return err
	}

	go func() {
	out:
		for {
			select {
			case <-ctx.Done():
				break out
			case <-time.After(c.pollDelay * time.Second):
				err := c.poll()
				if err != nil {
					c.controller.GetLogger().Err(err).Send()
				}
			}
		}
	}()

	return nil
}

// Stop implements device.Stop
func (c *HTTPClient) Stop() error {
	return nil
}

// poll retrieves messages periodically from the server
func (c *HTTPClient) poll() (err error) {
	fwInfo, err := c.getFirmwareInfo()
	if err != nil {
		return err
	}
	if !c.CheckFw(*fwInfo) {
		return nil
	}

	c.controller.StartTask()
	c.controller.GetScheduler().Submit(func() {
		err = c.StartUpdate(*fwInfo)
		if err != nil {
			c.controller.GetLogger().Err(err).Send()
		}
	})

	return nil
}
