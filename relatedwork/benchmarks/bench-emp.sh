#!/bin/bash

IP=$(docker inspect --format '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' server)

echo 'Dependencies installed.'

for n in 2 4 8 16 32 64 128 256 512 1024 2048
do 
    echo "Start cardinality search (CA-Agg) with ${n} sets..."
    docker exec server bash /psi/emp/bench-server-ca.sh $n $IP 3 &
    docker exec client bash /psi/emp/bench-client-ca.sh $n $IP 3
    sleep 1
    echo "Start existential search (X-Agg) with ${n} sets..."
    docker exec server bash /psi/emp/bench-server-x.sh $n $IP 3 &
    docker exec client bash /psi/emp/bench-client-x.sh $n $IP 3
done

echo 'All done!'
