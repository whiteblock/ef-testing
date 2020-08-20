import json
import sys
import os
import matplotlib.pyplot as plt
import signal
import numpy as np

MAX_FILES = 2

def signal_handler(fig, frame):
    print("Ctrl+c detected, exiting and closing all plots...")
    sys.exit(0)


class ResourceParser:
    def __init__(self):
        self.start = 0

    def plot_resources(self, resources):

        dt = np.diff(resources['time'])

        plt.figure()
        cpu = np.diff(resources['cpuSum']) / (3E9)  # 3 cores per node
        cpu_util = cpu / dt
        plt.plot(resources['time'][1:], cpu_util)
        plt.title("CPU Utilization")
        plt.ylabel("Utilization")
        plt.xlabel("Seconds")

        plt.figure()
        plt.plot(resources['time'], resources['memory'])
        plt.title("Memory Usage")
        plt.ylabel("Usage (MB)")
        plt.xlabel("Seconds")

        fig, ax = plt.subplots()
        ingressInst = np.diff(resources['net']['ingress']) / dt
        egressInst = np.diff(resources['net']['egress']) / dt
        plt.plot(resources['time'][1:], ingressInst, 'b')
        plt.plot(resources['time'][1:], egressInst, "orange")
        plt.title("Network Ingress and Egress")
        plt.ylabel("KB/s")
        plt.xlabel("Seconds")
        ax.legend(['Ingress', 'Egress'])

        plt.figure()
        plt.plot(resources['time'], resources['pkt']['drop'], '.')
        plt.plot(resources['time'], resources['pkt']['err'], 'r-')
        plt.title("Packet Drops and Errors")
        plt.ylabel("count")
        plt.xlabel("Seconds")

    def parse_file(self, file):
        time = []
        cpu = []
        mem = []
        egress = []
        ingress = []
        drop = []
        err = []

        last = 0
        # interp_factor = 0
        # cnt = 0
        with open(file, "r") as f:
            for line in f:

                stats = json.loads(line)
                if self.start == 0:
                    self.start = stats['unixNanoTime'] / 1E9

                if last == stats['unixNanoTime']:
                    continue

                last = stats['unixNanoTime']

                # if cnt % interp_factor != 0:
                #     cnt += 1
                #     continue

                time.append(stats['unixNanoTime'] / 1E9 - self.start)
                cpu.append(stats['cpuSum'])
                mem.append(stats['memory']['usage'] / 1E6)
                egress.append(stats['network']['eth1']['egress']['bytes'] / 1E3)
                ingress.append(stats['network']['eth1']['ingress']['bytes'] / 1E3)
                drop.append(stats['network']['eth1']['ingress']['dropped'] + stats['network']['eth1']['egress']['dropped'])
                err.append(stats['network']['eth1']['ingress']['errors'] + stats['network']['eth1']['egress']['errors'])
                # cnt += 1

        sorted_stats = sorted(zip(time, cpu, mem, egress, ingress, drop, err), key=lambda item: item[0])
        time, cpu, mem, egress, ingress, drop, err = zip(*sorted_stats)

        resources = {
            "time": time,
            "cpuSum": cpu,
            "memory": mem,
            "net": {
                "egress": egress,
                "ingress": ingress,
            },
            "pkt": {
                "drop": drop,
                "err": err
            }
        }

        return resources


if __name__ == '__main__':
    if len(sys.argv) == 2:
        path = sys.argv[1]
    else:
        print("usage: python plot_resource.py {file_or_dir}")
        exit()

    rPraser = ResourceParser()

    if os.path.isdir(path):
        files = os.listdir(path)
        os.chdir(path)
    else:
        files = [path]

    cnt = 0
    for f in files:
        geth_log = re.search("^geth-service[0-9]*$", f)
        if geth_log:
            resources = rPraser.parse_file(f)
            rPraser.plot_resources(resources)
            cnt += 1
            if cnt == MAX_FILES:
                break

    print("Hit Ctrl-c to close figures (you may need to click on a figure)")

    plt.tight_layout()
    plt.show()
