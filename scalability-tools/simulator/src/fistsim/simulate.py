import logging
import os
import traceback
from enum import Enum
from threading import Lock

import docker
from device.loader import ImportableDevicesHandler
from network import NetworkHandler
from server.loader import ImportableServerHandler

logging.basicConfig(format="%(asctime)s - %(levelname)s - %(message)s")
log = logging.getLogger("simulator")


class SimulationState(Enum):
    """Enumeration representing the possible states of a simulation."""

    STARTING = 0
    SETUP = 1
    RUNNING = 2
    STOPPING = 3
    STOPPED = 4


class Simulation:
    """This class implements the management logic of a simulation.
    A FIST simulation process is composed by a sequence of steps:
    - initialization and input config handling
    - iot platform startup
    - devices simulator startup
    - network environment setup
    - devices connection to iot platform
    - [optional] trigger of a simulation scenario
    - wait completion
    - results backup
    - cleanup

    The handling of net environment, devices and iot platform is defined in external classes.
    """

    def __init__(self, data, opath):
        self.opath = opath
        self.data = data

        log.setLevel(logging.getLevelName(self.data.get("logLevel", "info").upper()))

        self.lock = Lock()
        self.status = SimulationState.STARTING
        self.client = docker.from_env()

        self.server_handler = ImportableServerHandler.from_dict(data)
        self.devices_handler = ImportableDevicesHandler.from_dict(data, self.client)
        self.network_handler = NetworkHandler(self.client, self.data)

    def run(self):
        """Main starting point of a simulation scenario."""
        try:
            self._setup()
            self._trigger()
        except Exception as e:
            log.error(f"Error: {str(e)}")
            log.error(traceback.print_exc())
            log.error("Stopping simulation ...")
        finally:
            self._stop_and_cleanup()

    def _setup(self):
        self.status = SimulationState.SETUP

        log.info("###### Setup Server Configuration ######")
        self.server_handler.start_server()

        log.info("###### Start Device Simulators ######")
        self.devices_handler.setup_devices()

        log.info("###### Setup Network Environment ######")
        self.network_handler.remove_stale_container()
        if "network" in self.data:
            affected_containers = self.devices_handler.get_containers()
            self.network_handler.simulate(self.data["network"], affected_containers)

    def _trigger(self):
        self.status = SimulationState.RUNNING

        log.info("###### Start Devices Connection to Server ######")
        self.devices_handler.start_devices()

        if "simulation" in self.data and "task" in self.data["simulation"]:
            log.info("###### Trigger Tests ######")
            self.server_handler.trigger()
            log.info("###### Wait task completion ######")
            self.devices_handler.wait_finish()

    def _stop_and_cleanup(self):
        with self.lock:
            if self.status in [
                SimulationState.STOPPING,
                SimulationState.STOPPED,
            ]:  # Prevents concurrent _stop_and_cleanup calls
                return
            self.status = SimulationState.STOPPING

        log.info("###### Store Data and Clean Up ######")
        client_opath = os.path.join(self.opath, "devices")
        self.devices_handler.store_and_cleanup(client_opath)

        server_opath = os.path.join(self.opath, "server")
        self.server_handler.store_and_cleanup(server_opath, self.client)

        if "network" in self.data:
            network_opath = os.path.join(self.opath, "network")
            self.network_handler.store_and_cleanup(network_opath)

        with self.lock:
            self.status = SimulationState.STOPPED
