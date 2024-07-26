import logging
from datetime import datetime

import requests
from requests.auth import HTTPBasicAuth

log = logging.getLogger("simulator")

DEFAULT_USERNAME = "admin"
DEFAULT_PASSWORD = "admin"

GATEWAY_TOKEN_KEY = "authentication.gatewaytoken.key"
GATEWAY_TOKEN_ENABLE_KEY = "authentication.gatewaytoken.enabled"
TARGET_TOKEN_ENABLE_KEY = "authentication.targettoken.enabled"
ANONYMOUS_DOWNLOAD_ENABLE_KEY = "anonymous.download.enabled"

DEFAULT_SW_MODULE_NAME = "simulation_sw_module"
DEFAULT_SW_MODULE_VERSION = "v1.0"

DEFAULT_DISTRIBUTION_SET_NAME = "simulation_distribution_set"
DEFAULT_DISTRIBUTION_SET_VERSION = "v1.0"
DEFAULT_DISTRIBUTION_SET_DESCRIPTION = "distribution set for simulation tests"


class Tenant:
    def __init__(
        self,
        baseurl: str,
        username: str = DEFAULT_USERNAME,
        password: str = DEFAULT_PASSWORD,
    ) -> None:
        self.base_url = baseurl
        self.auth = HTTPBasicAuth(username, password)

    def __sendRequest(self, req: requests.Request) -> requests.Response:
        req.auth = self.auth
        session = requests.Session()
        prepareReq = req.prepare()
        return session.send(prepareReq)

    def createSWModule(
        self,
        name: str = DEFAULT_SW_MODULE_NAME,
        version: str = DEFAULT_SW_MODULE_VERSION,
        isOS: bool = False,
        description: str = "",
        vendor="",
    ) -> list:
        data = [
            {
                "vendor": vendor,
                "name": name,
                "description": description,
                "type": "os" if isOS else "application",
                "version": version,
            }
        ]
        req = requests.Request(
            "POST", f"http://{self.base_url}/rest/v1/softwaremodules", json=data
        )
        req.headers["Content-Type"] = "application/hal+json"
        response = self.__sendRequest(req)
        if response.status_code == 201:
            log.info("Software module created successfully")
        else:
            raise Exception("Failed to create software module:", response.text)

        data = response.json()
        ids = []
        for module in data:
            ids.append(module["id"])
        return ids

    def uploadArtifacts(self, id: int, filePath: str):
        with open(filePath, "r") as f:
            files = {"file": f}
            req = requests.Request(
                "POST",
                f"http://{self.base_url}/rest/v1/softwaremodules/{id}/artifacts",
                files=files,
            )
            response = self.__sendRequest(req)
            if response.status_code == 201:
                log.info("Artifacts uploaded successfully")
            else:
                raise Exception("Failed to upload artifacts:", response.text)

    def createDistributionSet(
        self,
        modules: list,
        name: str = DEFAULT_DISTRIBUTION_SET_NAME,
        version: str = DEFAULT_DISTRIBUTION_SET_VERSION,
        description: str = DEFAULT_DISTRIBUTION_SET_DESCRIPTION,
    ) -> int:
        modules_data = []
        for module in modules:
            modules_data.append({"id": module})
        data = [
            {
                "requiredMigrationStep": False,
                "name": name,
                "description": description,
                "type": "app",
                "version": version,
                "modules": modules_data,
            }
        ]
        req = requests.Request(
            "POST", f"http://{self.base_url}/rest/v1/distributionsets", json=data
        )
        # req.headers["Content-Type"] = "application/json"
        response = self.__sendRequest(req)
        if response.status_code == 201:
            log.info("Distribution set created successfully")
            return response.json()[0]["id"]
        else:
            raise Exception("Failed to create distribution set:", response.text)

    def createFilterByName(self, name_prefix: str) -> int:
        data = {
            "query": f"name=={name_prefix}*",
            "name": name_prefix,
        }
        req = requests.Request(
            "POST", f"http://{self.base_url}/rest/v1/targetfilters", json=data
        )
        req.headers["Content-Type"] = "application/json"
        response = self.__sendRequest(req)
        if response.status_code == 201:
            log.info("Target filter created successfully")
            return response.json()["id"]
        else:
            raise Exception("Failed to create target filter:", response.text)

    def getRegisteredTargetsCount(self) -> int:
        req = requests.Request(
            "GET", f"http://{self.base_url}/rest/v1/targets?q=updatestatus==registered"
        )
        req.headers["Content-Type"] = "application/json"
        response = self.__sendRequest(req)
        if response.status_code == 200:
            return response.json()["total"]
        else:
            raise Exception("Failed to get registered targets count:", response.text)

    def createRollout(
        self,
        filter: str,
        dsID: int,
        amount: int = 1,
        successThreshold: int = 50,
        errorThreshold: int = 80,
    ) -> int:
        data = {
            "distributionSetId": dsID,
            "targetFilterQuery": filter,
            "description": "Rollout for all simulated targets",
            "amountGroups": amount,
            "type": "forced",
            "confirmationRequired": False,
            "name": "simulationRollout",
            # "forcetime" : 1689592726012,
            "successCondition": {
                "condition": "THRESHOLD",
                "expression": str(successThreshold),
            },
            "successAction": {"expression": "", "action": "NEXTGROUP"},
            "errorAction": {"expression": "", "action": "PAUSE"},
            "errorCondition": {
                "condition": "THRESHOLD",
                "expression": str(errorThreshold),
            },
            "startAt": int(round(datetime.now().timestamp())),  # start immediately
        }
        req = requests.Request(
            "POST", f"http://{self.base_url}/rest/v1/rollouts", json=data
        )
        req.headers["Content-Type"] = "application/hal+json"
        req.headers["Accept"] = "application/hal+json"
        response = self.__sendRequest(req)
        if response.status_code == 201:
            log.info("Rollout created successfully")
            return response.json()["id"]
        else:
            raise Exception("Failed to create rollout:", response.text)

    # def startRollout(self, id: int):
    #     req = requests.Request(
    #         "POST", f"http://{self.base_url}/rest/v1/rollouts/{id}/start"
    #     )
    #     req.headers["Accept"] = "application/hal+json"
    #     response = self.__sendRequest(req)
    #     if response.status_code == 200:
    #         log.info("Rollout starts successfully")
    #     else:
    #         raise Exception("Failed to start rollout:", response.text)

    def setDefaultConfig(self):
        resps = []
        for key in [
            TARGET_TOKEN_ENABLE_KEY,
            GATEWAY_TOKEN_ENABLE_KEY,
            ANONYMOUS_DOWNLOAD_ENABLE_KEY,
        ]:
            data = {"value": True}
            req = requests.Request(
                "PUT", f"http://{self.base_url}/rest/v1/system/configs/{key}", json=data
            )
            req.headers["Content-Type"] = "application/hal+json"
            resps.append(self.__sendRequest(req))
        for response in resps:
            if response.status_code != 200:
                raise Exception("Failed to setup default configuration:", response.text)
        log.info("Default configuarion set up successfully")

    def setGatewayToken(self, token: str):
        data = {"value": token}
        req = requests.Request(
            "PUT",
            f"http://{self.base_url}/rest/v1/system/configs/{GATEWAY_TOKEN_KEY}",
            json=data,
        )
        req.headers["Content-Type"] = "application/hal+json"
        response = self.__sendRequest(req)
        if response.status_code == 200:
            log.info("Gateway token configured successfully")
        else:
            raise Exception("Failed to configure gateway token:", response.text)
