import json
import sys
import os
import matplotlib.pyplot as plt
import signal
import re
from operator import itemgetter
import math
from statistics import mean

# this script with chdir into the folder of parsed logs
SAVE_DIR = "../../blockprop_plots"


def signal_handler(fig, frame):
    print("Ctrl+c detected, exiting and closing all plots...")
    sys.exit(0)


class BlockPropagationParser:
    def __init__(self):
        # self.fig, self.ax = plt.subplots()
        self.blocks = {}
        self.num_nodes = 0
        self.fifty_pct_times = []
        self.hundred_pct_times = []
        self.reorgs = {}
        self.total_reorgs = 0
        self.total_final_blocks = 0

    def plot_block_prop_times(self, test_info):
        name = test_info["testName"]
        series, size = name.split("-")
        test_id = test_info["testID"]
        os.makedirs(SAVE_DIR, exist_ok=True)
        plt.figure()
        plt.title(f"100% Block Prop Times\n Series {series} - Block Size {size}B")
        plt.ylabel("Milliseconds")
        plt.xlabel("Block Index")
        plt.plot(self.hundred_pct_times, '.')
        plt.tight_layout()
        plt.savefig(f"{SAVE_DIR}/{name}_{test_id[:8]}_100pct_times.png")

        plt.figure()
        plt.title(f"51% Block Prop Times\n Series {series} - Block Size {size}B")
        plt.ylabel("Milliseconds")
        plt.xlabel("Block Index")
        # plt.ylim((0, 300))
        plt.plot(self.fifty_pct_times, '.')
        plt.tight_layout()
        plt.savefig(f"{SAVE_DIR}/{name}_{test_id[:8]}_51pct_times.png")

    def parse_file(self, file):
        # Assumes all log lines have a block hash value in its k-v pairs.
        with open(file, "r", errors='ignore') as f:
            reorg_cnt = 0
            for line in f:
                try:
                    logline = json.loads(line)
                except json.decoder.JSONDecodeError:
                    continue

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
                    self.total_final_blocks = max(logline["Values"]["number"],
                                                  self.total_final_blocks)
                if logline["message"].startswith("Successfully sealed new"):
                    self.blocks[hash]["mined"] = logline["unixNanoTime"] / 1e6
                if logline["message"].startswith("Chain reorg detected"):
                    # number of blocks added from reorg
                    # (logs don't indicate how many blocks dropped in reorg)
                    # reorg_cnt += int(logline["Values"]["add"])
                    reorg_cnt += 1

            self.reorgs[file] = reorg_cnt
            self.total_reorgs += reorg_cnt

    def sort_import_times(self):
        for hash, data in self.blocks.items():
            data["importTimes"].sort()

    def calc_prop_times(self):
        for hash, data in self.blocks.items():
            # the miner does not import the block, subtract 1
            if len(data["importTimes"]) == (self.num_nodes - 1):
                if data["mined"] == 0:
                    # print("missing block mining time, skipping data point...")
                    continue

                # 51% mark is half the number of nodes
                # The miner never prints out that it imports a segment, so we
                # subract 1
                half = math.ceil(self.num_nodes / 2) - 2
                fifty_pct = data["importTimes"][half] - data["mined"]

                # TODO: change variable names to 95pct. this was a last minute
                # change to better represent data.
                hundred_pct = data["importTimes"][self.num_nodes - 2] - data["mined"]

                self.blocks[hash]["fifty_pct"] = fifty_pct
                self.fifty_pct_times.append(fifty_pct)
                self.blocks[hash]["hundred_pct"] = hundred_pct
                self.hundred_pct_times.append(hundred_pct)

    def seen_by_majority(self):
        cnt = 0
        for hash, data in self.blocks.items():
            if len(data["importTimes"]) > math.ceil(self.num_nodes / 2):
                cnt += 1
        return cnt

    def truncate_init_stats(self):
        """Removes the first 10 block stats. Nodes generate Ethash DAG at
        different times which invalidates the first handful of stats.

        Note: python dicts are insertion ordered.
        """
        self.fifty_pct_times = self.fifty_pct_times[10:]
        self.hundred_pct_times = self.hundred_pct_times[10:]


# Attempts to detect failed nodes and compiles the best stats for a single test
def compile_stats(test_info):
    path = f'parsed_logs/{test_info["testName"]}_{test_info["testID"]}/'

    try:
        files = os.listdir(path)
    except FileNotFoundError:
        print(f"Error: logs for {path} missing, skipping...")
        return

    os.chdir(path)

    # key: number of nodes assumed alive, value: num blocks with valid stats
    valid_stats = {}

    for i in reversed(range(5)):
        parser = BlockPropagationParser()
        for f in files:
            geth_log = re.search("^geth-service[0-9]*$", f)
            if geth_log:
                try:
                    parser.parse_file(f)
                    parser.num_nodes += 1
                except FileNotFoundError:
                    # sometimes we are missing logs...
                    continue

        parser.num_nodes -= i
        parser.sort_import_times()
        parser.calc_prop_times()
        parser.truncate_init_stats()
        valid_stats[len(parser.fifty_pct_times)] = parser.num_nodes
        plt.close('all')
        del parser

    parser = BlockPropagationParser()

    for f in files:
        geth_log = re.search("^geth-service[0-9]*$", f)
        if geth_log:
            try:
                parser.parse_file(f)
            except FileNotFoundError:
                # sometimes we are missing logs...
                continue

    # use number of nodes with most valid block stats
    parser.num_nodes = valid_stats[max(valid_stats)]
    parser.sort_import_times()
    parser.calc_prop_times()
    parser.truncate_init_stats()
    parser.plot_block_prop_times(test_info)

    plt.close('all')
    if len(parser.fifty_pct_times) == 0:
        print(f"Error: no results for {path}. valid_stats: {valid_stats}. " +
              f"finalizedBlocks: {parser.total_final_blocks}")
        os.chdir("../..")
        return

    stat = {
        "testID": test_info["testID"],
        "testName": test_info["testName"],
        "totalBlocksSeen": len(parser.blocks),
        "nodesAliveInTest": parser.num_nodes,
        "blocksSeenByMajority": parser.seen_by_majority(),
        "blocksImportedByAll": len(parser.fifty_pct_times),
        "finalizedBlocks": parser.total_final_blocks,
        "avgFiftyPctTime": mean(parser.fifty_pct_times),
        "avgHundredPctTime": mean(parser.hundred_pct_times),
        "totalReorgs": parser.total_reorgs,
    }

    if len(parser.fifty_pct_times) < 120:
        print(f'Warning: test {test_info["testName"]} has below 120 block stats')
        print(f'valid_stats: {valid_stats}. finalizedBlocks: {parser.total_final_blocks}')

    os.chdir("../..")
    return stat


if __name__ == '__main__':
    master_log = []
    all_stats = []

    with open("master_autoexec.log", "r") as f:
        for line in f:
            master_log.append(json.loads(line))

    for test_info in master_log:
        stat = compile_stats(test_info)
        if stat:
            all_stats.append(stat)

    with open("compiled_stats.json", "w") as f:
        json.dump(all_stats, f, indent=4)
