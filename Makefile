local:
	genesis local geth.yaml

stop:
	docker stop $$(docker ps -q)

teardown: 
	genesis local teardown

.PHONY: local stop teardown