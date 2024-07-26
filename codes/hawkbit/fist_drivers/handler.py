import logging
import os
import sys
import time
from dataclasses import dataclass
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
    number_of_devices: int
    gateway_token: Optional[str] = None
    firmware: Optional[str] = None

    def start_server(self):
        utils.start_server_from_docker_compose(self.docker_compose)
        utils.wait_healhty_http_endpoint(self.endpoint)

        create_devices(self.endpoint, self.gateway_token)

    def trigger(self):
        if self.firmware:
            trigger_update(
                self.endpoint,
                self.firmware,
                self.clients_name_prefix,
                self.number_of_devices,
            )

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
            number_of_devices=data["client"]["numberOfDevices"],
            gateway_token=data.get("client", {})
            .get("args", {})
            .get("gatewayToken", None),
            firmware=firmware,
        )


def create_devices(endpoint: str, gateway_token: Optional[str]):
    try:
        log.info("Initialize server configuration...")
        tenant = Tenant(baseurl=endpoint)
        tenant.setDefaultConfig()

        if gateway_token is not None:
            log.info("Set gateway token...")
            tenant.setGatewayToken(gateway_token)
    except Exception as e:
        log.info(f"Error: {str(e)}")
        sys.exit(1)


def trigger_update(
    endpoint: str, firmware: str, client_name_prefix: str, number_of_devices: int
):
    # try:
    tenant = Tenant(baseurl=endpoint)

    log.info("Creating OTA Update...")
    sw_module_id = tenant.createSWModule()[0]
    artifact = firmware
    tenant.uploadArtifacts(sw_module_id, artifact)
    ds_id = tenant.createDistributionSet([sw_module_id])

    # checks if every device is connected and ready to get the firmware
    log.info("Checking targets status...")
    ready_count = tenant.getRegisteredTargetsCount()
    while ready_count < number_of_devices:
        log.debug(
            f"Targets are not ready [{ready_count}/{number_of_devices}]. Waiting..."
        )
        time.sleep(10)
        ready_count = tenant.getRegisteredTargetsCount()
    log.info(f"Targets are ready")

    log.info("Launching OTA Update...")
    _ = tenant.createRollout(f"name=={client_name_prefix}*", ds_id)
    # sleep(2)    # TODO: check how long to wait
    # tenant.startRollout(rollout_id)
    # except Exception as e:
    #     log.info(f"Error: {str(e)}")
    #     sys.exit(1)
