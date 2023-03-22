# Benchmarks for Related Work

## Overview
We benchmarked the following two PSI implementations and compared them to our framework:  (1) `emp` is a generic solution built on top of an SMC compiler called [EMP-toolkit](https://github.com/emp-toolkit/emp-tool) that we designed in Section 11.3 of our paper. `2PC` is a Circuit-PSI protocol based on [1] (the code is taken from the repository accompanying the paper).

We used a Docker infrastructure to have an identical user space for building, and running the implementations to benchmark them.

**Note:** The directory in which you find this README will be mounted in the Docker containers to make the related work code accessible from the Docker containers.

The benchmarks consist of a preparation phase in which:
- we build the Docker images that will be used,
- we start 2 Docker containers (one for the client, the other for the server),
- we build the related works inside one of the containers (and it will thus be immediately accessible in the other container),
- we simulate network delays,
- and optionally, we test that the containers can "speak" to each other with a 100 ms delay.

And finally, we run a benchmark script for each related work, which will run the correct binaries with the correct option in the Docker containers.

For ease of use, we also provided a Makefile which can be used to orchestrate operations.

### Implementations of related work

*EMP-toolkit.*
We used the EMP compiler to create an SMC protocol for two circuits implementing document search. Both circuits take as input the client set as well as all server sets, and determine, for each server set, whether the client set is contained within it. In terms of our framework, this corresponds to a PSI base-layer and F-Match matching layer. The SMC-CA-Agg version then applies cardinality aggregation (e.g., counting the number of full matches), whereas the SMC-X-Agg version applies existential aggergation (e.g., determining if one of the server sets is a full match).

We implemented both versions in EMP in the [EMP Companion repository](https://github.com/spring-epfl/emp-sh2pc/tree/master/test). In particular, the [`ms-ca.cpp`](https://github.com/spring-epfl/emp-sh2pc/blob/master/test/ms_ca.cpp) file defines the SMC-CA-Agg version, whereas the [`ms-x.cpp`](https://github.com/spring-epfl/emp-sh2pc/blob/master/test/ms_x.cpp) file defines the SMC-X-Agg version. To support measuring transfer cost, we also [modified the EMP Toolkit](https://github.com/spring-epfl/emp-tool/) to output the communciation cost.

*2PC Circuit PSI.*
To run proper benchmarks of the Circuit-PSI implementation [1] we had to make changes to the original source code to enable access to more precise benchmark measurements and control over the input. The library with modified code can be found [in the companion 2PC-Circuit-PSI repositiory](https://github.com/spring-epfl/2PC-Circuit-PSI/).

*SpOT-PSI.*
While we were able to run and evaluate SpOT [2] code on one desktop, we were unable to reliably reproduce the result in our containers. We had issues with the compatibility of the code, boost library, and compiler configuration which prevented us from building the executable in the Docker container.
If you are interested in evaluating this work, please follow the instruction in the original [code repository](https://github.com/osu-crypto/SpOT-PSI).

## Prerequisites
You will need Docker and the docker-compose script on your machine to run these benchmarks, as well as a few GB of storage to store the Docker images, the code of the related works and the collected data.

For ease of use, we provided a Makefile to orchestrate operations, if you do not want to, or can not install Make on your host, we also provide the commands run by Make for each of these steps at the end of this README.

**Note:** If you have a very recent version of Docker, the docker-compose script might already be integrated into the `docker` command as a plugin, in which case you can either modify the Makefile to replace instances of `docker-compose` by `docker compose`, or run the command manually without using Make.

## Preparation Phase
Our benchmarking infrastructure is located in the `benchmarks` directory.

**Important:** We use git submodules for handling related work codes. If you have not cloned the repository in a recursive mode, run the following command to initiate and retrieve third-party repositories:

```
$ git submodule update --init --recursive
```

You need to build the Docker images and start the containers from the `benchmark` working directory.
The containers are based on a Debian 11 "bullseye" image and contain all the dependencies to run the benchmarks, as well as a compilation toolchain to build the related works.
Moreover, the `relatedwork` directory of your host file system will be mounted in the Docker containers to make the related work code accessible from them.

```
$ cd benchmarks
$ make build
$ make start
```

**Note:** The `relatedwork` directory is mounted as `/psi` in Docker containers.

Once the containers are running, we need to build the related works.
We are doing these compilations inside one of the Docker containers where the compilation toolchain and dependencies are already installed.
We automated these tasks with scripts, which can be run inside the containers via a Make command.
```
$ make prepare
```

Finally, you will need to simulate network delays and test that you indeed have a network delay between the 2 containers.
```
$ make delays
(optional) make test
```

The `make test` command runs `ping` inside the two containers. You need to manually check if the round trip time is 100 ms.

The name of the Docker containers used in these benchmarks are `server` and `client`.

## Running the Benchmarks
We provide you with 2 scripts to benchmark 2PC-Circuit-PSI and emp-sh2pc.
Once the Docker infrastructure is running and the related work is compiled, you can run these scripts to run the benchmarks.

**Warning:** These scripts will take a long time to complain when they run over the full set of parameters. If you are just testing these scripts, or do not want to reproduce the full measurements, please modify the ranges in the respective scripts to the server set sizes that you want to test.

For 2PC-Circuit-PSI you can reproduce our measurements by running:
```
bash bench-2pc.sh
```
The results are written to `../2-PC-Circuit-PSI/benchmarks/` where `circuit-time-client.log` and `circuit-time-server.log` contain the respective runtime for the client and the server. These CSV files encode the number of server sets, the total time elapsed (s), the total user time (s), and the total kernel time (s).

For EMP you can reproduce our measurements by running:
```
bash bench-emp.sh
```
The runtimes (and internal measurements) are written stored in `../emp/` in the files `emp_ca_client.log`, `emp_ca_server.log`, `emp_x_server.log` and `emp_x_client.log`.


## Stopping the Experiment
At the end of the experiment, you can stop the Docker containers used for the benchmarks with a Make command
```
make stop
```

If some remaining Docker containers were not removed correctly (i.e. if you shut down your laptop while the containers were running), you can remove them by running a Make command to remove all stopped containers:
```
make clean
```


## Running the Benchmarks without Make

If you do not want to use or can not use Make on your host to run the benchmarks, here are the commands that you can run directly with their corresponding make command equivalent.
```
cd benchmarks

# make build
docker-compose build

# make start
docker-compose up -d

# make prepare
docker exec server bash /psi/benchmarks/install.sh

# make delays
docker exec client tc qdisc add dev eth0 root netem delay 50ms rate 100Mbit
docker exec server tc qdisc add dev eth0 root netem delay 50ms rate 100Mbit

# make test
docker exec client ping -c1 server
docker exec server ping -c1 client

# make stop
docker-compose down

# make clean
docker rm $(docker ps --filter status=exited -q)
```

## Refrences
[1] Nishanth Chandran, Divya Gupta, and Akash Shah. 2022. Circuit-PSI With Linear Complexity via Relaxed Batch OPPRF. PoPETs (2022).

[2] Benny Pinkas, Mike Rosulek, Ni Trieu, and Avishay Yanai. 2019. SpOT-Light: Lightweight Private Set Intersection from Sparse OT Extension. In CRYPTO.
