import logging
import os
import subprocess
import time
from typing import Any, Union

import requests

log = logging.getLogger("simulator")
CONTAINER_FILTER_LABEL = "label"
CONTAINER_FILTER_KEY = "com.docker.compose.project.config_files"
VALIDATE_PERIOD = 10


def run_script(path: str, cwd: Union[str, None] = None, env: Any = None):
    res = subprocess.run(path, cwd=cwd, env=env)
    if res.returncode != 0:
        raise Exception(f"Subprocess returned with code {res.returncode}")
    return


def start_server_from_docker_compose(composefile: str):
    log.info("Starting server...")
    result = subprocess.run(f"docker compose -f {composefile} up -d".split(" "))
    if result.returncode != 0:
        raise Exception("Fail to start server containers")


def stop_server_from_docker_compose(composefile: str):
    log.info("Stopping server...")
    result = subprocess.run(f"docker compose -f {composefile} down -v".split(" "))
    if result.returncode != 0:
        raise Exception("Fail to stop server containers")


def stop_server_from_script(composefile: str):
    log.info("Stopping server...")
    result = subprocess.run(f"docker compose -f {composefile} down -v".split(" "))
    if result.returncode != 0:
        raise Exception("Fail to stop server containers")


def wait_healhty_http_endpoint(endpoint: str):
    while True:
        try:
            response = requests.get(f"http://{endpoint}")
            if response.status_code == 200:
                log.info("Server is ready.")
                return
        except Exception as e:
            log.info(f"Server is not ready: {str(e)}")
        time.sleep(VALIDATE_PERIOD)


def save_servers_containers_logs(
    output_path: str, docker_compose: str, docker_client: Any
):
    os.makedirs(output_path, exist_ok=True)

    containers = docker_client.containers.list(
        all=True,
        filters={CONTAINER_FILTER_LABEL: f"{CONTAINER_FILTER_KEY}={docker_compose}"},
    )
    for container in containers:
        data = container.logs()
        output_file = os.path.join(output_path, f"{container.name}_logs.txt")
        with open(output_file, "wb") as f:
            f.write(data)
        log.info(f"[{container.name}] Logs saved to {output_file}")
