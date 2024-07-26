import signal
import docker
import time
import sys

# Define the signal handler function
def handle_stop_signal(signal_num, frame):
    print("Received stop signal. Exiting...")
    exit(0)

# Register the signal handler for SIGINT (Ctrl+C) and SIGTERM (termination signal)
signal.signal(signal.SIGINT, handle_stop_signal)
signal.signal(signal.SIGTERM, handle_stop_signal)

def main(container_id):
    client = docker.from_env()
    container = client.containers.get(container_id)

    gib = 1073741824
    mb = 1024 * 1024

    stats = container.stats(stream=False)
    cpu_count = stats['cpu_stats']['online_cpus']
    print(f'Core Number: {cpu_count}\n')
    mem_limit = stats['memory_stats']['limit'] / gib
    print(f'Memorz limit: {mem_limit}GiB\n')
    network_stats = stats['networks']
    prev_network_in = sum(network_stats[n]['rx_bytes'] for n in network_stats)/mb
    prev_network_out = sum(network_stats[n]['tx_bytes'] for n in network_stats)/mb
    prev_network_in_err = sum(network_stats[n]['rx_errors'] for n in network_stats)
    prev_network_in_drop = sum(network_stats[n]['rx_dropped'] for n in network_stats)
    prev_network_out_err = sum(network_stats[n]['rx_errors'] for n in network_stats)
    prev_network_out_drop = sum(network_stats[n]['rx_dropped'] for n in network_stats)

    print("\nTimestamp CPU(%) Mem-Usage(GiB) Mem(%) Net-IO(MB) Net-Err Net-Drop")

    # Main loop
    last_syste_cpu_usage = stats['cpu_stats']['system_cpu_usage']
    last_cpu_usage = stats['cpu_stats']['cpu_usage']['total_usage']
    stats_stream = container.stats(stream=True, decode=True)
    for stats in stats_stream:
        cpu_total_usage = stats['cpu_stats']['cpu_usage']['total_usage'] - last_cpu_usage
        last_cpu_usage = stats['cpu_stats']['cpu_usage']['total_usage']
        system_cpu_usage = stats['cpu_stats']['system_cpu_usage'] - last_syste_cpu_usage
        last_syste_cpu_usage = stats['cpu_stats']['system_cpu_usage']
        cpu_percent =  (cpu_total_usage / system_cpu_usage) * cpu_count

        mem_usage = stats['memory_stats']['usage'] / gib
        mem_percent = mem_usage / mem_limit

        network_stats = stats['networks']
        network_in = sum(network_stats[n]['rx_bytes'] for n in network_stats)/mb
        network_in_delta = network_in-prev_network_in
        prev_network_in = network_in
        network_out = sum(network_stats[n]['tx_bytes'] for n in network_stats)/mb
        network_out_delta = network_out-prev_network_out
        prev_network_out = network_out
        network_in_err = sum(network_stats[n]['rx_errors'] for n in network_stats)/mb
        network_in_err_delta = network_in_err-prev_network_in_err
        prev_network_in_err = network_in_err
        network_in_drop = sum(network_stats[n]['rx_dropped'] for n in network_stats)/mb
        network_in_drop_delta = network_in_drop-prev_network_in_drop
        prev_network_in_drop = network_in_drop
        network_out_err = sum(network_stats[n]['rx_errors'] for n in network_stats)/mb
        network_out_err_delta = network_out_err-prev_network_out_err
        prev_network_out_err = network_out_err
        network_out_drop = sum(network_stats[n]['rx_dropped'] for n in network_stats)/mb
        network_out_drop_delta = network_out_drop-prev_network_out_drop
        prev_network_out_drop = network_out_drop

        print("{} {:.2%} {:.2f} {:.2%} {:.2f}/{:.2f} {:.2f}/{:.2f} {:.2f}/{:.2f}".format(stats["read"],
                                                                        cpu_percent, 
                                                                        mem_usage, 
                                                                        mem_percent, 
                                                                        network_in_delta, 
                                                                        network_out_delta, 
                                                                        network_in_err_delta, 
                                                                        network_out_err_delta, 
                                                                        network_in_drop_delta, 
                                                                        network_out_drop_delta))
time.sleep(1)

if __name__ == "__main__":
    # Execute the main function
    main(sys.argv[1])

