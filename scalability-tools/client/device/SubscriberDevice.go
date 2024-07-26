package device

// type SubscriberDevice struct {
// 	templates.SubscriberAPI
// 	controller templates.Controller
// }

// func NewSubscriberDevice(api templates.SubscriberAPI, controller templates.Controller, pollDelay int) *SubscriberDevice {
// 	return &SubscriberDevice{
// 		SubscriberAPI: api,
// 		controller:    controller,
// 	}
// }

// func (d *SubscriberDevice) Start(ctx context.Context) error {
// 	var err error
// 	defer func() {
// 		// report stats to the central monitor
// 		if err != nil {
// 			d.controller.Connect(false)
// 		}
// 	}()

// 	err = d.SubscriberAPI.Connect()
// 	if err != nil {
// 		return err
// 	}

// 	err = d.SubscriberAPI.RegisterDevice()
// 	if err != nil {
// 		return err
// 	}

// 	go d.handler(ctx)

// 	return nil
// }

// func (d *SubscriberDevice) Stop() error {
// 	return d.SubscriberAPI.Disconnect()
// }

// func (d *SubscriberDevice) handler(ctx context.Context) {
// 	msgChan := d.SubscriberAPI.GetMessageChannel()
// out:
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			break out
// 		case action := <-msgChan:
// 			err := d.handleAction(action)
// 			if err != nil {
// 				d.controller.GetLogger().Err(err).Send()
// 			}
// 		}
// 	}
// }

// func (d *SubscriberDevice) handleAction(data templates.Action) error {
// 	switch data.ActionType {
// 	case templates.ACTION_UPDATE:
// 		go func() {
// 			err := d.SubscriberAPI.StartUpdate(data)
// 			if err != nil {
// 				d.controller.GetLogger().Err(err).Send()
// 			}
// 		}()
// 	}

// 	return nil
// }
