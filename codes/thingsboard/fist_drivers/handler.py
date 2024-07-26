import logging
import os
import time
from dataclasses import dataclass
from hashlib import sha256
from time import sleep
from typing import Optional

from fistsim.server import utils
from fistsim.server.base import ServerHandler as ServerHandlerInterface
from tenant import Tenant

log = logging.getLogger("simulator")


@dataclass
class ServerHandler(ServerHandlerInterface):
    endpoint: str
    docker_compose: str
    clients_name_prefix: str
    clients_number: int
    firmware: Optional[str] = None

    def start_server(self):
        utils.start_server_from_docker_compose(self.docker_compose)
        utils.wait_healhty_http_endpoint(self.endpoint)

        time.sleep(2 * 60)  # TODO why?
        create_devices(self.endpoint, self.clients_name_prefix, self.clients_number)

    def trigger(self):
        if self.firmware:
            trigger_update(self.endpoint, self.firmware)

    def store_and_cleanup(self, output_path, docker_client):
        utils.save_servers_containers_logs(
            output_path, self.docker_compose, docker_client
        )
        utils.stop_server_from_docker_compose(self.docker_compose)

    @classmethod
    def from_dict(cls, data):
        firmware = None
        if "simulation" in data:
            fw_path = data.get("simulation", {}).get("args", {}).get("path")
            fw = data.get("simulation", {}).get("args", {}).get("firmware")
            firmware = os.path.join(fw_path, fw)

        return cls(
            endpoint=data["server"]["endpoint"],
            docker_compose=data["server"]["dockerCompose"],
            clients_name_prefix=data["client"]["namePrefix"],
            clients_number=data["client"]["numberOfDevices"],
            firmware=firmware,
        )


def trigger_update(endpoint: str, firmware: str):
    tenant = Tenant(baseurl=endpoint)
    tenant.login()

    log.info("Creating OTA Update...")
    device_profile_id = tenant.get_default_device_profile_id()

    with open(firmware, "rb") as f:
        data = f.read()
    checksum = sha256(data).digest().hex()
    fw_id = tenant.create_firmware(
        device_profile_id=device_profile_id, checksum=checksum
    )
    tenant.update_firmware(firmware, fw_id)
    sleep(5)

    log.info("Launching OTA Update...")
    tenant.launch_update(fw_id, device_profile_id)


def create_devices(endpoint: str, name_prefix: str, num: int) -> None:
    tenant = Tenant(baseurl=endpoint)
    tenant.login()

    device_profile_id = tenant.get_default_device_profile_id()
    for i in range(1, num + 1):
        token = f"{name_prefix}{i}"
        try:
            tenant.register_device(
                device_name=token,
                access_token=token,
                device_profile_id=device_profile_id,
            )
        except Exception as e:
            log.error(f"Error: {str(e)}")
