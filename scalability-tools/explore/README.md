# FIST Results Explorer

This Python project implements utility functions for easy results exploration, logging and plotting.

## How to Run 

This project is intended to be build and installed on a local Python environment.

Build the module:
```
./build.sh
```

## Usage 

### Statistics

Imports:
```python
from fistexplore import simulation_stats as ss
```

Load results statistics:
```python
ROOT="results/hawkbit-ddi/test-case-1"

files = ss.get_all_files_from_root(ROOT)
configs = ss.get_all_config_files(files)
all_stats = ss.get_simulations_task_stats(configs)
compl_stats, compl_agg_stats = ss.get_completed_task_simulations(all_stats)
```

If the specified folder contains a .yaml file, we assume that there are results to load.

It is also possible to load multiple scenarios at once by selecting a parent folder:
```python
ROOT="results/hawkbit-ddi"

// same implementation
```

### Docker stats

Since each scenario may use considerably different fields and configurations, we do not implement any platform specific result post-processing.

Imports:
```python
from fistexplore import docker_stats as ds
from fistexplore import logs as l
from fistexplore import plot_docker_stats as pds
```

Plot docker stats:
```python
ROOT = "results/hawkbit-ddi/test-case-1"

docker_stats = ds.load_stats(ROOT + "docker_stats.csv")
```

There are many ways to plot resources usage. 
The following example compares multiple docker containers stats:
```python
pds.plot_stats_comparison_v2(docker_stats)
```

The following example plots additional informations related to the stats and simulation state:
```python
logs = l.load_logs(ROOT + "simulator_logs.log")
pds.plot_container_stats(docker_stats, logs)
```

Unwanted containers can be filtered as follows:
```python
pds.filter_containers = ["docker-tc", "device-simulator-0"]
``` 