## Tools

#### Raw Logs

Logs are hosted via Google Cloud Storage:

 * TBA

#### Plots

Plots can be found under the following directories:

 * `blockprop_plots`: Scatter plots of the block propagation times of each
 block in a single test run.
 * `graphs`: Plots of average block propagation times and reorg rates
 across tests in a series.

#### syslogng parsing example

Build `parser` to parse Genesis syslog-ng logs for block propagation logs
by container.

    make
    ./parser -t 42a8799e-7881-4430-9dbb-ed304d1bd224 ef-test-42a879.log

See help of `parser` for more information.

    python calc_blockproptime.py 42a8799e-7881-4430-9dbb-ed304d1bd224/

#### cadvisor parsing example

Build cadvisor resource processor from source

    git clone git@github.com:whiteblock/pre-processors.git
    cd pre-processors
    make
    mv bin/resources ../  #move it to the same place as `parser`
    ./resources 42a8799e-7881-4430-9dbb-ed304d1bd224.stat -s rstat-42a8799e-7881-4430-9dbb-ed304d1bd224/
    python plot_resources.py rstats-42a8799e-7881-4430-9dbb-ed304d1bd224/
    # set var MAX_FILES to limit how many graphs to show

You can also run plot_resources.py on a single file:

    python plot_resources.py rstat-42a8799e-7881-4430-9dbb-ed304d1bd224/geth-service55

#### Geth Log Lines of Interest

From the miner:
```
INFO [08-18|21:44:49.399] Successfully sealed new block            number=3 sealhash="6e35ab…c2b939" hash="ad5095…de1e8b" elapsed=3.741s
```

From a receiving node:
```
INFO [08-18|21:44:49.979] Imported new chain segment               blocks=1 txs=177 mgas=42.088 elapsed=378.170ms   mgasps=111.293 number=3 hash="ad5095…de1e8b" dirty=1.81KiB
```