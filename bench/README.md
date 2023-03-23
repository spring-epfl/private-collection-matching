# Benchmark

The code in this directory facilitates benchmarking of our framework as well as postprocessing of all measurements and creating the plots used in the figures. In general, there is no need to rerun any of the benchmarks to reproduce the figures as we stored our measurements in `data/agg`. But this README provides instructions for how to reproduce these.

## Installing 
We use bash and python scripts in this directory. You can download and install python from [here](https://www.python.org/).

### Installing dependencies
You can install python dependencies using:
```
$ python3 -m venv venv
$ source venv/bin/activate
$ pip install -r requirements.txt
```

## Benchmarking our Framework
The 3 benchmarking programs `doc_search`, `chem_search`, and `small_domain_bench` that we introduced for our framework -- see `GoPSI/` -- evaluate a single parameter setting in each execution, e.g., using a fixed number of server sets. To produce the measurements in the paper, we called these individual programs many times. To reproduce the measurements in the paper run:

```
./bench.sh
```

We modified the script in the repository to only run with a relatively small number of server settings. Running the script with these parameters takes less than 30 minutes to run. If you wish to reproduce the full dataset you have to (1) edit the script to adjust the ranges of the evaluation loops (the original values are commented out), and (2) ensure you have prepared the full chemical database by following the instructions in `chemical/`.

After running the `bench.sh` script, the new measurements will have ended up in `data/bench/`, in separate `.json` files for each experiment. Our following scripts expect these files to be joined. You can do so by running:

```
cd ../data/bench/
../../bench/aggregate.sh doc_ca_ms.json          doc_ca_ms_*.json
../../bench/aggregate.sh doc_x_ms.json           doc_x_ms_*.json
../../bench/aggregate.sh pcm_chem_ca_agg.json    chem_p15_*.json
../../bench/aggregate.sh pcm_chem_x_agg.json     chem_p15_ns*_xms.json
../../bench/aggregate.sh 7700_sm_ca_256_agg.json sm_ca_256_*.json
../../bench/aggregate.sh 7700_sm_ca_4k_agg.json  sm_ca_4k_*.json
```

If you wish to use the new measurements instead of the supplied files, make sure the aggregated files to `data/agg`:

```
mv doc_ca_ms.json doc_x_ms.json pcm_chem_ca_agg.json pcm_chem_x_agg.json 7700_sm_ca_256_agg.json 7700_sm_ca_4k_agg.json ../agg/
```

## Benchmarking related work

We refer to the `relatedwork/` directory for all information on how to install and benchmark the related work that we evaluate in the paper. The resulting log files from EMP need a little bit of postprocessing and parsing to make them easier to use. The `parse_emp.py` tool takes care of this. If you want to overwrite the supplied files, run:


```
python3 parse_emp.py ../relatedwork/emp/emp_ca_client.log > ../bench/raw/emp_ca_client.csv
python3 parse_emp.py ../relatedwork/emp/emp_ca_server.log > ../bench/raw/emp_ca_server.csv
python3 parse_emp.py ../relatedwork/emp/emp_x_client.log > ../bench/raw/emp_x_client.csv
python3 parse_emp.py ../relatedwork/emp/emp_x_server.log > ../bench/raw/emp_x_server.csv
```

The Circuit-PSI file scan directly be used, they only need to be moved to the right place (again, if you want to overwrite previous results):

```
cp ../relatedwork/2PC-Circuit-PSI/benchmarks/circuit-time-client.log ../data/raw/circuit-time-client.csv
cp ../relatedwork/2PC-Circuit-PSI/benchmarks/circuit-time-server.log ../data/raw/circuit-time-server.csv
```

The original data that we collected for `SpOT` are stored in `data/agg/spot.csv`. But we were unfortunately not able to reliably build SpOT and therefore do not include instructions for gathering these data.


## Reproducing Figures

We supplied a Jupyter notebook to reproduce the graphs in the paper. You can either explore it directly on GitHub or open it locally by calling:

```
jupyter notebook plotting.ipynb
```

The notebook uses the utilities provided in `data_cleaner.py` to import the CSV files, see the "Load Data" parts of the notebook for how these utilities are used to ingest the csv files that we produced earlier. Our measurements are supplied in `data/agg` so you should be able to produce the figures without rerunning the experiments.
