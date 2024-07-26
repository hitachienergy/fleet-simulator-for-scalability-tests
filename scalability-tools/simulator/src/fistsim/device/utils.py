import logging
import subprocess
from typing import Any, Union

log = logging.getLogger("simulator")


def run_script(path: str, cwd: Union[str, None] = None, env: Any = None):
    res = subprocess.run(path, cwd=cwd, env=env)
    if res.returncode != 0:
        raise Exception(f"Subprocess returned with code {res.returncode}")
    return
