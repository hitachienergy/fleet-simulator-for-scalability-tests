from abc import ABC, abstractmethod


class ServerHandler(ABC):
    @abstractmethod
    def start_server(self):
        pass

    @abstractmethod
    def trigger(self):
        pass

    @abstractmethod
    def store_and_cleanup(self, output_path, docker_client):
        pass
