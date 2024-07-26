import json
import logging
import os
import subprocess
from abc import abstractmethod
from dataclasses import dataclass, field
from time import sleep
from typing import Dict, List, Tuple

import docker
import requests
import yaml
from docker.errors import APIError, NotFound
from docker.models.containers import Container

log = logging.getLogger("simulator")

CONTAINER_TEMPLATE_PATH = "/app/plugin"
CONTAINER_ENV_CLIENT_NUM = "CLIENT_NUM"
CONTAINER_ENV_IDX_OFFSET = "IDX_OFFSET"
CONTAINER_ENV_INFLUENCE = "INFLUENCE"
CONTAINER_ENV_PORT = "STATUS_SERVER_PORT"

CONTAINER_NAME_PREFIX = "device-simulator"
CONTAINER_PORT_BASE = 8086

SIMULATOR_ENDPOINT_HOST = "localhost"
SIMULATOR_ENDPOINT_PORT = 8086
VALIDATE_PERIOD = 10


class DevicesHandler:
    @abstractmethod
    def setup_devices(self):
        """Creates a number of device manager containers.

        Raises:
            RuntimeError: this exception is generated whenever the initialization process of any of the devices containers fails.
        """
        pass

    @abstractmethod
    def start_devices(self):
        """Starts a number of device manager containers. Depending on the simulaiton configuration containers can be initialized as follows:
        - parallel mode - starts immediately all the containers and then wait for their status to be ready
        - sequential mode - starts containers and waits for the ready status one by one

        Raises:
            RuntimeError: Generated when any of the initialized containers is stopped during the start phase.
        """
        pass

    @abstractmethod
    def wait_finish(self):
        """Periodically checks containers exit code, which is generated when the simulation scenario has been completed.
        In case multiple containers are running, this method waits for the first running container termination and moves to the next one.
        """
        pass

    @abstractmethod
    def store_and_cleanup(self, opath: str):
        """Saves all the relevant data generated during the simulation, then stops and delete each container.

        Args:
            opath (str): the logs destination folder.
        """
        pass


