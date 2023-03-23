# Private Collection Matching Benchmarks

In the paper we solved two problems using our framework: chemical similarity testing and document search. The go package contained in this directory contains two benchmarking scripts, one for each problem. Internally these scripts rely on a partial implementation of our framework in Go. The implementations use the `Lattigo` library for access to the BFV somewhat homomorphic encryption scheme.

As explained in the paper, due to the limited multiplicative depth of the BFV encryption scheme, we could not implement all possible combinations of matching and aggregation functions supported by our framework.

## Installation and testing

We provide two options for running the benchmarks: (1) running all the code in a docker container, or (2) installing the code locally.

### Running in docker
We provide a docker container to run our system without worrying about dependencies.

```
$ cd GoPSI
$ docker build -t gopsi .
$ docker run -it --rm -v "$(pwd)"/../chemistry:/GoPSI/chemistry -v "$(pwd)"/../data:/GoPSI/data gopsi
```

This command mounts the `chemistry` and `data` directories inside the docker container so that they are available there. You can now run commands as usual. The binaries have already been build. For example:

```
(gopsi) $ cd /GoPSI/cmd/chem_search
(gopsi) $ ./chem_search
```

### Manual install
You can also install our system manually outside the docker.

#### Install dependencies
Our framework is written in Go. You can download and install go following the instruction [here](https://go.dev/dl/). After installing GoLang, the go command will take care of the remaining dependencies.

```
$ cd GoPSI
$ go get ./...
```

#### Build the executable
You can build the CLI executable as follows:
```
$ cd GoPSI/cmd/chem_search
$ go build
```
```
$ cd GoPSI/cmd/doc_search
$ go build
```
```
$ cd GoPSI/cmd/small_domain_bench
$ go build
```

### Testing

Our implementations come with a collection of tests. You can run these tests (either from the docker container or locally) by running:

```
$ cd GoPSI/pkg/psm
$ go test
```

Running all tests takes 5-10 minutes.

As long as you call `go test` directly from the `psm` directory this will work directly (either from the docker container or locally). Otherwise, please update the `const FPS_MINI_PATH` path variable in `config.go` to point to the absolute path to of the `data/raw_chem/fps-mini.txt` file.


## Benchmarking programs

We implemented three benchmarking programs to evaluate the framework in different configurations:  `cmd/chem_search/chem_search`, `cmd/doc_search/doc_search`, and `cmd/small_domain_bench/small_domain_bench`. Passing in the `-h` flag prints the help message.

The following table describes the configurations of the framework tested by the scripts.

| Programs             | PSI Layer              | Matching Layer         | Aggregation Layer      |
|----------------------|------------------------|------------------------|------------------------|
| `chem_search`        | Small Domain + ePSI-CA | Tversky Match (Tv-PSM) | Naive / X-Agg / CA-Agg |
| `doc_search`         | Small Input  + PSI     | Full Match (F-PSM)     | Naive / X-Agg / CA-Agg |
| `small_domain_bench` | Small Domain + ePSI-CA | No Matching            | Naive                  |

Whereas `chem_search` and `doc_search` evaluate the framework in full (for different configurations of the PSI and matching layers) the `small_domain_bench` script is used to benchmark the use of BFV for computing intersection cardinalities.

*Shared parameters.* The precise setting of the benchmark is controlled through command-line arguments. We recap here the shared parameters:

 * `-ns` The number of server sets (default 1024). For example: the number of chemical compounds for `chem_search` and the number of documents for `doc_search`.
 * `-logn` the BFV polynomial degree in bits (default 15, supported values 12--15). For example, for `-logn 13`, the program uses the `P_{8k}` configuration from the paper.
 * `-o file` The filename to which to write the JSON benchmarking results (default "bench.json")
 * `-r int` The number of times to repeat the experiment (default 1)
 * `-bar` If supplied, shows a progress bar
 * `-v` If supplied give verbose output.

The `chem_search` and `doc_search` programs additionally take the type of aggregation as an input:

 * `-agg string` Specifies the aggregation function used to compute the collection-wide response ['' (naive), 'x-ms', 'ca-ms'] (default "x-ms")

*Note.* For benchmarking purposes the randomness seed is fixed between runs. When using the framework directly (see below), you should use `rand.Seed()` to ensure the randomness is different for each run.

*Running full benchmarks.* Please see `../bench/` for the scripts that we used to run these individual benchmarking programs and produce the data in the paper.

### Benchmarking Chemical Similarities

The `chem_search` benchmarking program can be used to measure performance in the chemical similarity setting (see Section 11.1 in the paper). The evaluation in the paper (see Figure 5) contain performance results using existential aggregation (with `-agg x-ms`) and cardinality aggregation (with `-agg ca-ms`). In addition to the common parameters above, this benchmark program supports the following options:

 * `-chemdb-path PATH` specifies a fingerprint source file. The number of fingerprints in the file should be at least as big as the number of server sets. If omitted (or empty), the program will generate random compound fingerprints. This repository comes with a precomputed set of 8000 fingerprints in `data/raw_chem/fps-mini.txt`. If you want a larger (non-random) input, please see `chemistry`/ for how to compute it.
 * `-sd-domain-size int` specifies the size of the compound finger print (small domain size, default 256).

Here is an example run of 1 measurement (`-r 1`) with the server using 1024 (`-ns 1024`) real molecular fingerprints (`-chemdb-path ../../../data/raw_chem/fps-mini.txt`) and cardinality aggregation (`-agg ca-ms`):

```
# cd cmd/chem_search
# ./chem_search -r 1 -chemdb-path ../../../data/raw_chem/fps-mini.txt -ns 1024 -v -agg ca-ms

2023-03-21T18:16:59+01:00 INF Setting logger level to 'Trace'.
2023-03-21T18:16:59+01:00 INF Param:
Number of query replicates in the ciphertext: 1
Small domain => SdBitVecLen: 256
Small input  => Max client element per ctx: 16, max server size: 2048

Use chemicals loaded from a database
Running benchmark with 1024 sets at 2023-03-21 18:16:59.898709 +0100 CET m=+0.400127146.
2023-03-21T18:17:03+01:00 INF Create a small domain query.
2023-03-21T18:17:03+01:00 INF server: running small domain psi
2023-03-21T18:17:03+01:00 INF server: computing psi-ca
2023-03-21T18:17:08+01:00 INF server: running tversky.
2023-03-21T18:17:09+01:00 DBG Number of Tv ciphertexts: 8
2023-03-21T18:17:11+01:00 DBG Number of batched Tv ciphertexts: 1
2023-03-21T18:17:11+01:00 INF server: convert tversky scores to binary matching.
2023-03-21T18:17:39+01:00 INF server: running ca-ms aggregation
2023-03-21T18:17:41+01:00 INF client: evaluating the response
2023-03-21T18:17:41+01:00 INF client: aggregated response.

***************************************************
* Computation
* #server sets:               1024
* Random set gen:             247ns
* Query:                      45.214965ms
* Response:                   38.037593937s
* Evaluation:                 31.440625ms
* Query Marshal:              11.817456ms
* Resp Marshal:               22.190683ms
* Client total  =>  88.473046ms
* Server total  =>  38.05978462s
***************************************************
* Communication
* Query size:                 6144 KB
* Response size:              6144 KB
***************************************************
* Key generation
* Time:                       3.215719593s
* Public key size:            7680 KB
* Relin key size:             60 MB
* Rotate key size:            510 MB
***************************************************
Answer:  [29]
```

### Benchmarking Document Search
The `doc_search` benchmarking program can be used to measure performance in the document search setting (see Section 11.2 in the paper). The evaluation in the paper (see Figure 6) contain performance results using existential aggregation (with `-agg x-ms` and `-logn 15`) and cardinality aggregation (with `-agg ca-ms` and `-logn 13`). In addition to the common parameters above, this benchmark program supports the following options:

 * `-hash-per-kw int` The number of hash functions used for each keyword. Determines the false-positive rate. Must be a power of 2. (default 2)
 * `-max-doc int` Maximum number of keywords in a query. Must be a power of 2. (default 128)
 * `-max-q int` Maximum number of keywords in a query. Must be a power of 2. (default 8)

The program will generate random document and search keywords. These inputs do not influence the runtime. Moreover, for the `-hash-per-kw` option, we only multiply the maximum query size to simulate the cost and randomly choose all elements. In other words, we do not apply multiple hash functions on the same input to provide the functionality.

Here is an example run of 1 measurement (`-r 1`) with the server using 2048 (`-ns 2048`) documents, with 128 keywords per document (`-max-doc 128`) and 8 query keywords (`-max-q 8`), cardinality aggregation (`-agg ca-ms`) using P_{8k} as BFV parameters (`-logn 13`):

```
# cd cmd/doc_search
# ./doc_search -r 1 -ns 1024 -max-doc 128 -max-q 8 -agg ca-ms -logn 13

Running benchmark with 1024 sets at 2023-03-21 18:31:45.180055 +0100 CET m=+0.046021977.

***************************************************
* Computation
* #server sets:               1024
* Random set gen:             111ns
* Query:                      3.189136ms
* Response:                   11.362480717s
* Evaluation:                 1.607075ms
* Query Marshal:              140.564µs
* Resp Marshal:               130.681µs
* Client total  =>  4.936775ms
* Server total  =>  11.362611398s
***************************************************
* Communication
* Query size:                 384 KB
* Response size:              384 KB
***************************************************
* Key generation
* Time:                       145.902298ms
* Public key size:            512 KB
* Relin key size:             3 MB
* Rotate key size:            22 MB
***************************************************
Answer:  [0]
```


### Benchmarking Small Domain PSI Base Layer

In addition to the document and chemical compound search executables that leverage the full framework, we provide a small benchmark program to facilitate reproducing our small-domain PSI-CA measurements in Appendix D.1. You can find it in `small_domain_bench/small_domain_bench`. This tool takes as input:

 * `-sd-domain-size int` the size of the small domain

The program will create random inputs, but inputs do not influence the run-time.

Here is an example run of 1 measurement (`-r 1`) that computes the intersection cardinality of intersecting the client set with 2048 server sets (`-ns 2048`) assuming a small domain of size 256 items (`-sd-domain-size 256`) and using `P_{8k}` (`-logn 13`):

```
# cd cmd/small_domain_bench
# ./small_domain_bench -r 1 -ns 2048 -sd-domain-size 256 -logn 13 -v

***************************************************
* Computation
* #server sets:               2048
* Random set gen:             160ns
* Query:                      2.974611ms
* Response:                   2.644589505s
* Evaluation:                 1.633185ms
* Query Marshal:              211.798µs
* Resp Marshal:               164.552µs
* Client total  =>  4.819594ms
* Server total  =>  2.644754057s
***************************************************
* Communication
* Query size:                 384 KB
* Response size:              384 KB
***************************************************
* Key generation
* Time:                       140.290308ms
* Public key size:            512 KB
* Relin key size:             3 MB
* Rotate key size:            22 MB
***************************************************

<...Result Omitted...>
```

## Internal Implementation

Our benchmarking programs use our implementation of the PCM framework. This framework implementation is not general. Our implementation is heavily optimized for the use cases in the paper: document and chemical search. These optimization are not compatible with all possible layer configurations and layer combinations outside our two scenario may not work out of the box. Yet, if your scenario matches one of the scenarios in the paper, the implementation of the framework in `pkg/psm` can be used directly.

The following example shows how the framework can be used and highlights different options. First, we define BFV parameters (using `GetBFVParam` and setup parameters for the framework (using `NewPSIParams`). Then we initialize the framework for a specific operational point (using `NewQueryType`). Thereafter, client and server interact following the pattern in the paper. The client shares the key with the server (`cl.GetKey()`), and sends the query itself (using `cl.Query`). The server computes a response (using `sv.Respond`) which the client evaluates to obtain the answer (using `cl.EvalResponse`)

```go
func Example() {
    var serverSets [][]uint64
    var clientSet []uint64

    bfvParams := GetBFVParam(15)       // BFV params with N=2^15
    pp := NewPSIParams(bfvParams, 128) // A framework params with 128 bit security

    // Query options:
    // 1st positional parameters (small domain): false (use small input), true (use small domain)
    // psi layer: PSI_PS, PSI_CA
    // Matching layer: MATCHING_NONE, MATCHING_TVERSKY, MATCHING_TVERSKY_PLAIN, MATCHING_FPSM
    // Aggregation layer: AGGREGATION_NAIVE, AGGREGATION_X_MS, AGGREGATION_CA_MS
    // Check 'types.go' for more information.
    queryType, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY, AGGREGATION_NAIVE)
    if err != nil {
        panic(err)
    }

    // Setup phase
    cl := NewClient(pp)
    sv, err := NewServer(pp, serverSets)
    if err != nil {
        panic(err)
    }
    clKey := cl.GetKey()

    // Query
    query, err := cl.Query(clientSet, *queryType)
    if err != nil {
        panic(err)
    }
    resp, err := sv.Respond(query, clKey)
    if err != nil {
        panic(err)
    }
    ans := cl.EvalResponse(clientSet, query, resp)
    _ = ans
}
```
