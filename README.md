# ef-testing
Ethereum Foundation Testing


Rough calculations to estimate transactions per second. Numbers do not need to be accurate, we just need a ballpark figure for sufficient tx rate and transaction size to saturate the network.

 * 21,000 gas to send a simple eth tx
 * 32 bytes is 20k gas
 * 2/s

Etherchain Stats:
 * 800,000 transactions per day (fluctuates a lot)
 * 6500 blocks a day
 * Rough average of 37,000 bytes per block
 * Avg_bytes / (avg_txs / avg_blocks) = 37000 / (800,000 / 6500) ~ 300.625 bytes per tx
 * Avg_bytes_per_tx * gas_per_byte = 300 * 20000/32 = 187,500 gas per tx
 * 187,500 + 21000 gas per transaction = 207500
 * 8m gas limit -> 38 tx / block time - try 76 tps
 * 16m gas limit -> 76 tx / block time - try 152 tps
 * 32m gas limit -> 152 tx / block time - try 300 tps for all tests to ensure saturation


target: 640 KB block size

 * 3200 bytes tx turns into a tx of 238152 gas
 * estimate of 74 gas per byte in a tx
 * 2e6 gas per tx
 * 400m gas limit
 * single data point: 240354.2366 gas per transaction

Sending ETH from a single address at each node. Set `--txpool.accountslots` 
accordingly.