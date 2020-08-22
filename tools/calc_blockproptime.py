import json
import sys
import os
import matplotlib.pyplot as plt
import signal
import re
from operator import itemgetter
import math
from statistics import mean

def signal_handler(fig, frame):
    print("Ctrl+c detected, exiting and closing all plots...")
    sys.exit(0)


class BlockPropagationParser:
    def __init__(self):
        self.fig, self.ax = plt.subplots()
        self.blocks = {}
        self.num_nodes = 0
        self.fifty_pct_times = []
        self.hundred_pct_times = []

    def plot_block(self, utxo_set):
        # total_balance = 0
        # node, overall_tps = min(uParser.tps.items(), key=itemgetter(1))
        # node, max_tps = max(uParser.tps.items(), key=itemgetter(1))

        # for addr in utxo_set:
        #     plt.plot(utxo_set[addr]['time'], utxo_set[addr]['count'])
        #     plt.title("Total Balance of All Receive Addresses\n" +
        #               "(one line per recieve addr)\n" +
        #               "Network TPS: {}\n".format(overall_tps) +
        #               "Max Measured TPS: {}".format(max_tps))
        #     plt.ylabel("Balance (All Assets)")
        #     plt.xlabel("Seconds")
        #     total_balance += max(utxo_set[addr]['count'])
        # self.total_end_balances.append(total_balance)
        pass

    def parse_file(self, file):
        # Assumes all log lines have a block hash value in its k-v pairs.
        with open(file, "r") as f:
            for line in f:
                logline = json.loads(line)
                hash = logline["Values"]["hash"]
                if hash not in self.blocks:
                    self.blocks[hash] = {
                        "mined": 0,
                        "importTimes": []
                    }
                if logline["message"].startswith("Imported new chain segment"):
                    self.blocks[hash]["importTimes"].append(
                        int(logline["unixNanoTime"] / 1e6)
                    )
                if logline["message"].startswith("Successfully sealed new"):
                    self.blocks[hash]["mined"] = logline["unixNanoTime"] / 1e6

    def sort_import_times(self):
        for hash, data in self.blocks.items():
            data["importTimes"].sort()

    def calc_prop_times(self):
        for hash, data in self.blocks.items():
            if len(data["importTimes"]) >= (self.num_nodes - 1):
                if data["mined"] == 0:
                    print("missing block mining time, skipping...")
                    continue

                # 51% mark is half the number of nodes
                # The miner never prints out that it imports a segment, so we
                # subract 1
                half = math.ceil(self.num_nodes / 2) - 2
                try: 
                    fifty_pct = data["importTimes"][half] - data["mined"]
                except IndexError:
                    print(self.num_nodes)
                hundred_pct = data["importTimes"][self.num_nodes - 2] - data["mined"]

                self.blocks[hash]["fifty_pct"] = fifty_pct
                self.fifty_pct_times.append(fifty_pct)
                self.blocks[hash]["hundred_pct"] = hundred_pct
                self.hundred_pct_times.append(hundred_pct)
    def count_final_blocks(self):
        cnt = 0
        for hash, data in self.blocks.items():
            if len(data["importTimes"]) > 1:
                cnt += 1
        return cnt


if __name__ == '__main__':
    if len(sys.argv) == 2:
        path = sys.argv[1]
    else:
        print("usage: python calc_blockproptime.py {geth_json_log_dir}")
        exit()

    parser = BlockPropagationParser()
    parser.num_nodes = 30

    if os.path.isdir(path):
        files = os.listdir(path)
        os.chdir(path)
    else:
        print("Input arg must be a directory of pre-processed geth logs.")

    for f in files:
        geth_log = re.search("^geth-service[0-9]*$", f)
        if geth_log:
            try:
                blockPropTimes = parser.parse_file(f)
            except FileNotFoundError:
                # sometimes we are missing logs...
                continue

    parser.sort_import_times()
    parser.calc_prop_times()
    b = json.dumps(parser.blocks, indent=4, ensure_ascii=False)
    print(f"Blocks Considered Final: {parser.count_final_blocks()}")
    print(f"Test ID: {sys.argv[1]}")
    print(f"Total Blocks Seen: {len(parser.blocks)}")
    print(f"Blocks considered final: {parser.count_final_blocks()}")
    print(f"Blocks with Complete Stats: {len(parser.fifty_pct_times)}")
    print(f"51% block propagation time avg: {mean(parser.fifty_pct_times)} ms")
    print(f"100% block prop. time avg: {mean(parser.hundred_pct_times)} ms")

    print("Hit Ctrl-c to close figures (you may need to click on a figure)")
    # plt.tight_layout()
    # plt.show()
