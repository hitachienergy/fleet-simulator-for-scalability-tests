import http.client
import logging
import os
from time import sleep
from typing import Container

import docker
import requests
from docker.errors import NotFound
from docker.models.containers import Container

log = logging.getLogger("simulator")
DOCKER_TC_NAME = "docker-tc"
DOCKER_TC_IMAGE = "lukaszlach/docker-tc"

DOCKER_TC_HOST = "localhost"
DOCKER_TC_PORT = 4080
DOCKER_TC_BASE_URL = f"{DOCKER_TC_HOST}:{DOCKER_TC_PORT}"

VALIDATE_PERIOD = 5


def _get_burst_from_rate(v: str):
    # LINUX_KERNEL_HZ = 100  # getconf CLK_TCK # user level linux hz rate
    LINUX_KERNEL_HZ = 250  # egrep '^CONFIG_HZ_[0-9]+' /boot/config-`uname -r` # kernel level linux hz rate
    measure, unit = "", ""
    for i, c in enumerate(v):
        if not c.isdigit():
            measure, unit = v[:i].strip(), v[i:].strip()
            break

    # converts docker-tc rate string into integer representing bits per second
    valid_units = {
        "bps": 2**0 * 8,
        "kbps": 2**10 * 8,
        "mbps": 2**20 * 8,
        "gbps": 2**30 * 8,
        "bitps": 2**0,
        "kbitps": 2**10,
        "mbitps": 2**20,
        "gbitps": 2**30,
    }

    if unit not in valid_units:
        raise ValueError(f"Unit {unit} not handled by docker-tc")

    # https://unix.stackexchange.com/questions/100785/bucket-size-in-tbf
    bit = int((float(measure) * valid_units[unit]) / LINUX_KERNEL_HZ)
    byte = bit / 8
    kilobyte = int(byte / 1024 + 1) * 5
    return f"{kilobyte}kb"


class NetworkHandler:
    """This class provides handlers and interfaces to the docker-tc container which implements network simulation features."""

    def __init__(self, client: docker.DockerClient, data: dict):
        self.client = client
        self.data = data

    def init(self):
        """Starts the docker-tc container.

        Raises:
            RuntimeError: raised when the container suddenly terminates during the initialization phase.
        """
        log.info("Init network simulation...")

        environment = {}
        if "network" in self.data and "rate" in self.data["network"]:
            rate = self.data["network"]["rate"]

            burst = _get_burst_from_rate(rate)
            environment = {
                "TC_QDISC_TBF_BURST": burst,
                "TC_QDISC_TBF_LATENCY": "50ms",
            }

        self.container = self.client.containers.run(
            DOCKER_TC_IMAGE,
            name=DOCKER_TC_NAME,
            network_mode="host",
            restart_policy={"Name": "always"},
            cap_add=["NET_ADMIN"],
            detach=True,
            volumes={
                "/var/run/docker.sock": {"bind": "/var/run/docker.sock", "mode": "rw"},
                "/var/docker-tc": {"bind": "/var/docker-tc", "mode": "rw"},
            },
            environment=environment,
        )

        while True:
            sleep(VALIDATE_PERIOD)
            self.container.reload()
            if self.container.status == "running":
                break
            elif self.container.status == "exited":
                raise RuntimeError("Fail to simulate network conditions")
        log.info("Network simulator container 'docker-tc' started")

    def add(self, containers: list):
        """Adds containers into the affected list of docker-tc.

        Args:
            containers (list): _description_
        """
        for container in containers:
            response = requests.put(f"http://{DOCKER_TC_BASE_URL}/{container.short_id}")
            if response.status_code != 200:
                log.error(f"Fail to scan '{container.short_id}': {response.text}")

    # simulate adds the network conditions to the listed contaienrs
    def simulate(self, config: dict, containers: list):
        """Initializes the docker-tc container and applies network limiting rules to a list of containers.

        Args:
            config (dict): a dictionary of docker-tc rules
            containers (list): the list of containers that will be directly affected by the provided rules.

        Raises:
            RuntimeError: if the docker-tc container is not able to apply network limiting rules to any of the provided containers.
        """
        self.init()

        for container in containers:
            data = "&".join([f"{k}={v}" for k, v in config.items()])
            response = requests.post(
                f"http://{DOCKER_TC_BASE_URL}/{container.name}", data=data
            )
            if response.status_code != 200:
                log.error(
                    f"Fail to set network config for '{container.name}': {response.text}"
                )

        conn = http.client.HTTPConnection(DOCKER_TC_HOST, DOCKER_TC_PORT)
        conn.request("LIST", "/")
        response = conn.getresponse()
        if response.status != 200:
            raise RuntimeError("Unable to connect to network simulator")
        log.info("Network simulation is set up. Details:")
        log.info(response.read().decode("utf-8").rstrip())

    def remove_stale_container(self):
        """Removes any residual docker-tc container."""
        try:
            container = self.client.containers.get(DOCKER_TC_NAME)
            if container and isinstance(container, Container):
                container.remove(force=True)
                log.info("Removed stale docker-tc container")
        except NotFound:
            return

    def store_and_cleanup(self, opath: str):
        """Stores logs generated by the docker-tc container during the simulation.

        Args:
            opath (str): the logs destination folder.
        """
        try:
            container = self.client.containers.get(DOCKER_TC_NAME)
            if container and isinstance(container, Container):
                cpath = os.path.join(opath, str(container.name))
                os.makedirs(cpath)
                data = container.logs()
                with open(f"{cpath}/logs.txt", "wb") as f:
                    f.write(data)
                log.info(f"[{container.name}] Logs saved to {cpath}")
                container.remove(force=True)
        except NotFound:
            return
