import logging
from time import sleep

import requests

log = logging.getLogger("simulator")

DEFAULT_USERNAME = "tenant@thingsboard.org"
DEFAULT_PASSWORD = "tenant"

DEFAULT_FW_NAME = "simulation_test"
DEFAULT_FW_VERSION = "1.0"
DEFAULT_FW_CHECKSUM_AGL = "SHA256"


class Tenant:
    def __init__(self, baseurl: str) -> None:
        self.base_url = f"http://{baseurl}"

    def login(
        self, username: str = DEFAULT_USERNAME, password: str = DEFAULT_PASSWORD
    ) -> None:
        url = "{}/api/auth/login".format(self.base_url)
        headers = {"Content-Type": "application/json"}
        payload = {"username": username, "password": password}

        while True:
            try:
                response = requests.post(url, headers=headers, json=payload)
                if response.status_code == 200:
                    self.token = response.json()["token"]
                    return
                else:
                    break
            except Exception as e:
                log.error(f"Error: {str(e)}")
                sleep(10)
        raise Exception(
            "Login token generation failed with status code:", response.text
        )

    def register_device(
        self, device_name: str, access_token: str, device_profile_id: str
    ) -> str:
        url = "{}/api/device?accessToken={}".format(self.base_url, access_token)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        payload = {
            "name": device_name,
            "deviceProfileId": {
                "id": device_profile_id,
                "entityType": "DEVICE_PROFILE",
            },
            "additionalInfo": {},
        }

        response = requests.post(url, headers=headers, json=payload)
        if response.status_code == 200:
            log.debug("Device {} registered successfully.".format(device_name))
            return response.json()["id"]["id"]
        else:
            raise Exception("Fail to register device:", response.text)

    def update_firmware(self, firmware_file: str, firmware_id: str) -> None:
        url = "{}/api/otaPackage/{}".format(self.base_url, firmware_id)
        headers = {"X-Authorization": "Bearer " + self.token}
        with open(firmware_file, "rb") as f:
            files = {"file": (firmware_file, f, "application/octet-stream")}
            response = requests.post(
                url, headers=headers, files=files, data={"checksumAlgorithm": "SHA256"}
            )
            if response.status_code == 200:
                log.info("Firmware updated successfully")
            else:
                raise Exception("Failed to update firmware:", response.text)

    def create_firmware(
        self,
        device_profile_id: str,
        checksum: str,
        checksum_alg: str = DEFAULT_FW_CHECKSUM_AGL,
        title: str = DEFAULT_FW_NAME,
        version: str = DEFAULT_FW_VERSION,
    ) -> str:
        url = "{}/api/otaPackage".format(self.base_url)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        payload = {
            "deviceProfileId": {
                "id": device_profile_id,
                "entityType": "DEVICE_PROFILE",
            },
            "type": "FIRMWARE",
            "title": title,
            "version": version,
            "tag": "{}_{}".format(title, version),
            "checksumAlgorithm": checksum_alg,
            "checksum": checksum,
            "additionalInfo": {},
        }
        response = requests.post(url, headers=headers, json=payload)
        if response.status_code == 200:
            id = response.json()["id"]["id"]
            log.info("Firmware created successfully. Package id: {}".format(id))
            return id
        else:
            raise Exception("Failed to create firmware:", response.text)

    def launch_update(self, firmware_id: str, profile_id: str) -> None:
        profile = self.get_device_profile(profile_id)
        url = "{}/api/deviceProfile".format(self.base_url)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        profile["firmwareId"] = {"id": firmware_id, "entityType": "OTA_PACKAGE"}
        response = requests.post(url, headers=headers, json=profile)
        if response.status_code == 200:
            log.info("Update launches successfully")
        else:
            raise Exception("Failed to launch update: {}".format(response.text))

    def get_default_device_profile_id(self) -> str:
        url = "{}/api/deviceProfileInfo/default".format(self.base_url)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        response = requests.get(url, headers=headers)
        if response.status_code == 200:
            return response.json()["id"]["id"]
        else:
            raise Exception("Failed to get device profile: {}".format(response.text))

    def get_device_profile(self, id) -> dict:
        url = "{}/api/deviceProfile/{}".format(self.base_url, id)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        response = requests.get(url, headers=headers)
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception("Failed to get device profile: {}".format(response.text))

    def create_default_device_profile(self):
        profile_content = {
            "name": "simulator",
            "description": "Device profile for device simulators",
            "type": "DEFAULT",
            "transportType": "DEFAULT",
            "provisionType": "DISABLED",
            "profileData": {
                "configuration": {"type": "DEFAULT"},
                "transportConfiguration": {"type": "DEFAULT"},
                "provisionConfiguration": {"type": "DISABLED"},
            },
        }

        url = "{}/api/deviceProfile".format(self.base_url)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        response = requests.post(url, headers=headers, json=profile_content)
        if response.status_code != 200:
            raise Exception("Fail to create device profile: {}".format(response.text))

        id = response.json()["id"]["id"]
        self.configure_default_device_profile(id)
        return id

    def configure_default_device_profile(self, id: str):
        url = "{}/api/deviceProfile/{}/default".format(self.base_url, id)
        headers = {
            "Content-Type": "application/json",
            "X-Authorization": "Bearer " + self.token,
        }
        response = requests.post(url, headers=headers)
        if response.status_code != 200:
            raise Exception("Fail to create device profile: {}".format(response.text))
