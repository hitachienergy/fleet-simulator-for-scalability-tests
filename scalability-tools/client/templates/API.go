package templates

/*
Nothing in this file is unused or referenced in the project
*/

// type ActionType int

// const (
// 	ACTION_IDLE ActionType = iota
// 	ACTION_UPDATE
// )

// type Action struct {
// 	ActionType ActionType
// 	Data       interface{}
// }

// // DeviceAPI represents the API of the general devices
// type DeviceAPI interface {
// 	StartUpdate(Action) error
// 	Connect() error
// 	Disconnect() error
// }

// // PollerAPI represents the API of the devices that poll the platform periodically for new information (e.g. HTTP devices)
// type PollerAPI interface {
// 	DeviceAPI
// 	RetrieveInfo() (Action, error)
// }

// // DeviceAPI represents the API of the devices that subscribes a topic and wait the platform to push messages to them (e.g. AMQP, MQTT)
// type SubscriberAPI interface {
// 	DeviceAPI
// 	RegisterDevice() error
// 	GetMessageChannel() chan Action
// }

// // APIFactory is a factory to create device apis
// type APIFactory interface {
// 	NewAPI(Controller, string, map[string]interface{}) (DeviceAPI, error)
// }
