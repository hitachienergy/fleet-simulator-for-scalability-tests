package hawkbit

import (
	"context"
	"hitachienergy/scalability-test-client/templates"
	"time"
)

// func RunDDI() {
// 	log.Info().Msg("Start Hawkbit DDI Client...")
// 	gatewayToken := ""
// 	if len(os.Args) > 1 {
// 		gatewayToken = os.Args[1]
// 	}
// 	basepoint := "localhost:8080"
// 	if len(os.Args) > 2 {
// 		basepoint = os.Args[2]
// 	}
// 	client := NewDDIClient(device.NewSimulationController(&log.Logger, "simulator0", "test",
// 		func(success bool) {
// 			fmt.Println("Connected!")
// 		}, func(id string, start time.Time, duration time.Duration, success bool) {
// 			fmt.Println("Finished")
// 		}),
// 		"DEFAULT", 3, basepoint, gatewayToken, false)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	client.Start(ctx)

// 	stopSignal := make(chan os.Signal, 1)
// 	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

// 	<-stopSignal
// 	cancel()
// 	log.Info().Msg("Stop Hawkbit DDI Client...")
// }

type DDIClient struct {
	pollDelay time.Duration

	*DDIRestApi
	*DDIUpdateManager
	controller templates.Controller
}

// NewDDIClient creates a new DDI client instance
func NewDDIClient(controller templates.Controller, tenant string, pollDelay int, baseEndpoint string, gatewayToken string, useHTTPPool bool) *DDIClient {
	api := newDDIRestApi(controller.GetIdentifier(), tenant, baseEndpoint, gatewayToken, useHTTPPool)
	c := DDIClient{
		DDIRestApi: api,
		controller: controller,
		pollDelay:  time.Duration(pollDelay),
	}
	c.DDIUpdateManager = newDDIUpdateManager(&c)
	return &c
}

// Start implements Device.Start interface
func (c *DDIClient) Start(ctx context.Context) error {
	err := c.poll()
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

// Stop implements Device.Stop interface
func (c *DDIClient) Stop() error {
	return nil
}

// poll retrieves information from the server and do the update if needed
func (c *DDIClient) poll() (err error) {
	// link, err := c.GetRequiredLink(ConfirmationBase)
	// if err != nil {
	// 	return err
	// }
	// if len(link) > 0 {
	// 	actionID, err := GetActionId(link)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = c.SendConfirmation(actionID, "confirmed")
	// 	return err
	// }

	link, err := c.getRequiredLink(DeploymentBase)
	if err != nil || len(link) == 0 {
		return err
	}

	actionID, err := getActionId(link)
	if err != nil {
		return err
	}

	deployment, err := c.getActionWithDeployment(actionID)
	if err != nil || deployment == nil {
		return err
	}

	c.controller.StartTask()

	c.controller.GetScheduler().Submit(func() {
		err := c.startUpdate(actionID, deployment)
		if err != nil {
			c.controller.GetLogger().Err(err).Send()
		}
	})

	return nil
}
