import json
import sys
import os
import matplotlib.pyplot as plt
import signal
import re
from operator import itemgetter
import math
from statistics import mean

FIG_SAVE_DIR = "figures"


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

    def plot_block_prop_times(self):
        os.makedirs("figures", exist_ok=True)
        plt.figure()
        plt.title("100% Block Prop Times")
        plt.ylabel("Milliseconds")
        plt.xlabel("Block Index")
        plt.plot(parser.hundred_pct_times, '.')
        plt.tight_layout()
        plt.savefig(f"{FIG_SAVE_DIR}/hundred_pct_times.png")

        plt.figure()
        plt.title("51% Block Prop Times")
        plt.ylabel("Milliseconds")
        plt.xlabel("Block Index")
        # plt.ylim((0, 300))
        plt.plot(parser.fifty_pct_times, '.')
        plt.tight_layout()
        plt.savefig(f"{FIG_SAVE_DIR}/fifty_pct_times.png")

    def parse_file(self, file):
        # Assumes all log lines have a block hash value in its k-v pairs.
        with open(file, "r") as f:
            reorg_cnt = 0
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
                    print("missing block mining time, skipping...")
                    continue

                # 51% mark is half the number of nodes
                # The miner never prints out that it imports a segment, so we
                # subract 1
                half = math.ceil(self.num_nodes / 2) - 2
                fifty_pct = data["importTimes"][half] - data["mined"]
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


if __name__ == '__main__':
    if len(sys.argv) == 2:
        path = sys.argv[1]
    else:
        print("usage: python calc_blockproptime.py {geth_json_log_dir}")
        exit()

    parser = BlockPropagationParser()

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
                parser.num_nodes += 1
            except FileNotFoundError:
                # sometimes we are missing logs...
                continue

    actual_num_nodes = parser.num_nodes

    """Force number of nodes here in case a node fails in a test. Parsing will
    only take into account blocks if exactly `parser.num_nodes` nodes print
    that a block was imported."""
    parser.num_nodes = 90

    parser.sort_import_times()
    parser.calc_prop_times()
    parser.truncate_init_stats()
    parser.plot_block_prop_times()

    # Uncomment this to inspect the dictionary of blocks
    # b = json.dumps(parser.blocks, indent=4, ensure_ascii=False)
    # print(b)

    print(f"Test ID: {sys.argv[1]}")
    print(f"Nodes: {actual_num_nodes}")
    print(f"Total Blocks Seen: {len(parser.blocks)}")
    print(f"Blocks imported by over 50% nodes: {parser.seen_by_majority()}")
    print(f"Blocks stats with {parser.num_nodes} import times: {len(parser.fifty_pct_times)}")
    print(f"Finalized blocks: {parser.total_final_blocks}")
    print(f"51% block prop. time avg: {mean(parser.fifty_pct_times):.2f} ms")
    print(f"100% block prop. time avg: {mean(parser.hundred_pct_times):.2f} ms")
    print(f"reorgs total: {parser.total_reorgs}")
    print(f"avg reorgs per block: {parser.total_reorgs / parser.total_final_blocks}")


    # Uncomment to show the plots (they are also saved)
    # print("Hit Ctrl-c to close figures (you may need to click on a figure)")
    # plt.show()
