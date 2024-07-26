# Platform Drivers

This folder contains everything needed to run a platform. This may include scripts and Docker-related files.

We provide examples of two already integrated platforms:
- The `hawkbit` folder contains Docker compose scripts for the [Eclipse Hawkbit](https://projects.eclipse.org/projects/iot.hawkbit) platform.
- The `thingsboard` folder contains bash and Docker compose scripts for the [ThingsBoard](https://thingsboard.io/) platform.

Moreover, although it is not mandatory, we suggest creating an additional folder named `fist_drivers` that contains implementations for the Simulation Manager component. The path to these is then provided inside the YAML configuration of a simulation.
