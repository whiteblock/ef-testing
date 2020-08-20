curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"eth_getBlockByNumber",
    "params":["0x1", true],
    "id":1
}' -H 'content-type:application/json;' 10.0.7.254:8545 | jq | tail -n 20

curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"eth_blockNumber",
    "params":[],
    "id":1
}' -H 'content-type:application/json;' 10.0.7.254:8545 | jq

exit

curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"eth_getBlockByNumber",
    "params":["0x3e", true],
    "id":1
}' -H 'content-type:application/json;' 0.0.0.0:8545 | jq | tail -n 20

curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"eth_blockNumber",
    "params":[],
    "id":1
}' -H 'content-type:application/json;' 0.0.0.0:8545 | jq


curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"net_peerCount",
    "params":[],
    "id":1
}' -H 'content-type:application/json;' 0.0.0.0:8545 | jq

curl -sX POST --data '{
    "jsonrpc":"2.0",
    "method":"admin_nodeInfo",
    "params":[],
    "id":1
}' -H 'content-type:application/json;' 0.0.0.0:8545 | jq