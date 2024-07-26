import importlib.util
import logging
import os
import sys
from dataclasses import dataclass

import docker

from .base import BaseDevicesHandler, DevicesHandler

log = logging.getLogger("simulator")


def _import(script_path: str):
    # to resolve imported script local imports
    sys.path.append(os.path.dirname(script_path))

    module_name = "imported_module_2"
    spec = importlib.util.spec_from_file_location(module_name, script_path)

    if not spec or not spec.loader:
        raise ImportError(f"Specified module is not importable {script_path}")

    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    return module.DevicesHandler


@dataclass
class ImportableDevicesHandler(DevicesHandler):
    devices_handler: BaseDevicesHandler

    def setup_devices(self):
        self.devices_handler.setup_devices()

    def start_devices(self):
        self.devices_handler.start_devices()

    def wait_finish(self):
        self.devices_handler.wait_finish()

    def store_and_cleanup(self, opath: str):
        self.devices_handler.store_and_cleanup(opath)

    def get_containers(self):
        return self.devices_handler.containers

    @classmethod
    def from_dict(cls, data: dict, client: docker.DockerClient):
        if "client" not in data or "driver" not in data["client"]:
            log.warn(
                "No devices handler driver specified, using BaseDeviceHandler implementation instead"
            )
            return cls(devices_handler=BaseDevicesHandler.from_dict(data, client))
        else:
            return cls(
                devices_handler=_import(data["client"]["driver"]).from_dict(
                    data, client
                )
            )
