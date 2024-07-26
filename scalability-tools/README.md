# Scalability tools

This folder contains all the necessary tools to set up and run a FIST simulation.

The `scalability-tools` folder is composed of:

- **Client Subproject** - The Golang implementation of the devices manager. This is where we simulate realistic devices and device task execution.
- **Docker-tc-patch Folder** - Contains instructions on how to set up the network environment simulation. In particular, here we provide a patch for the open-source project docker-tc with minor component setup changes.
- **Examples** - Contains all the YAML configurations used to define how a simulation runs, where to find resources in the file system. These files provide an easy way to configure your test scenarios.
- **Explore Subproject** - A Python utility for simulation results exploration. Since a simulation generates a lot of files and statistics, this component makes it easy to access results and plot resource usage.
- **Scripts** folder - Contains utility bash files and provides automatic test runs.
- **Simulator** - The Python project that implements a simulation workflow given an input YAML file (provided in the `scalability-tools/examples` folder).

The [simulator](simulator) folder contains the FIST Simulator Manager and the entry point of a simulation, written in Python. 
The [client](client) folder implements the Fleet Manager (devices simulation), written in GoLang.
The [docker-tc](docker-tc) folder implements the network environment simulation. This folder contains a patch for a [public GitHub repository](https://github.com/lukaszlach/docker-tc) with minor code changes. For more details and instructions on how to set up docker-tc, look at the docker-tc-patch folder.
The [explore](explore) folder implements some utility functions for simulations results exploration, such as stats extraction and plotting.
The [scripts](scripts) folder implements some Bash scripts used to execute multiple runs autonomously.