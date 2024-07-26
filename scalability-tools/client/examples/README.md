# Device examples

## Required Interface 

The user-implemented device will be imported into the program at runtime as a plugin. 
It should contain two part: `device constructor` and `device`.

Devices constructor instances must implement the [DeviceFactory](../templates/DeviceFactory.go#DeviceFactory) interface.
Devices instances must implement the [Client](../templates/Device.go#Device) interface.

Look at the examples provided for the Thingsboard and Hawkbit platforms on how to implement all the required elements:
- [./thingsboard/HTTPDefault.go](./thingsboard/HTTPDefault.go)
- [./hawkbit/DDIDefault.go](./hawkbit/DDIDefault.go)
- [./hawkbit/DMFDefault.go](./hawkbit/DMFDefault.go)
