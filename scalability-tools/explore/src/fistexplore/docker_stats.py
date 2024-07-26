import math

import numpy as np
import pandas as pd


def load_stats(file):
    stats = pd.read_csv(file)
    stats["datetime"] = pd.to_datetime(stats.datetime)
    return stats


units = {
    "b": 1,
    "kb": 10**3,
    "mb": 10**6,
    "gb": 10**9,
    "tb": 10**12,
    "kib": 2**10,
    "mib": 2**20,
    "gib": 2**30,
    "tib": 2**40,
}
inv_units = {k[::-1]: v for k, v in units.items()}


def string_to_byte(size):
    _size = size.replace(" ", "")
    i = math.inf
    inv_size = _size[::-1].lower()
    unit = None
    for inv_unit in inv_units.keys():
        new_i = len(_size) - len(inv_unit)
        if inv_size.startswith(inv_unit) and new_i < i:
            i = new_i
            unit = inv_unit[::-1]

    num = float(_size[:i])

    if not unit:
        print(f"Unhandled unit for value {size}")
        return np.nan

    return int(num * units[unit])


def get_container_stats(container, stats):
    c_stats = stats[stats.container == container].sort_values("datetime")
    c_stats = c_stats[
        ["datetime", "cpu_perc", "mem_usage", "net_i", "net_o", "block_i", "block_o"]
    ]
    c_stats["datetime"] = pd.to_datetime(c_stats["datetime"])
    c_stats["mem_usage"] = c_stats["mem_usage"].apply(string_to_byte)
    c_stats["net_i"] = c_stats["net_i"].apply(string_to_byte)
    c_stats["net_o"] = c_stats["net_o"].apply(string_to_byte)
    c_stats["block_i"] = c_stats["block_i"].apply(string_to_byte)
    c_stats["block_o"] = c_stats["block_o"].apply(string_to_byte)
    c_stats["net_o_diff"] = c_stats.net_o.diff()
    c_stats["net_i_diff"] = c_stats.net_i.diff()
    c_stats["block_o_diff"] = c_stats.block_o.diff()
    c_stats["block_i_diff"] = c_stats.block_i.diff()
    delta_time = c_stats["datetime"].diff().dt.total_seconds()
    c_stats["net_o_sec"] = (c_stats["net_o_diff"] / delta_time).fillna(0).astype(int)
    c_stats["net_i_sec"] = (c_stats["net_i_diff"] / delta_time).fillna(0).astype(int)
    c_stats["block_o_sec"] = (
        (c_stats["block_o_diff"] / delta_time).fillna(0).astype(int)
    )
    c_stats["block_i_sec"] = (
        (c_stats["block_i_diff"] / delta_time).fillna(0).astype(int)
    )
    return c_stats


def get_total_cpu_usage(stats):
    return (
        stats.pivot(index="datetime", values="cpu_perc", columns="container")
        .fillna(0)
        .sum(axis=1)
    )
