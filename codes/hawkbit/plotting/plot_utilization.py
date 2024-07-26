import matplotlib.pyplot as plt
from matplotlib.ticker import FormatStrFormatter
import numpy as np
import argparse
import util

def plot(ipath, usage_filename, opath, title, save):
    start_time, end_time, device_num = util.read_client_logs(ipath, f'logs.txt')
    timestamps, cpu, mem, net = util.read_file_usage_stale(ipath, usage_filename, start_time, end_time)
    if (len(timestamps) == 0): return

    x_data = np.array(timestamps) - start_time
    cpu_data = np.array(cpu[util.CPU_PERCENT])
    mem_data = np.array(mem[util.MEM_PERCEMT])
    # cpu_data = np.mean(np.array(cpu[util.CPU_PERCENT]).reshape(-1, 5), axis=1)
    # mem_data = np.mean(np.array(mem[util.MEM_PERCEMT]).reshape(-1, 5), axis=1)

    # fig, ax = plt.subplots(nrows=7, ncols=1, gridspec_kw={'height_ratios': [3, 1, 1, 1, 1, 1, 1]}, figsize=(12, 24))
    fig = plt.figure(figsize=(24, 10))
    fig.subplots_adjust(wspace=0.3, hspace=0.3)

    ax1=plt.subplot2grid((6, 2), (0, 0), rowspan=6)
    # ax1 = ax[0]
    ax1.plot(cpu_data)
    ax1.set_ylabel('CPU Ultilization')
    ax1.set_yticks(np.arange(0, 600, 100))
    ax1.set_ylim([0, 550])
    ax1.tick_params(axis='y', labelcolor='tab:blue', labelsize=9)
    ax1.yaxis.set_major_formatter(FormatStrFormatter('%d%%'))
    ax1.set_xlim(0, end_time-start_time)
    ax1.set_xlabel("Time (s)")
    ax1.grid(linestyle= '--')

    ax1_1 = ax1.twinx()
    ax1_1.plot(x_data, mem_data, color="tab:orange")
    ax1_1.set_ylabel('Memory Ultilization')
    ax1_1.set_ylim([0, 4])
    ax1_1.set_yticks(np.arange(0, 4, 0.5)) # 0 - 100
    ax1_1.tick_params(axis='y', labelcolor='tab:orange', labelsize=9)
    ax1_1.yaxis.set_major_formatter(FormatStrFormatter('%d%%'))

    # ax1.set_title("CPU & Memory Utilization", loc="center")
    # notes = f'Total CPU Cores: {cpu[util.CPU_NUM]}\nMemory Limits: {mem[util.MEM_LIMIT]}GiB\nDevice Number: {device_num}'
    # ax1.text(0, -0.2, notes, fontsize=10, transform=plt.gca().transAxes)
    ax1.set_title(f'Total CPU Cores: {cpu[util.CPU_NUM]}\nMemory Limits: {mem[util.MEM_LIMIT]}GiB\nDevice Number: {device_num}', loc="left", fontsize=9)

    net_y_data = [util.NET_IN, util.NET_OUT, util.NET_IN_ERR, util.NET_OUT_ERR, util.NET_IN_DROP, util.NET_OUT_DROP]
    net_labels = ["Network input", "Network output", "Network input error", "Network output error", "Network input drop", "Network output drop"]
    net_ylims = [[0, None], [0, None], [-1, 1], [-1, 1], [-1, 1], [-1, 1]]
    net_yticks = [np.arange(0, 15, 6), np.arange(0, 30, 10), [0], [0], [0], [0]]
    net_color = ["tab:green", "tab:red", "tab:purple", "tab:brown", "tab:pink", "tab:cyan"]

    for i in range(0, 6):
        axi = plt.subplot2grid((6, 2), (i, 1))
        axi.plot(x_data, net[net_y_data[i]], color=net_color[i])
        axi.set_xlim(0, end_time-start_time)
        if i < 5:
            axi.xaxis.set_major_formatter(FormatStrFormatter(''))
        else:
            axi.set_xlabel("Time (s)")

        # if (net_yticks[i] != None).all():
        axi.set_yticks(net_yticks[i])
        
        axi.set_title(net_labels[i], fontsize=10)
        axi.yaxis.set_major_formatter(FormatStrFormatter('%dMB'))
        axi.set_ylim(net_ylims[i])
        axi.tick_params(axis='y', labelsize=9)
        axi.grid()

    # ax2.plot(x_data, net[util.NET_OUT], label="Network output bytes")
    # ax2.plot(x_data, net[util.NET_IN_ERR], linestyle='--', label="Network input error bytes")
    # ax2.plot(x_data, net[util.NET_OUT_ERR], linestyle=(0, (3, 1, 1, 1, 1, 1)), label="Network output error bytes", alpha=0.7)
    # ax2.plot(x_data, net[util.NET_IN_DROP], linestyle='dashdot', label="Network input drop bytes", alpha=0.7)
    # ax2.plot(x_data, net[util.NET_OUT_DROP], linestyle='dotted', label="Network output drop bytes", alpha=0.7)
    # ax2.set_title("Network I/O", loc="center")
    # ax2.legend(loc='lower left', bbox_to_anchor=(0, -0.4), fontsize=10)
    # ax2.set_xlim(0, end_time-start_time)
    # ax2.set_xlabel("Time (s)")
    # ax2.grid(linestyle= '--')
    
    plt.suptitle(f'{title}')

    if not save:
        plt.show()
    else:
        plt.savefig(f'{opath}/{title}.pdf', bbox_inches='tight')
        plt.clf()

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", "--input", help="Input data path", default="")
    parser.add_argument("-s", "--save", help="Save plot", action="store_true")
    parser.add_argument("-o", "--output", help="Output data path", default="")
    parser.add_argument("-b", "--base", help="Base data path", default="../data")
    args = parser.parse_args()

    ipath = args.base
    if len(args.input) > 0:
        ipath = args.input
    opath = args.base
    if len(args.input) > 0:
        opath = args.output

        
    plot(ipath, "server_usage.txt", opath, "Server Utilization", args.save)
    plot(ipath, "db_usage.txt", opath, "DB Utilization", args.save)
    plot(ipath, "mq_usage.txt", opath, "MQ Utilization", args.save)