.PHONY: clean clean-images build start stop prepare delays test


clean:
	docker container rm $(shell docker ps --filter status=exited -q)

clean-images:
	docker image rm benchmarks_server benchmarks_client
	docker image prune -f

build:
	git submodule update --init --recursive
	docker-compose build

start:
	docker-compose up -d

stop:
	docker-compose down

prepare:
	docker exec server bash /psi/benchmarks/install.sh

delays:
	docker exec client tc qdisc add dev eth0 root netem delay 50ms rate 100Mbit
	docker exec server tc qdisc add dev eth0 root netem delay 50ms rate 100Mbit

test:
	docker exec client ping -c1 server
	docker exec server ping -c1 client
