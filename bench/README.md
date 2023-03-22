# Benchmark
The code in this directory facilitates benchmarking and comparing our system and related work.
The scripts are structured as follows:

 - `bench.sh`: runs the benchmark experiments in our paper.
 - `plotting.ipynb`: a jupyter notebook to produce the plots in our paper.
 - `data_cleaner.py`: helper functions to clean and load aggregated data in `data/agg`
 - `aggregate.sh`: combines multiple GoPSI measurements (`.json` files) into a single json. 
 - `pase_emp.py`: a cli to extract measurement from EMP log files.

## Installing 
We use bash and python scripts in this directory. You can download and install python from [here](https://www.python.org/).

### Installing dependencies
You can install python dependencies using:
```
$ python3 -m venv venv
$ source venv/bin/activate
$ pip install -r requirements.txt
```

## Running benchmark
Our 3 benchmarking programs `doc_search`, `chem_search`, and `small_domain_bench` run our framework with a fixed number of server sets in each execution. We provide the `bench.sh` script to run our programs with varying numbers of server sets with the same configuration as our measurements in Section 11.

## Cleaning and loading data
The output of our benchmark script for `EMP`, `SpOT`, and `Circuit-PSI` are human-readable, but it is hard to directly import them into code. In the `data_cleaner.py` we developed scripts that enable us to automatically import these data as a panda dataframe. 

### Document search 
You can use `load_full_df` and `create_individual_dfs` with the existing measurement files to load all measurements for document search (Sections 11.2 and 11.3). You can use the 'load data' part of `plotting.ipynb` as an example of how to load document search data.

### Chemical search
Running our `chem_search` program in `GoPSI/cmd/chem_search` produces measurements in the json format. You ca use the `aggregate.sh` script provided in this directory to merge multiple json files into a single json file. Finally, you can use `load_json_bench` function from the `data_cleaner.py` to load this file as a pandas dataframe. You can use 'Chemical figures - load data' part of `plotting.ipynb` as an example of how to load chemical search data. 

### Small domain search
We compare the performance of running PSI-CA on many small domain sets in Appendix D.1 of our paper. We generate this figure as 'Small-domain PSI-CA benchmark figure' in `plotting.ipynb`. This comparison uses data from two systems: Ruan et al. and ours. Since Ruan et al. do not provide code, we use their micro-benchmark to compute the cost of performing PSI-CA in a many-set setting. You can read this computation in 'Small-domain PSI-CA benchmark figure - load data' section of our `plotting.py`. We executed our small-domain program with the same machine configuration as Ruan et al. (CPU: Core i7 7700, RAM: 16GB). The measurements are available as `7700_sm_ca_256_agg.json` and `7700_sm_ca_4k_agg.json` in the `data\agg` directory. You can load these measurements with the `load_json_bench` in the same manner as our chemical search measurements.

## Plots
We have uploaded our measurement data in `data/agg`, so you can use explore the data in `plotting.ipynb` without having to run experiments.
