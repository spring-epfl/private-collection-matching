#!/bin/bash

IP=$(docker inspect --format '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' server)

#for n in 1 2 4 8 16 32 64 128 256 512 1024 2048 4096 8096
for n in 2 4 8 16 32 64 128 256 512 1024 2048
do
	echo "Running circuit PSI with $n sets..."
	docker exec server bash /psi/2PC-Circuit-PSI/benchmarks/bench-server.sh $n $IP 3 &
	docker exec client bash /psi/2PC-Circuit-PSI/benchmarks/bench-client.sh $n $IP 3 
	sleep 0.5
done
