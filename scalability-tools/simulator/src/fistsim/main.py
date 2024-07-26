import argparse
import logging
import os
import shutil
import signal
import sys
from ctypes import ArgumentError
from datetime import timedelta
from threading import Timer

import stats
import yaml
from simulate import Simulation

logging.basicConfig(format="%(asctime)s - %(levelname)s - %(message)s")

log = logging.getLogger("simulator")


WATCHER_PERIOD = 10


def handle_signal(signum=None, frame=None):
    """Executes specific actions on signal generation.
    The function only handles SIGINT and SIGTERM signals.

    Args:
        signum (_type_, optional): Identifier of a singal. Defaults to None.
        frame (_type_, optional): Ignored. Defaults to None.
    """
    if signum == signal.SIGINT:
        log.info("Detected user CTRL-C signal. Stopping simulation ...")
        simulation._stop_and_cleanup()
        sys.exit(2)
    elif signum == signal.SIGTERM:
        log.info("Detected timeout signal. Stopping simulation ...")
        simulation._stop_and_cleanup()
        sys.exit(3)
    else:
        log.info(f"Unhandled signal {signum}")


def convert_to_timedelta(s: str) -> timedelta:
    """Utility function to convert string like durations into python timedelta objects.

    Args:
        s (str): the string value to convert into timedelta. Expected values follow the syntax [NUMBER][s|m|h|d|w].

    Raises:
        ValueError: in case the unit of measure is not one of s, m, h, d, w.

    Returns:
        timedelta: the equivalent value to the input string.
    """
    units = {"s": "seconds", "m": "minutes", "h": "hours", "d": "days", "w": "weeks"}
    value = int(s[:-1])
    unit = units.get(s[-1].lower())
    if unit:
        return timedelta(**{unit: value})
    else:
        raise ValueError(f"Invalid time unit: {s[-1]} in string {s}")


def set_timeout_thread(timeout: str):
    """Create and starts a daemon thread that fires SIGTERM signal after a configurable timeout period.
    When the timer expires, the termination process of the simulator is triggered, regardless of the current state of a simulation.

    Args:
        timeout (str): a string value representing the timeout duration. The value must follow the syntax [NUMBER][s|m|h|d|w].
    """
    delay = convert_to_timedelta(timeout).total_seconds()

    def _inner():
        log.info("Simulation timer expired")
        os.kill(os.getpid(), signal.SIGTERM)

    log.info("Starting timeout thread ...")
    t = Timer(delay, _inner)
    t.daemon = True
    t.start()


def set_stats_thread(output_folder: str):
    """Create and starts a daemon thread that periodically saves the output of 'docker stats' into a file.
    The thread is periodically fired to capture docker containers statistics over time during the simulation.

    Args:
        output_folder (str): the folder where the csv file will be saved
    """
    stats.save_stats(output_folder)

    timer = Timer(
        WATCHER_PERIOD,
        set_stats_thread,
        kwargs={"output_folder": output_folder},
    )
    timer.daemon = True
    timer.start()


if __name__ == "__main__":
    # handle scripts input arguments
    parser = argparse.ArgumentParser(
        description="Scalability Test Tool on Fleet Management Platforms"
    )
    parser.add_argument(
        "--configFile", type=str, help="Path to configuration file in yaml format"
    )
    parser.add_argument(
        "--config", type=str, help="Configuration of the simulation in yaml format"
    )
    args = parser.parse_args()

    # load simulation configuration
    raw_data = None
    if args.config != None:
        raw_data = args.config
    elif args.configFile:
        config_file = os.path.abspath(args.configFile)
        config_dir = os.path.dirname(config_file)
        with open(config_file, "r") as file:
            raw_data = file.read()
    else:
        raise ArgumentError(
            "One of --config or --configFile arguments must be provided"
        )
    data = yaml.safe_load(raw_data)

    # register system signals handlers
    signal.signal(signal.SIGINT, handle_signal)  # user CTRL-C
    signal.signal(
        signal.SIGTERM, handle_signal
    )  # generic termination (here used for the timeout thread)

    # setup output folder
    log.info("###### Setup outputs directory ######")
    opath = data["output"]["path"]
    if os.path.exists(opath):
        shutil.rmtree(opath)
    os.makedirs(opath)

    # save config to output folder
    with open(os.path.join(opath, "config.yaml"), "w") as f:
        f.write(raw_data)

    # redirect logs to file
    logs_opath = os.path.join(opath, "simulator_logs.log")
    handler = logging.FileHandler(logs_opath)
    formatter = logging.Formatter(
        "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    handler.setFormatter(formatter)
    log.addHandler(handler)

    # setup daemon threads (will automatically stop when the main thread ends)
    log.info("###### Setup timeout and docker stats threads ######")
    if "timeout" in data and data["timeout"] is not None:
        set_timeout_thread(data["timeout"])
    if opath != None:
        set_stats_thread(opath)

    # execute simulation
    simulation = Simulation(data, opath)
    simulation.run()
