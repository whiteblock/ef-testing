# Ethereum Foundation Testing

#### Useful info
 * 21,000 gas to send a simple Eth tx
 * Concurrency patterns for account nonce: https://ethereum.stackexchange.com/questions/39790/concurrency-patterns-for-account-nonce
 * Ethereum json-rpc list: https://eth.wiki/json-rpc/API
 * go-ethereum CLI options: https://github.com/ethereum/go-ethereum/wiki/Command-Line-Options
 * Decimal-Hex Calc: https://www.mathsisfun.com/binary-decimal-hexadecimal-converter.html
 * `curl_commands.sh` for json-rpc templates

#### Etherchain Avg Stats
 * 800,000 transactions per day (fluctuates a lot)
 * 6500 blocks a day
 * 37,000 bytes per block
 * Avg_bytes / (avg_txs / avg_blocks) = 37000 / (800,000 / 6500) ~ 300.625 bytes per tx
 * gas_per_byte = 74 (seen in local tests)
 * Avg_bytes_per_tx * gas_per_byte = 300 * 74 = 22,200 gas per tx
 * 187,500 + 21000 gas per transaction = 207500
 * 8m gas limit -> 38 tx / block time - try 76 tps
 * 16m gas limit -> 76 tx / block time - try 152 tps
 * 32m gas limit -> 152 tx / block time - try 300 tps for all tests to ensure saturation

#### Experimental Observations

These help us understand the actual outcomes of certain settings (such as
setting tx size). Observations from running *2 local nodes*.

 * 3200 byte txs turn into a tx of ~238152 gas
 * estimate of 74 gas per byte in a tx
 * ~236800 gas per tx 
    * Set tx gas-price larger than this to avoid "intrinsic gas too low"
 * 400m gas limit
 * 1st exp. single data point: 240354.2366 gas per tx
    * 3311.3078 bytes per tx
 * 2nd exp. single data point: 238141.1594 gas per tx
 * 3rd exp single data point: 239130.2273 gas per tx
    * 3331.90 bytes per tx

Sending ETH from a single address at each node. Set `--txpool.accountslots` 
to allow a lot of concurrent transactions in the mempool.

#### Block Size to Gas Limit Needed

 * 40KB Blocks: 2945743.0351 gas limit 
    * ~12.37 tx/block

 * 640KB Blocks: 47mil gas limit (try 52 mil)
    * Test results with 47mil gas limit on 2 nodes (`genesis local`): 

    {"difficulty":{"max":199886,"standardDeviation":20981.299647556945,"mean":162157.4734411085},"totalDifficulty":{"max":140428372,"standardDeviation":40515682.18970304,"mean":65171068.34411084},"gasLimit":{"max":47000000,"standardDeviation":10320.095611324547,"mean":46998445.1143187},"gasUsed":{"max":42100552,"standardDeviation":4268547.787777598,"mean":41653278.82678979},"blockTime":{"max":368,"standardDeviation":12.652347024741374,"mean":2.3761574074074066},"blockSize":{"max":586956,"standardDeviation":59402.61354278756,"mean":580195.2840646649},"transactionPerBlock":{"max":177,"standardDeviation":17.95010692926018,"mean":175.16050808314077},"uncleCount":{"max":1,"standardDeviation":0.19693215894518937,"mean":0.040415704387990775},"tps":{"max":177,"standardDeviation":56.81412697238439,"mean":136.80314743151735},"blocks":866}





