# YAML Configurations

This folder contains the FIST simulations configurations in [YAML](https://yaml.org/) format.

This configuration file should define the following elements:
- `target`: the symbolic name of the IoT platform that will be used. This value is mostly used inside the Python simulator manager and Go devices manager to select the corresponding handler.
- [optional] `timeout`: this value defines how much time a single simulation should run before being timed out. Values should follow the syntax \d+(s|m|h|d|w).
- [optional] `logLevel`: another utility field to globally define the log level used by all the processes. It's important to note that both the simulation manager and devices manager already integrate this value but not the IoT platform.
- `server`: all the configurable server fields.
- `client`: all the configurable client fields.
- `network`: network simulation policies.
- [optional] `simulation`: the triggerable simulation scenario and devices behavior. If not defined, the FIST simulator will only attempt to connect simulated devices to the IoT platform.
- `output`: the output folder where to save results, logs, and simulation statistics.

The `server` block must contain the following fields:
- `driver`: where to find the IoT platform custom drivers.

The `client` block must contain the following fields:
- `containerStartMode`: how to start containers ("parallel" or "sequential" modes).
- `devicesRegisterMode`: how devices register to the IoT platform ("parallel" or "sequential" modes).
- `numberOfContainers`: the number of containers that will be created.
- `numberOfDevices`: the number of devices that will be created and partitioned across containers.

The `network` tag must contain one or more of the following fields: `corrupt`, `delay`, `duplicate`, `drop`, `rate`.
For more details on each value semantic, look at the docker-tc [GitHub](https://github.com/lukaszlach/docker-tc) repository documentation.

The `simulation` can contain the following elements:

- `task`: the symbolic name of the task.
- `args`: the additional parameters associated with the task name.
- `dummyWork`: additional device behavior to simulate dummy work tasks.
  - `number`: number of devices will execute the dummy task (mutually exclusive with `percent`).
  - `percent`: percentage of devices will execute the dummy task (mutually exclusive with `number`).
  - `duration`: duration of the dummy task.
  - `variation`: variation applied randomly to the duration of a dummy task.
  - `period`: how often the dummy task will be repeated during the real task scenario execution.
- `crash`: additional parameter associated with random device crash.
  - `number`: number of devices will execute the dummy task (mutually exclusive with `percent`).
  - `percent`: percentage of devices will execute the dummy task (mutually exclusive with `number`).
  - `within`: from the start of the main task, how much time will elapse before a crash occurs.
- `seed`: random generation seed.

For instance, we provide the configuration of 2 IoT platforms and various scenarios: [Eclipse Hawkbit](hawkbit) and [Thingsboard](thingsboard).


