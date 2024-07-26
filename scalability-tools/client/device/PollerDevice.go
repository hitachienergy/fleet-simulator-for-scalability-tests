package device

// type PollerDevice struct {
// 	templates.PollerAPI
// 	controller templates.Controller
// 	pollDelay  time.Duration
// }

// func NewPollerDevice(api templates.PollerAPI, controller templates.Controller, pollDelay int) *PollerDevice {
// 	return &PollerDevice{
// 		PollerAPI:  api,
// 		controller: controller,
// 		pollDelay:  time.Duration(pollDelay),
// 	}
// }

// func (d *PollerDevice) Start(ctx context.Context) error {
// 	var err error
// 	defer func() {
// 		// report stats to the central monitor
// 		d.controller.Connect(err == nil)
// 	}()

// 	// setup api
// 	err = d.PollerAPI.Connect()
// 	if err != nil {
// 		return err
// 	}

// 	// register device
// 	err = d.poll()
// 	if err != nil {
// 		return err
// 	}

// 	// start listener
// 	go d.listen(ctx)

// 	return nil
// }

// func (d *PollerDevice) Stop() error {
// 	return d.PollerAPI.Disconnect()
// }

// func (d *PollerDevice) listen(ctx context.Context) {
// out:
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			break out
// 		case <-time.After(d.pollDelay * time.Second):
// 			err := d.poll()
// 			if err != nil {
// 				d.controller.GetLogger().Err(err).Send()
// 			}
// 		}
// 	}
// }

// func (d *PollerDevice) poll() (err error) {
// 	data, err := d.PollerAPI.RetrieveInfo()
// 	if err != nil {
// 		return err
// 	}

// 	switch data.ActionType {
// 	case templates.ACTION_UPDATE:
// 		go func() {
// 			err := d.PollerAPI.StartUpdate(data)
// 			if err != nil {
// 				d.controller.GetLogger().Err(err).Send()
// 			}
// 		}()
// 	}

// 	return nil
// }
