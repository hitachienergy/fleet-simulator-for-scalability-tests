package device

// type DeviceType int

// const (
// 	BASE DeviceType = iota
// 	UPDATE
// )

// // Device is a simulation of one end device
// type Device struct {
// 	templates.Device
// 	*SimulationController
// }

// // NewDevice creates a new instance of Device
// func NewDevice(controller *SimulationController, config config.Config, clientFactory templates.DeviceFactory) (s *Device, err error) {
// 	client, err := clientFactory.NewDevice(controller)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s = &Device{
// 		Device:               client,
// 		SimulationController: controller,
// 	}
// 	return s, nil
// }

// // Start starts the device and connects it to the remote platform
// func (d *Device) Start(sharedCtx context.Context) error {
// 	deviceCtx := d.configureDeviceCtx(sharedCtx)
// 	return d.Device.Start(deviceCtx)
// }
