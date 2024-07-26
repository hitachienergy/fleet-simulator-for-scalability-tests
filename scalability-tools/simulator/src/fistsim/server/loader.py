import importlib.util
import logging
import os
import sys
from dataclasses import dataclass

from .base import ServerHandler

log = logging.getLogger("simulator")


def _import(script_path: str):
    # to resolve imported script local imports
    sys.path.append(os.path.dirname(script_path))

    module_name = "imported_module"
    spec = importlib.util.spec_from_file_location(module_name, script_path)

    if not spec or not spec.loader:
        raise ImportError(f"Specified module is not importable {script_path}")

    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    return module.ServerHandler


@dataclass
class ImportableServerHandler(ServerHandler):
    server_handler: ServerHandler

    def start_server(self):
        self.server_handler.start_server()

    def trigger(self):
        self.server_handler.trigger()

    def store_and_cleanup(self, output_path, docker_client):
        self.server_handler.store_and_cleanup(output_path, docker_client)

    @classmethod
    def from_dict(cls, data):
        if "server" not in data or "driver" not in data["server"]:
            raise ImportError("Unspecified importable IoT drivers")
        asd = _import(data["server"]["driver"]).from_dict(data)
        return cls(server_handler=asd)
