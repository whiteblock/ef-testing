curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"eth_getBlockByNumber",
    "params":["0x4a0", true],
    "id":1
}' -H 'content-type:application/json;' 10.0.7.254:8545 | jq

curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"eth_blockNumber",
    "params":[],
    "id":1
}' -H 'content-type:application/json;' 10.0.7.254:8545 | jq