@dataclass
class BaseDevicesHandler(DevicesHandler):
    client: docker.DockerClient
    data: dict
    network_name: str
    num_of_devices: int
    num_of_containers: int
    container_base_port: int
    containers: List[Container]
    finished_containers: List[Container]
    volumes: Dict
    base_env_variabels: Dict

    # simulate will starts all the device containers
    def setup_devices(self):
        try:
            self.client.networks.get(self.network_name)
            log.info(f"Using docker network: ${self.network_name}")
        except NotFound:
            log.info(f"Creating docker network: ${self.network_name}")
            self.client.networks.create(self.network_name)

        self.base_env_variabels["CONFIG"] = yaml.dump(self.data)

        log.info(
            f"Start simulation in container mode. {self.num_of_containers} container(s) with total {self.num_of_devices} device(s) to be created."
        )

        avg_device_per_container = self.num_of_devices // self.num_of_containers
        offset = 1
        influences = self._split_simulation_influence()

        for i in range(self.num_of_containers):
            influence = json.dumps(influences[i])
            log.info(f"Container {i}, {influence}")

            name = f"{CONTAINER_NAME_PREFIX}-{i}"
            self._remove_duplicated_container(name)

            port = self.container_base_port + i
            env_variables = self.base_env_variabels.copy()
            env_variables[CONTAINER_ENV_CLIENT_NUM] = avg_device_per_container
            env_variables[CONTAINER_ENV_IDX_OFFSET] = offset
            env_variables[CONTAINER_ENV_INFLUENCE] = influence
            env_variables[CONTAINER_ENV_PORT] = port

            container = self.client.containers.run(
                "device-simulator",
                detach=True,
                network=self.network_name,
                ports={f"{port}/tcp": port},
                name=name,
                volumes=self.volumes,
                environment=env_variables,
                labels={"com.docker-tc.enabled": "1"},
            )

            if not isinstance(container, Container):
                raise RuntimeError("Container initialization failed")

            offset += avg_device_per_container

            self.containers.append(container)

        self._wait_state(self.containers, "ready")

    def _remove_duplicated_container(self, name: str):
        """Remove existing containers (running or not) by name.

        Args:
            name (str): the name of the container to remove.
        """
        for container in self.client.containers.list(all=True):
            if isinstance(container, Container) and container.name == name:
                log.info(f"Container {name} already exists. Killing it.")
                if container.status == "running":
                    container.kill()
                container.remove()

    def _split_simulation_influence(self) -> List[dict]:
        """Split the task simulated by devices across all containers equally.

        Returns:
            Tuple[dict, dict]: the influece configuration of each container.
        """

        def _split(value, times):
            splitted = [0] * times
            for i in range(times):
                splitted[i] = value // times
            for i in range(value % times):
                splitted[i] += 1
            return splitted

        def _count_affected_devices(by, num_of_clients):
            affected = 0
            if "simulation" in self.data and by in self.data["simulation"]:
                if "number" in self.data["simulation"][by]:
                    affected = int(self.data["simulation"][by]["number"])
                elif "percent" in self.data["simulation"][by]:
                    affected = int(
                        int(self.data["simulation"][by]["percent"][:-1])
                        / 100
                        * num_of_clients
                    )
            return affected

        num_of_clients = self.data["client"]["numberOfDevices"]
        num_of_containers = self.data["client"]["numberOfContainers"]

        num_of_dummy_workers = _count_affected_devices("dummyWork", num_of_clients)
        dummy_workers_per_container = _split(num_of_dummy_workers, num_of_containers)

        num_of_crash_workers = _count_affected_devices("crash", num_of_clients)
        crash_workers_per_container = _split(num_of_crash_workers, num_of_containers)

        return [
            {"dummyWork": i, "crash": j}
            for i, j in zip(dummy_workers_per_container, crash_workers_per_container)
        ]

    def start_devices(self):
        start_mode = self.data["client"].get("containerStartMode", "parallel")
        simulator_endpoint_port = int(
            self.data["client"].get("httpServerBasePort", SIMULATOR_ENDPOINT_PORT)
        )
        for i, c in enumerate(self.containers):
            port = simulator_endpoint_port + i
            response = requests.post(f"http://{SIMULATOR_ENDPOINT_HOST}:{port}/start")
            if response.status_code == 200:
                log.info(f"[{c.name}] Devices connection to server started")
            else:
                log.error(f"[{c.name}] Failed to start devices connection to server")
                raise RuntimeError()

            if start_mode == "sequential":
                log.info("Waiting for container connection to the server...")
                self._wait_state([c], "connected", i)

        if start_mode == "parallel":
            log.info("Waiting for containers connection to the server...")
            self._wait_state(self.containers, "connected")

    def _wait_state(self, containers, endpoint: str, port_offset: int = 0):
        """Utility method that waits for a set of containers state to be reached.
        The containers state is retrieved through different HTTP endpoints.

        Args:
            containers (_type_): the method will whait for these containers to reach a specific state.
            endpoint (str): The HTTP endpoint provided by each container.
            port_offset (int, optional): Configure the first container port to use. Defaults to 0.

        Raises:
            RuntimeError: generated when an http request fails or a container is stopped unexpectedly.
        """
        simulator_endpoint_port = int(
            self.data["client"].get("httpServerBasePort", SIMULATOR_ENDPOINT_PORT)
        )
        for i, container in enumerate(containers):
            port = simulator_endpoint_port + i + port_offset
            while True:
                url = f"http://{SIMULATOR_ENDPOINT_HOST}:{port}/{endpoint}"
                try:
                    response = requests.get(url)
                    if response.status_code == 200:
                        log.debug(f"[{container.name}] Positive response from {url}.")
                        break
                    elif response.status_code == 503:
                        log.debug(
                            f"[{container.name}] Negative response from {url}. Waiting..."
                        )
                    else:
                        log.error(
                            f"[{container.name}] Unexpected HTTP code {response.status_code}: {response.text}"
                        )
                        raise RuntimeError()
                except Exception as e:
                    container.reload()
                    if container.status != "exited":
                        log.debug(
                            f"[{container.name}] Negative response from {url}. Waiting..."
                        )
                    else:
                        raise RuntimeError(f"Devices manager fail: {str(e)}")
                sleep(VALIDATE_PERIOD)

    def wait_finish(self):
        if len(self.containers) == 0:
            return

        for container in self.containers:
            exitCode = container.wait()
            if exitCode["StatusCode"] == 0:
                log.info(f"{container.name} finished successfully")
                self.finished_containers.append(container)
            else:
                log.info(f"{container.name} ends with error. Exit code: {exitCode}")

    def store_and_cleanup(self, opath: str):
        self._store_logs(opath)
        self._store_stats(opath)
        self._stop()
        self._clean_up()

    def _store_logs(self, opath: str):
        for container in self.containers:
            cname = str(container.name)
            cpath = os.path.join(opath, cname)
            os.makedirs(cpath)
            data = container.logs()
            with open(f"{cpath}/logs.txt", "wb") as f:
                f.write(data)
            log.info(f"[{container.name}] Logs saved to {cpath}")

    def _store_stats(self, opath: str):
        for container in self.finished_containers:
            cname = str(container.name)
            cpath = os.path.join(opath, cname)
            executeResult = subprocess.run(
                f"docker cp {container.name}:/app/results/. {cpath}".split(" "),
                stdout=subprocess.PIPE,
            )
            if executeResult.returncode != 0:
                log.error(
                    f"Fail to save simulation analysis (code: {executeResult.returncode}): {executeResult.args}"
                )
            else:
                log.info(f"[{container.name}] Analysis saved to {cpath}")

    def _stop(self):
        simulator_endpoint_port = int(
            self.data["client"].get("httpServerBasePort", SIMULATOR_ENDPOINT_PORT)
        )
        for i, container in enumerate(self.containers):
            container.reload()
            if container.status == "running":
                log.info(f"[{container.name}] Stopping container... ")
                requests.post(
                    f"http://{SIMULATOR_ENDPOINT_HOST}:{simulator_endpoint_port + i}/stop"
                )
                try:
                    # stop the container itself
                    container.stop(timeout=10)
                    log.info(f"[{container.name}] Container stopped ")
                except APIError as e:
                    log.error(
                        f"[{container.name}] Container stopped with error: {str(e)}"
                    )
            else:
                log.info(f"[{container.name}] Container already stopped ")

    def _clean_up(self):
        for container in self.containers:
            container.remove(force=True)

    @classmethod
    def from_dict(cls, data: dict, client: docker.DockerClient):
        """_summary_

        Args:
            data (dict): _description_
            client (docker.DockerClient): _description_

        Raises:
            Exception: _description_
            ValueError: _description_

        Returns:
            _type_: _description_
        """
        num_of_devices = data["client"]["numberOfDevices"]
        if num_of_devices <= 0:
            raise Exception(
                "The number of devices (client.number) must be greater than zero"
            )

        num_of_containers = data["client"].get("numberOfContainers", 1)
        if num_of_containers <= 0:
            raise ValueError("The number of containers must be greater than zero")

        network_name = data["client"].get("network", "host")
        template_path = os.path.abspath(data["client"]["template"])
        volumes = {
            template_path: {
                "bind": CONTAINER_TEMPLATE_PATH,
                "mode": "ro",
            },
        }

        return cls(
            client=client,
            data=data,
            network_name=network_name,
            num_of_devices=num_of_devices,
            num_of_containers=num_of_containers,
            container_base_port=data["client"].get(
                "httpServerBasePort", CONTAINER_PORT_BASE
            ),
            containers=list(),
            finished_containers=list(),
            volumes=volumes,
            base_env_variabels=dict(),
        )
