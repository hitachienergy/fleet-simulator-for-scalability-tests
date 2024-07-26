# Device-Simulation Framework

## System Architecture

The Device Simulator consists of three major parts:

- *Fleet Manager*: This is the entry point to start the simulators. It controls and monitors the lifecycle of all the Device Simulators.

- *Device Simulator*: This is a simulator for an IoT device that is able to communicate with the fleet management platform and perform OTA updates. There could be multiple device simulators in a single simulation.

- *HTTP Server*: External users and programs can get the current simulation stats by sending HTTP requests to the server.

### Device Simulator 

A single Device Simulator has two important components: *API* and *Controller*.

- *API*: This is a platform-specific API that enables the simulator to communicate with different platforms and performs different task execution logics. It will be loaded dynamically during runtime.

- *Controller*: A controller holds all the details a device might need during the simulation. It also provides functionalities for dummy tasks or crash simulations and produces well-formatted logs and stats.

![architecture](documents/docs/simulator.png)

## Getting Started

### Dependencies

golang: 1.20.5

### How to Run 

Build the simulator locally:
```bash
go build -o simulator .
```

Build the docker image:
```bash
./build_image.sh
```

### Interact with the simulator

By default, the HTTP server is running on [http://localhost:8086](http://localhost:8086)

| Content | Link | 
| --- | --- |
| Registration Completion | /ready |
| OTA Update Stats | /stats |
| Stop simulation | /stop |


## Device Implementation

Here are the [examples](examples) of Eclipse Hawkbit and Thingsboard IoT platforms