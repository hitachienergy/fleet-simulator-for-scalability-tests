# FIST
A plugin-and-play scalability test tool for fleet management systems

## System Architecture
---
The system is constituted by four main part. 

- *Simulation Manager*: the entry point of the whole simulator. It controls the whole simulation process.

- *Server Containers*: the target platform to be tested

- *Fleet Simulator Containers*: the containers where the device simulators will run ([link](client/README.md))

- *Docker-TC*: a network simulator that works specifically for docker containers. It can work for both server contaienrs and device simulator containers ([link](https://github.com/lukaszlach/docker-tc))

![architecture](docs/architecture.png)

## How to Run 
---
### Dependency
Python version: 3.10.12

Docker version: 24.0.4. (API version: 1.43)

### Run the Simulator 
```bash
# build fleet-simulator image
cd scalability-tools/client
./build_image.sh

## run the whole simulator
cd ../..
python3 simulator.py --configFile <path_to_configure_file>
```


