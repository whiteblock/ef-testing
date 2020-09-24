import os
import json
import matplotlib.pyplot as plt
from statistics import mean

BLOCK_SIZES = ["40kB", "80kB", "160kB", "640kB", "1280kB"]
SERIES = ["1", "2", "3", "4", "5", "6", "final"]
SAVE_DIR = "graphs"
SERIES_TITLES = {
    "1": "Control",
    "2": "Network Latency",
    "3": "Increased Network Latency",
    "4": "Packet Loss",
    "5": "Bandwidth",
    "6": "Topology",
    "final": "Mixture"
}

CASE_TITLES = {
    "1": ["Control"],
    "2": ["50 ms", "80 ms", "120 ms", "150 ms", "50ms latency"],
    "3": ["200 ms", "300 ms", "400 ms", "500 ms", "50ms latency"],
    "4": ["0.01%", "0.1%", "0.5%", "1%", "50ms latency"],
    "5": ["5 Mbps", "20 Mbps", "50 Mbps", "80 Mbps", "1000 Mbps", "50ms latency"],
    "6": ["BA = 2", "BA = 4", "BA = 6", "BA = 10", "50ms latency"],
    "final": ["Mixutre", "50ms latency"]
}


def plot_graph(graphs, title, key, ylabel, fname):
    os.makedirs(SAVE_DIR, exist_ok=True)
    for series in SERIES:
        fig, ax = plt.subplots()
        if series in ["1"]:
            cases = ["a"]
        elif series in ["5"]:
            cases = ["a", "b", "c", "d", "e"]
        elif series in ["final"]:
            cases = [""]
        else:
            cases = ["a", "b", "c", "d"]

        for c in cases:
            plt.plot(graphs[series + c][key])

        if series not in ["1"]:
            # add series 1 plot as  control
            plt.plot(graphs["2a"][key], '--')

        plt.xticks(range(len(BLOCK_SIZES)), BLOCK_SIZES)
        plt.ylabel(ylabel)
        plt.xlabel("Block Size")
        ax.legend(CASE_TITLES[series])
        plt.title(f"{title}\n Series {series} {SERIES_TITLES[series]}")
        plt.savefig(f"{SAVE_DIR}/series{series}_{fname}.png")


if __name__ == '__main__':
    graphs = {}

    with open("compiled_stats.json", "r") as f:
        compiled_stats = json.load(f)

    for test in compiled_stats:
        series, block_size = test["testName"].split("-")
        block_size = block_size + 'B'
        if series not in graphs:
            graphs[series] = {}

        if block_size not in graphs[series]:
            graphs[series][block_size] = {
                "avg_fifty": [],
                "avg_hundred": [],
                "reorgs_per_block": [],
            }

        graphs[series][block_size]["avg_fifty"].append(test["avgFiftyPctTime"])
        graphs[series][block_size]["avg_hundred"].append(test["avgHundredPctTime"])
        graphs[series][block_size]["reorgs_per_block"].append(test["totalReorgs"] / test["finalizedBlocks"])

    # organize data into lists for graphing
    max_50pct_diff = 0
    max_100pct_diff = 0
    for series, block_size in graphs.items():
        graphs[series] = {
            "fifty_pct_times": [],
            "hundred_pct_times": [],
            "reorgs_per_block": [],
        }
        for size in BLOCK_SIZES:
            graphs[series]["fifty_pct_times"].append(mean(block_size[size]["avg_fifty"]))
            graphs[series]["hundred_pct_times"].append(mean(block_size[size]["avg_hundred"]))
            graphs[series]["reorgs_per_block"].append(mean(block_size[size]["reorgs_per_block"]))

        max_50pct_diff = max(abs(graphs[series]["fifty_pct_times"][0] -  graphs[series]["fifty_pct_times"][1]), max_50pct_diff)
        max_100pct_diff = max(abs(graphs[series]["hundred_pct_times"][0] -  graphs[series]["hundred_pct_times"][1]), max_100pct_diff)

    print(f"Max 51\% difference between two test iterations: {max_50pct_diff}")
    print(f"Max 100\% difference between two test iterations: {max_100pct_diff}")

    plot_graph(
        graphs,
        "Avg. 51% Block Propagation Times",
        "fifty_pct_times",
        "Millesconds",
        "51pct_times"
    )
    plot_graph(
        graphs,
        "Avg. 95% Block Propagation Times",
        "hundred_pct_times",
        "Millesconds",
        "95pct_times"
    )
    plot_graph(
        graphs,
        "Reorg Events per Finalized Block",
        "reorgs_per_block",
        "Avg. Events per Block",
        "reorg_rates"
    )
