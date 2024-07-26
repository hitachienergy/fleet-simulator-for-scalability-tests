package hawkbit

import (
	"context"
	"encoding/json"
	"fmt"
	"hitachienergy/scalability-test-client/templates"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/xerrors"
)

// func RunDMF() {
// 	log.Info().Msg("Start Hawkbit DMF Client...")
// 	basepoint := "guest:guest@localhost:5672"
// 	if len(os.Args) > 1 {
// 		basepoint = os.Args[1]
// 	}
// 	client := NewDMFClient(device.NewSimulationController(&log.Logger, "simulator0", "test",
// 		func(success bool) {
// 			fmt.Println("Connected!")
// 		}, func(id string, start time.Time, duration time.Duration, success bool) {
// 			fmt.Println("Finished")
// 		}),
// 		"DEFAULT", basepoint, "/", "simulator.replyTo", false, 0, map[string]string{
// 			"targetType": "DMF",
// 		})
// 	ctx, cancel := context.WithCancel(context.Background())
// 	err := client.Start(ctx)
// 	if err != nil {
// 		log.Fatal().Err(err).Send()
// 	}

// 	stopSignal := make(chan os.Signal, 1)
// 	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

// 	<-stopSignal
// 	cancel()
// 	log.Info().Msg("Stop Hawkbit DMF Client...")
// }

// ISSUE: DMF devices connection phase is composed of 2 parts: AMQP channel creation + connection, first ping message request and reply
// Using ".client.deviceRegisterMode: sequential" breaks the simulation because we do not wait for the ping part
// Thus the OTA update task will start immediately and for a smaller number of devices, depending on hardware performances
// On faster machines, the problem may never occur

type DMFClient struct {
	controller templates.Controller
	attributes map[string]string

	pingPeriod uint
	pingTracer *pingTracer

	*DMFAmqpService
	*DMFUpdateManager
}

// NewDMFClient creates a new instance of DMF client
func NewDMFClient(controller templates.Controller, tenant string, baseEndpoint string, virtualHost string,
	exchangeName string, useHTTPPool bool, pingPeriod uint, attributes map[string]string) *DMFClient {
	service := newDMFAmqpService(controller.GetIdentifier(), tenant, baseEndpoint,
		virtualHost, exchangeName, useHTTPPool)
	c := &DMFClient{
		controller:     controller,
		pingPeriod:     pingPeriod,
		attributes:     attributes,
		DMFAmqpService: service,
	}
	c.DMFUpdateManager = newDMFUpdateManager(c)
	return c
}

// Start implements Device.Start
func (c *DMFClient) Start(ctx context.Context) (err error) {
	// create AMQP channel
	err = c.DMFAmqpService.startService(ctx)
	if err != nil {
		c.controller.Connect(false)
		return err
	}

	err = c.createDevice(DMFCreate{
		Name: c.id,
		AttributeUpdate: DMFAttributesUpdate{
			Attributes: c.attributes,
			Mode:       ATTRIBUTES_MODE_MERGE,
		},
	})
	if err != nil {
		c.controller.Connect(false)
		return err
	}

	// send ping request and wait for a response
	// wait for exactly 1 ping response and update DeviceController with connect status
	err = c.periodicallyPing(ctx)
	if err != nil {
		c.controller.Connect(false)
		return err
	}
	message := <-c.receiveChann
	if message.Headers[AMQP_KEY_TYPE] != AMQP_TYPE_PING_RESPONSE {
		c.controller.Connect(false)
		return xerrors.Errorf("Received non-ping first message, %s", message.Headers[AMQP_KEY_TYPE])
	}
	err = c.processMsgs(message)
	if err != nil {
		c.controller.GetLogger().Err(err).Send()
		c.controller.Connect(false)
		return err
	}

	// c.controller.Connect(true)

	// messages handling
	go func() {
	out:
		for {
			select {
			case <-ctx.Done():
				break out
			case message := <-c.receiveChann:
				err := c.processMsgs(message)
				if err != nil {
					c.controller.GetLogger().Err(err).Send()
				}
			}
		}
	}()

	return nil
}

// Stop implements Device.Stop
func (c *DMFClient) Stop() error {
	// err := c.stopService()
	// return err
	return nil
}

