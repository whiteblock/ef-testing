local:
	genesis local geth.yaml

prod:
	genesis run geth-prod.yaml

stop:
	docker stop $$(docker ps -q)

teardown: 
	genesis local teardown

lint:
	@for f in $(shell ls *.yaml); do echo -- $${f}; genesis lint $${f}; done

.PHONY: local stop teardown lint