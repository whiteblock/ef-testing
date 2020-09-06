local:
	genesis local geth.yaml

prod:
	genesis run test-yaml/6c-40k.yaml --json paccode

dev:
	genesis run geth-prod.yaml --dev --json paccode

stop:
	docker stop $$(docker ps -q)

teardown: 
	genesis local teardown

lint:
	@for f in $(shell ls test-yaml/*.yaml); do echo -- $${f}; genesis lint $${f}; done

pull:
	docker pull gcr.io/whiteblock/helpers/ethereum/accounts:master
	docker pull gcr.io/whiteblock/helpers/ethereum/genesis:master
	docker pull gcr.io/whiteblock/helpers/geth/keystore:master
	docker pull gcr.io/whiteblock/helpers/ethereum/static-peers:master
	docker pull gcr.io/whiteblock/helpers/ethereum/topology:master
	docker pull gcr.io/whiteblock/helpers/ethereum/tx:ef
	docker pull gcr.io/whiteblock/helpers/ethereum/await-blocks:master
	docker pull gcr.io/whiteblock/helpers/ethereum/record:master
	docker pull gcr.io/whiteblock/helpers/ethereum/tps-logger:master
	docker pull gcr.io/whiteblock/helpers/ethereum/block-logger:master
	docker pull gcr.io/whiteblock/helpers/ethereum/viewer:latest


.PHONY: local stop teardown lint pull