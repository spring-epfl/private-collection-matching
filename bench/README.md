# Benchmark
The code in this directory facilitates benchmarking and comparing our system and related work.
The scripts are structured as follows:

 - `bench.sh`: runs the benchmark experiments in our paper.
 - `plotting.ipynb`: a jupyter notebook to produce the plots in our paper.
 - `data_cleaner.py`: helper functions to clean and load aggregated data in `data/agg`
 - `aggregate.sh`: combines multiple GoPSI measurements (`.json` files) into a single json. 
 - `pase_emp.py`: a cli to extract measurement from EMP log files.

## running benchmark
Update after building docker.


## Plots
We have uploaded our measurement data in `data/agg`, so you can use explore the data in `plotting.ipynb` without having to run experiments.