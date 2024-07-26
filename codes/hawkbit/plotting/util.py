from datetime import datetime, timedelta
import pytz

CPU_PERCENT = "cpu_percent"
CPU_NUM = "cpu_num"
MEM_USAGE = "mem_usage"
MEM_PERCEMT = "mem_percent"
MEM_LIMIT = "mem_limit"
NET_IN = "in"
NET_OUT = "out"
NET_IN_ERR = "in_err"
NET_OUT_ERR = "out_err"
NET_IN_DROP = "in_drop"
NET_OUT_DROP = "out_drop"


def read_file_usage(p, filename, starttime, endtime):
    timezone_offset = pytz.timezone('Europe/Zurich').utcoffset(datetime.now()).total_seconds()
    timestamps = []
    cpu = {
	    CPU_PERCENT : []
    }
    mem = {
		MEM_PERCEMT : [],
		MEM_USAGE : []
    }
    net = {
		NET_IN : [],
		NET_OUT : [],
		NET_IN_ERR : [],
		NET_OUT_ERR : [],
		NET_IN_DROP : [],
		NET_OUT_DROP : [],
    }
    header = True
    with open(p + "/" + filename, 'r') as f:
        for line in f:
            data = line.split()
            if "Core Number" in line:
                cpu[CPU_NUM] = int(data[-1])
            if "Memorz limit" in line:
                mem[MEM_LIMIT] = float(data[-1].strip("GiB"))

            if len(data) != 7:
                continue
            if header:
                header = False
                continue
            
            adjust_timestr = data[0][:-1][:25]
            if len(adjust_timestr) < 26:
                adjust_timestr = adjust_timestr.ljust(25, '0')
            curr_time = (datetime.strptime(adjust_timestr+data[0][-1], '%Y-%m-%dT%H:%M:%S.%fZ') + timedelta(seconds=timezone_offset)).timestamp() 
            if curr_time < starttime or curr_time > endtime:
                continue

            timestamps.append(curr_time)
            cpu[CPU_PERCENT].append(float(data[1].rstrip("%")))
            mem[MEM_USAGE].append(float(data[2]))
            mem[MEM_PERCEMT].append(float(data[3].rstrip("%")))

            net_data = data[4].split("/")
            net[NET_IN].append(float(net_data[0]))
            net[NET_OUT].append(float(net_data[1]))
            net_data = data[5].split("/")
            net[NET_IN_ERR].append(float(net_data[0]))
            net[NET_OUT_ERR].append(float(net_data[1]))
            net_data = data[6].split("/")
            net[NET_IN_DROP].append(float(net_data[0]))
            net[NET_OUT_DROP].append(float(net_data[1]))
            
    return timestamps, cpu, mem, net

def read_file_usage_stale(p, filename, starttime, endtime):
    timezone_offset = pytz.timezone('Europe/Zurich').utcoffset(datetime.now()).total_seconds()
    timestamps = []
    cpu = {
	    CPU_PERCENT : []
    }
    mem = {
		MEM_PERCEMT : [],
		MEM_USAGE : []
    }
    net = {
		NET_IN : [],
		NET_OUT : [],
		NET_IN_ERR : [],
		NET_OUT_ERR : [],
		NET_IN_DROP : [],
		NET_OUT_DROP : [],
    }
    header = True
    prev_in = 0
    prev_out = 0
    prev_in_err = 0
    prev_out_err = 0
    prev_in_drop = 0
    prev_out_drop = 0
    with open(p + "/" + filename, 'r') as f:
        for line in f:
            data = line.split()
            if "Core Number" in line:
                cpu[CPU_NUM] = int(data[-1])
            if "Memorz limit" in line:
                mem[MEM_LIMIT] = float(data[-1].strip("GiB"))

            if len(data) != 7:
                continue
            if header:
                header = False
                continue
            
            adjust_timestr = data[0][:-1][:25]
            if len(adjust_timestr) < 26:
                adjust_timestr = adjust_timestr.ljust(25, '0')
            curr_time = (datetime.strptime(adjust_timestr+data[0][-1], '%Y-%m-%dT%H:%M:%S.%fZ') + timedelta(seconds=timezone_offset)).timestamp() 
            
            net_data = data[4].split("/")
            curr_in = float(net_data[0])
            curr_out = float(net_data[1])
            net_data = data[5].split("/")
            curr_in_err = float(net_data[0])
            curr_out_err = float(net_data[1])
            net_data = data[6].split("/")
            curr_in_drop = float(net_data[0])
            curr_out_drop = float(net_data[1])
            if curr_time >= starttime and curr_time <= endtime:
                timestamps.append(curr_time)
                cpu[CPU_PERCENT].append(float(data[1].rstrip("%")))
                mem[MEM_USAGE].append(float(data[2]))
                mem[MEM_PERCEMT].append(float(data[3].rstrip("%")))

                
                net[NET_IN].append(curr_in-prev_in)
                net[NET_OUT].append(curr_out-prev_out)
                net[NET_IN_ERR].append(curr_in_err-prev_in_err)
                net[NET_OUT_ERR].append(curr_out_err-prev_out_err)
                net[NET_IN_DROP].append(curr_in_drop-prev_in_drop)
                net[NET_OUT_DROP].append(curr_out_drop-prev_out_drop)
            prev_in = curr_in
            prev_out = curr_out
            prev_in_err = curr_in_err
            prev_out_err = curr_out_err
            prev_in_drop = curr_in_drop
            prev_out_drop = curr_out_drop
            
    return timestamps, cpu, mem, net

def read_client_logs(p, filename):
    start_set = set()
    end_set = set()
    start_time = -1
    end_time = -1
    device_num = 0
    with open(p + "/" + filename, 'r') as f:
        for line in f:
            data = line.strip().split(" ")
            if len(data)<2:
                continue
            timestamp = ' '.join(data[0:2])
            if "Simulate download" in line:
                object = data[-1]
                if object in start_set:
                    continue
                start_set.add(object)
                device_num += 1
                if start_time < 0:
                    start_time = datetime.strptime(timestamp, "%Y-%m-%d %H:%M:%S.%f").timestamp()
            elif "complete" in line:
                object = data[-2]
                if object in end_set:
                    continue
                end_set.add(object)
                end_time = datetime.strptime(timestamp, "%Y-%m-%d %H:%M:%S.%f").timestamp()
    return start_time, end_time, device_num