// periodicallyPing pings the server periodically, or only once if no ping period is set
func (c *DMFClient) periodicallyPing(ctx context.Context) error {
	c.pingTracer = newPingTracer()
	counter := 0
	if c.pingPeriod == 0 {
		id := fmt.Sprintf("%d-%s", counter, c.id)
		counter += 1
		c.pingTracer.add(id)
		// c.controller.GetLogger().Info().Msgf("Send ping %s", id)
		err := c.ping(id)
		return err
	}

	go func() {
	out:
		for {
			select {
			case <-ctx.Done():
				break out
			case <-time.After(time.Duration(c.pingPeriod) * time.Second):
				id := fmt.Sprintf("%d-%s", counter, c.id)
				counter += 1
				c.pingTracer.add(id)
				c.controller.GetLogger().Info().Msgf("Send ping %s", id)
				err := c.ping(id)
				if err != nil {
					c.controller.GetLogger().Err(err).Send()
				}
			}
		}
	}()
	return nil
}

// processMsgs processes AMQP messages based on the message types
func (c *DMFClient) processMsgs(message amqp.Delivery) (err error) {
	if message.Headers[AMQP_KEY_TENANT] != c.tenant {
		return
	}
	switch message.Headers[AMQP_KEY_TYPE] {
	case AMQP_TYPE_EVENT:
		if message.Headers[AMQP_KEY_THING_ID] != c.id {
			return
		}
		return c.handleEvent(message)
	case AMQP_TYPE_PING_RESPONSE:
		return c.handlePing(message)
	}
	return nil
}

// handleEvent handles event messages
func (c *DMFClient) handleEvent(message amqp.Delivery) (err error) {
	c.controller.GetLogger().Debug().Msgf("Receive EVENT message: %s", message.Headers[AMQP_KEY_TOPIC])
	switch message.Headers[AMQP_KEY_TOPIC] {
	case TOPIC_DOWNLOAD:
		return c.handleUpdate(message)
	case TOPIC_DOWNLOAD_AND_INSTALL:
		return c.handleUpdate(message)
	case TOPIC_CANCEL_DOWNLOAD:
		return c.handleCancel(message)
	case TOPIC_MULTI_ACTION:
		return c.handleMultiAction(message)
	case TOPIC_REQUEST_ATTRIBUTES_UPDATE:
		return c.handleAttributeUpdate(message)
	}
	return xerrors.Errorf("Invalid DMF Event Topic: %s", message.Headers[AMQP_KEY_TOPIC])
}

// handlePing handles the ping response messages
func (c *DMFClient) handlePing(message amqp.Delivery) (err error) {
	if c.pingTracer.checkAndDelete(message.CorrelationId) {
		c.controller.Connect(true)
		// c.controller.GetLogger().Info().Msgf("receive ping response %s", message.MessageId)
	}
	return nil
}

// handleAttributeUpdate handles the attributeUpdateRequest and sends an update attribute messages
func (c *DMFClient) handleAttributeUpdate(message amqp.Delivery) (err error) {
	return c.updateAttributes(c.attributes, ATTRIBUTES_MODE_MERGE)
}

// handleUpdate handles the update messages and launches the update
func (c *DMFClient) handleUpdate(message amqp.Delivery) (err error) {
	var action DMFAction
	err = json.Unmarshal(message.Body, &action)
	if err != nil {
		return err
	}

	c.controller.StartTask()
	c.controller.GetScheduler().Submit(
		func() {
			err := c.startUpdate(&action, message.Headers[AMQP_KEY_TOPIC] == TOPIC_DOWNLOAD_AND_INSTALL)
			if err != nil {
				c.controller.GetLogger().Err(err).Send()
			}
		},
	)

	return nil
}

// handleCancel handles the UpdateCancel messages and cancel the related updates
func (c *DMFClient) handleCancel(message amqp.Delivery) (err error) {
	var cancel DMFCancel
	err = json.Unmarshal(message.Body, &cancel)
	if err != nil {
		return err
	}

	exist := c.resetUpdate(cancel.ActionID)
	if !exist {
		return c.sendUpdateFeedback(DMFUpdateFeedback{
			ActionID:     cancel.ActionID,
			ActionStatus: UPDATE_CANCELED,
			Messages:     []string{"Action not processed yet."},
		})
	}
	return c.sendUpdateFeedback(DMFUpdateFeedback{
		ActionID:     cancel.ActionID,
		ActionStatus: UPDATE_CANCELED,
		Messages:     []string{"Simulation canceled."},
	})
}

// handleMultiAction handles the multiAction requests using the message of the highest priority
func (c *DMFClient) handleMultiAction(message amqp.Delivery) (err error) {
	var actions []DMFMultiAction
	err = json.Unmarshal(message.Body, &actions)
	if err != nil {
		return err
	}

	marshaledAction, err := json.Marshal(actions[0])
	if err != nil {
		return err
	}

	newMessage := message
	newMessage.Type = actions[0].Topic
	newMessage.Body = marshaledAction

	return c.handleEvent(newMessage)
}
