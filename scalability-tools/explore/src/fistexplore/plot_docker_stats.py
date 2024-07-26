import matplotlib.ticker as ticker
import pandas as pd
from matplotlib import pyplot as plt

from . import docker_stats as stats_utils

filter_containers = ["c1-ping-1", "c1-asset-api-1", "c1-mongo-express-1", "docker-tc"]

y_label = {
    "cpu_perc": "Cpu usage (%)",
    "mem_usage": "Memory usage (bytes)",
    "net_i_sec": "Network upload (bytes/sec)",
    "net_o_sec": "Network download (bytes/sec)",
    "block_i_sec": "Disk read (bytes/sec)",
    "block_o_sec": "Disk write (bytes/sec)",
}


def plot_stats_comparison(
    stats,
):
    _, axs = plt.subplots(6, figsize=(10, 25))  # , sharex=True
    # num_containers = stats.container.unique().shape[0]
    # colors = plt.cm.brg(np.linspace(0,1,num_containers))
    for i, c_name in enumerate(sorted(stats.container.unique())):
        c_stats = stats_utils.get_container_stats(c_name, stats)
        if c_name not in filter_containers:
            axs[0].plot(
                c_stats["datetime"], c_stats["cpu_perc"], label=c_name
            )  # , color=colors[i]
            axs[1].plot(
                c_stats["datetime"], c_stats["mem_usage"], label=c_name
            )  # , color=colors[i]
            axs[2].plot(
                c_stats["datetime"], c_stats["net_i_sec"], label=c_name
            )  # , color=colors[i]
            axs[3].plot(
                c_stats["datetime"], c_stats["net_o_sec"], label=c_name
            )  # , color=colors[i]
            axs[4].plot(
                c_stats["datetime"], c_stats["block_i_sec"], label=c_name
            )  # , color=colors[i]
            axs[5].plot(
                c_stats["datetime"], c_stats["block_o_sec"], label=c_name
            )  # , color=colors[i]

        axs[0].legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs[0].grid(True)
        axs[0].set_xlabel("Time")
        axs[0].set_ylabel(y_label["cpu_perc"])

        axs[1].legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs[1].grid(True)
        axs[1].set_xlabel("Time")
        axs[1].set_ylabel(y_label["mem_usage"])
        axs[1].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )

        axs[2].legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs[2].grid(True)
        axs[2].set_xlabel("Time")
        axs[2].set_ylabel(y_label["net_i_sec"])
        axs[2].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )

        axs[3].legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs[3].grid(True)
        axs[3].set_xlabel("Time")
        axs[3].set_ylabel(y_label["net_o_sec"])
        axs[3].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )

        axs[4].legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs[4].grid(True)
        axs[4].set_xlabel("Time")
        axs[4].set_ylabel(y_label["block_i_sec"])
        axs[4].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )

        axs[5].legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs[5].grid(True)
        axs[5].set_xlabel("Time")
        axs[5].set_ylabel(y_label["block_o_sec"])
        axs[5].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )

    plt.show()


def plot_stats_comparison_v2(
    stats,
):
    mem_units = ["mem_usage", "net_i_sec", "net_o_sec", "block_i_sec", "block_o_sec"]
    # num_containers = stats.container.unique().shape[0]
    # colors = plt.cm.brg(np.linspace(0,1,num_containers))
    for stat_name in y_label.keys():
        _, axs = plt.subplots(figsize=(7, 4))
        for _, c_name in enumerate(sorted(stats.container.unique())):
            if c_name in filter_containers:
                continue
            c_stats = stats_utils.get_container_stats(c_name, stats)
            axs.plot(
                c_stats["datetime"], c_stats[stat_name], label=c_name
            )  # , color=colors[i]

        # axs.legend(bbox_to_anchor=(0.5, -0.2), loc="upper center", ncol=2)
        axs.legend(loc="center left", bbox_to_anchor=(1, 0.5))
        axs.grid(True)
        axs.set_xlabel("Time")
        axs.xaxis.set_tick_params(rotation=45)
        axs.set_ylabel(y_label[stat_name])
        if stat_name in mem_units:
            axs.yaxis.set_major_formatter(
                ticker.FuncFormatter(lambda x, _: human_format(x))
            )

        plt.show()


def plot_cpu_usage(temp, num_of_cores):
    _, axs = plt.subplots(figsize=(7, 4))
    axs.plot(temp.index, temp, label="Total cores usage")
    axs.axhline(y=num_of_cores * 100, color="r", linestyle="-", label="Max cores usage")
    axs.grid(True)
    axs.set_xlabel("Time")
    axs.set_ylabel("Cpu cores usage")
    axs.legend()
    plt.show()


def plot_logs_info(ax, logs):
    for _, row in logs.iterrows():
        vline_date = pd.to_datetime(row["datetime"])
        ax.axvline(x=vline_date, color="r", linestyle="--")
        ax.text(vline_date, ax.get_ylim()[1], row["message"], va="top", rotation=90)


def plot_df_values(ax, df, column):
    # ax.plot(df["datetime"], df[column], 'o')
    for i in range(len(df) - 1):
        ax.hlines(
            df[column].iloc[i], df["datetime"].iloc[i], df["datetime"].iloc[i + 1]
        )


def human_format(num):
    magnitude = 0
    while abs(num) >= 1000:
        magnitude += 1
        num /= 1000.0
    return "%.1f%s" % (num, ["B", "KB", "MB", "GB", "TB", "PB"][magnitude])


def plot_container_stats(
    stats,
    logs,
):
    for c_name in sorted(stats.container.unique()):
        c_stats = stats_utils.get_container_stats(c_name, stats)

        plt.title(f"Container {c_name}")
        _, axs = plt.subplots(4, figsize=(10, 10), sharex=True)
        plot_df_values(axs[0], c_stats, "cpu_perc")
        plot_logs_info(axs[0], logs)
        axs[0].set_xlabel("Datetime")
        axs[0].set_ylabel("Cpu usage (%)")
        axs[0].grid(True)

        plot_df_values(axs[1], c_stats, "mem_usage")
        plot_logs_info(axs[1], logs)
        axs[1].set_xlabel("Datetime")
        axs[1].set_ylabel("Mem usage")
        axs[1].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )
        axs[1].grid(True)

        plot_df_values(axs[2], c_stats, "net_i_sec")
        plot_logs_info(axs[2], logs)
        axs[2].set_xlabel("Datetime")
        axs[2].set_ylabel("Network Input (B/sec)")
        axs[2].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )
        axs[2].grid(True)

        plot_df_values(axs[3], c_stats, "net_o_sec")
        plot_logs_info(axs[3], logs)
        axs[3].set_xlabel("Datetime")
        axs[3].set_ylabel("Network Output (B/sec)")
        axs[3].yaxis.set_major_formatter(
            ticker.FuncFormatter(lambda x, _: human_format(x))
        )
        axs[3].grid(True)

        plt.show()
