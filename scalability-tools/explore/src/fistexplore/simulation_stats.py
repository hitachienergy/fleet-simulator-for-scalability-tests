import os
from datetime import datetime

import pandas as pd
from pandas import json_normalize
from yaml import SafeLoader, load


def get_all_files_from_root(root):
    root_path = os.path.abspath(root)
    output = []
    for f in os.listdir(root_path):
        f_path = os.path.join(root_path, f)

        if os.path.isfile(f_path):
            output.append(f_path)
        elif os.path.isdir(f_path):
            output.extend(get_all_files_from_root(f_path))
    return output


def get_all_config_files(all_files):
    config_files = [f for f in filter(lambda y: y.endswith(".yaml"), all_files)]

    yaml_dfs = []
    for path_to_yaml in config_files:
        yaml_df = dict()
        with open(path_to_yaml) as yaml_file:
            yaml_content = load(yaml_file, Loader=SafeLoader)
            yaml_df = json_normalize(yaml_content, sep="_")
        results_folder_path = os.path.dirname(path_to_yaml)
        yaml_df["path"] = results_folder_path
        yaml_dfs.append(yaml_df)
    return pd.concat(yaml_dfs)


def parse_device_stats_file(path):
    stats = dict()
    stats["file_exists"] = os.path.exists(path)

    if stats["file_exists"]:
        logs = []
        with open(path, "r") as file:
            logs = file.readlines()

        stats["total_devices"] = int(logs[0].split(" ")[-1])
        stats["success_devices"] = int(logs[1].split(" ")[-1])
        stats["start_at"] = datetime.fromtimestamp(float(logs[2].split(" ")[-1]))
        stats["duration"] = float(logs[3].split(" ")[-1][:-2])
        stats["avg_exec_time"] = float(logs[4].split(" ")[-1][:-2])
        stats["std_exec_time"] = float(logs[7].split(" ")[-1])

    return stats


def get_simulations_task_stats(configs):
    all_stats = []
    for _, row in configs[["path", "client_numberOfContainers"]].iterrows():
        devices_folders_path = os.path.join(row["path"], "devices")
        if not os.path.exists(devices_folders_path):
            continue
        found_devices_folders = os.listdir(devices_folders_path)
        missing_device_containers = (
            len(found_devices_folders) < row["client_numberOfContainers"]
        )
        for device_folder_path in found_devices_folders:
            stats_file = os.path.join(
                devices_folders_path, device_folder_path, "simulator_analysis.txt"
            )
            container_stats = parse_device_stats_file(stats_file)
            container_stats["path"] = row["path"]
            container_stats["container"] = device_folder_path
            container_stats["missing_device_container"] = missing_device_containers
            all_stats.append(container_stats)
    return pd.DataFrame(all_stats)


def get_completed_task_simulations(stats):
    t = stats.groupby("path")
    t = (t["file_exists"].all()) & (~t["missing_device_container"].any())

    completed = t[t].index.to_list()

    d = stats[stats.path.isin(completed)].copy()
    d["end_at"] = d["start_at"] + pd.to_timedelta(d["duration"], unit="s")

    # compact multi container simulations stats
    group_by = d.groupby("path")
    agg = group_by.agg(
        avg_device_exec_time=("avg_exec_time", "mean"),
        std_device_exec_time=("std_exec_time", "mean"),
        avg_container_exec_time=("duration", "mean"),
        std_container_exec_time=("duration", "std"),
        total_devices=("total_devices", "sum"),
        success_devices=("success_devices", "sum"),
    )
    agg["total_duration"] = group_by.end_at.max() - group_by.start_at.min()
    return d, agg.reset_index()


def get_incompleted_task_simulations(stats):
    t = stats.groupby("path")
    t = (t["file_exists"].all()) & (~t["missing_device_container"].any())
    completed = t[t].index.to_list()
    return stats[~stats.path.isin(completed)].copy()


def get_aggregate_stats(stats):
    return (
        stats.groupby(["name", "num_devices"]).end_at.max()
        - stats.groupby(["name", "num_devices"]).start_at.min()
    )
