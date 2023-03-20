package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	. "github.com/spring-epfl/private-collection-matching/pkg/psm"
)

func RunCLIBench(cli_type string) {
	// Params

	set_name := "set"
	if cli_type == "document" {
		set_name = "document"
	} else if cli_type == "chemical" {
		set_name = "chemical compound"
	}
	nsPtr := flag.Int("ns", 1024, "Number of server "+set_name+".")
	lognPtr := flag.Int("logn", 15, "BFV polynomial degree")

	aggregationPtr := flag.String("agg", "x-ms", "Aggregation function used to compute the collection-wide response. ['' (naive), 'x-ms', 'ca-ms']")

	// Manage
	repPtr := flag.Int("r", 1, "Number of times repeating the experiment")
	progressBarPtr := flag.Bool("bar", false, "Add progress bar")
	verbosePtr := flag.Bool("v", false, "Verbose")
	outAddrPtr := flag.String("o", "bench.json", "Address of json output")

	var sdSize, maxDocQuerySize, maxDocSize, hashPerKw int
	var chembl string

	// chemical
	if cli_type == "chemical" {
		flag.IntVar(&sdSize, "sd-domain-size", 256, "Size of the compound fingerprint. Must be a power of 2.") // The size of MACCS keys is 167
		flag.StringVar(&chembl, "chemdb-path", "", "Address of a chemical fingerprint dataset. (if empty '', uses randomly generated compounds)")

	}
	// document
	if cli_type == "document" {
		flag.IntVar(&maxDocQuerySize, "max-q", 8, "Maximum number of keywords in a query. Must be a power of 2.")
		flag.IntVar(&maxDocSize, "max-doc", 128, "Maximum number of keywords in a query. Must be a power of 2.")
		flag.IntVar(&hashPerKw, "hash-per-kw", 2, "Number of hash functions used for each keyword. Determines the false-positive rate. Must be a power of 2.") // we simulate the cost of having multiple hash per keyword by increasing the set sizes without doing the hash management
	}
	// sd-comparison
	if cli_type == "sd-comparison" {
		flag.IntVar(&sdSize, "sd-domain-size", 256, "Size of the compound fingerprint. Must be a power of 2.")
	}

	flag.Parse()

	ENABLE_PROGRESS_BAR = *progressBarPtr
	if *verbosePtr {
		Logger = BuildLogger(zerolog.TraceLevel)
		Logger.Info().Msgf("Setting logger level to 'Trace'.")
	} else {
		Logger = BuildLogger(zerolog.ErrorLevel)
		Logger.Info().Msgf("Setting logger level to 'Error'.")
	}

	bfvParams := GetBFVParam(*lognPtr)
	pp := NewPSIParams(bfvParams, 128)
	if cli_type == "chemical" {
		if sdSize > 256 {
			if (sdSize & (sdSize - 1)) != 0 {
				panic("The small domain size (sd-domain-size) must be a power of 2.")
			}
			pp.SdBitVecLen = sdSize
		}
	} else if cli_type == "document" {
		pp.MaxClientElemPerCtx = maxDocQuerySize * hashPerKw
		pp.ClRepNum = int(bfvParams.N()) / pp.MaxClientElemPerCtx / maxDocSize
	}
	pp.Update()
	Logger.Info().Msgf("Param:\n%v\n", pp.Describe())

	// Build the query
	aggregation, agg_ok := ParseAggregationString(aggregationPtr)
	if !agg_ok {
		panic(errors.New("unknown aggregation type"))
	}
	var qt *QueryType
	var err error
	if cli_type == "chemical" {
		qt, err = NewQueryType(true, PSI_CA, MATCHING_TVERSKY, aggregation)
	} else if cli_type == "document" {
		qt, err = NewQueryType(false, PSI_PSI, MATCHING_FPSM, aggregation)
	} else if cli_type == "sd-comparison" {
		qt, err = NewQueryType(true, PSI_CA, MATCHING_NONE, AGGREGATION_NONE)
	}

	if err != nil {
		panic(err)
	}

	data := make([]BenchData, *repPtr)
	for i := 0; i < *repPtr; i++ {
		var sets [][]uint64
		var err error

		if cli_type == "chemical" {
			if chembl != "" {
				// Read compounds
				fmt.Println("Use chemicals loaded from a database")
				sets = ReadCompoundsFromFile(chembl, *nsPtr+1)
			} else {
				// random compounds
				sets, err = RandomDataSet(*nsPtr+1, 3, 64, sdSize)
				if err != nil {
					panic(err)
				}
			}
		} else if cli_type == "document" {
			// Random documents
			sets, err = RandomDataSet(*nsPtr+1, 8, maxDocSize-2, 10000)
			if err != nil {
				panic(err)
			}
			sets[0] = sets[0][:8] // sets[0] represent the query and must be smaller
		} else if cli_type == "sd-comparison" {
			// random sets
			sets, err = RandomDataSet(*nsPtr+1, 3, sdSize/2, sdSize)
			if err != nil {
				panic(err)
			}
		}

		fmt.Printf("Running benchmark with %v sets at %v.\n", *nsPtr, time.Now())
		data[i] = BenchHomoPSI(pp, sets, *qt)

		// run garbage collection
		runtime.GC()
	}

	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(*outAddrPtr, file, 0644)
}
