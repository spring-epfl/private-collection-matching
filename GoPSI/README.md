# Private Collection Matching Benchmarks

In the paper we solved two problems using our framework: chemical similarity testing and document search. The go package contained in this directory contains two benchmarking scripts, one for each problem. Internally these scripts rely on a partial implementation of our framework in Go. The implementations use the `Lattigo` library for access to the BFV somewhat homomorphic encryption scheme.

As explained in the paper, due to the limited multiplicative depth of the BFV encryption scheme, we could not implement all possible combinations of matching and aggregation functions supported by our framework.


## Structure of the code
Our framework is contained in `GoPSI/pkg/psm`. We provide a docker to ease reproducing our result, but we also have a guide on how to manually install and use our code. We provide two CLIs to perform document and chemical compound search. Moreover, we provide an overview of our code's structure and API. We have scrips to automate benchmarking our code in `bench` directory.

## Running in docker
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


## Manual install
You can also install our system manually outside the docker.

### Install dependencies
Our framework is written in Go. You can download and install go following the instruction [here](https://go.dev/dl/). After installing GoLang, the go command will take care of the remaining dependencies.

```
$ cd GoPSI
$ go get ./...
```

### Build the executable
You can build the CLI executable as follows:
```
$ cd GoPSI/cmd/chem_search
$ go build
```
```
$ cd GoPSI/cmd/doc_search
$ go build
```

## Testing

Our implementations come with a collection of tests. You can run these tests (either from the docker container or locally) by running:

```
$ cd GoPSI/pkg/psm
$ go test
```

Running all tests takes 5-10 minutes.

As long as you call `go test` directly from the `psm` directory this will work directly. Otherwise, please update the `const FPS_MINI_PATH` path variable in `config.go` to point to the absolute path to of the `data/raw_chem/fps-mini.txt` file.


## Benchmarking scripts

We implemented three benchmarking scripts to evaluate the framework in different configurations:  `cmd/chem_search/chem_search`, `cmd/doc_search/doc_search`, and `cmd/small_domain_bench/small_domain_bench`. Passing in the `-h` flag prints the help message.

*Note.* For benchmarking purposes the randomness seed is fixed between runs. When using the framework directly (see below), you should use `rand.Seed()` to ensure the randomness is different for each run.

*Shared parameters.* 

### Running CLI
You can run the `GoPSI/cmd/chem_search/chem_search` and `GoPSI/cmd/doc_search/doc_search` executables to access our CLI. Setting the `-h` prints the manual to use the CLI.


Common arguments between document and chemical search:
```
Usage: 
  -logn int
    	BFV polynomial degree (default 15)
  -ns int
    	Number of server document. (default 1024)
  -agg string
    	Aggregation function used to compute the collection-wide response. ['' (naive), 'x-ms', 'ca-ms'] (default "x-ms")
  -o string
    	Address of json output (default "bench.json")
  -r int
    	Number of times repeating the experiment (default 1)
  -bar
    	Add progress bar
  -v	Verbose
```


Parameters only applying to document search (`./doc_search/doc_search`):
```
usage:
  -hash-per-kw int
    	Number of hash functions used for each keyword. Determines the false-positive rate. Must be a power of 2. (default 2)
  -max-doc int
    	Maximum number of keywords in a query. Must be a power of 2. (default 128)
  -max-q int
    	Maximum number of keywords in a query. Must be a power of 2. (default 8)
```


Parameters only applying to chemical search (`./doc_search/chem_search`):
```
Usage:
  -chemdb-path string
    	Address of a chemical fingerprint dataset. (if empty '', uses randomly generated compounds)
  -sd-domain-size int
    	Size of the compound fingerprint(Small domain size) (default 167)
```

Besides the document and chemical compound search executables, we provide a CLI to ease reproducing our small-domain PSI-CA measurements in Appendix D.1. You can find this CLI in `small_domain_bench/small_domain_bench` and takes the following parameters:
````
Usage:
  -sd-domain-size int
    	Small domain size
```



### Running an chemical search engine application
We provide the command to run a chemical search (matching: tversky) that returns the matching status of all compounds without aggregation. Here, compounds are generated randomly. You can check `bench/bench` for more examples.

```
$ cd cmd
$ ./chem-search -logn 15 -agg ""
```

## Structure
Our framework has a client/server architecture and has a triple flow of `client.Query()`, `server.Response()`, and `client.EvalResponse()`.
Our implementation is heavily optimized for the use cases in the paper: document and chemical search. These optimization are not compatible with all possible layer configurations and layer combinations outside our two scenario may not work out of the box. 
We show the basic structure to use our code below:


```
func Example() {
    var serverSets [][]uint64
    var clientSet []uint64

    bfvParams := GetBFVParam(15)       // BFV params with N=2^15
    pp := NewPSIParams(bfvParams, 128) // A framework params with 128 bit se

    // Query options: 
    // psi layer: PSI_PS, PSI_CA
    // Matching layer: MATCHING_NONE, MATCHING_TVERSKY, MATCHING_TVERSKY_PLAIN, MATCHING_FPSM
    // Aggregation layer: AGGREGATION_NONE, AGGREGATION_X_MS, AGGREGATION_CA_MS
    // Check 'types.go' for more information.
    queryType, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY, AGGREGATION_NONE)
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
