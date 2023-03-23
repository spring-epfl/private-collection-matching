#!/bin/sh
repeats=2

exec_dir=../GoPSI/cmd/
data_dir=../data/
bench_dir=../data/bench

mkdir -p $bench_dir

#################################################################
# Running document search 
# Matching F-PSM
# Uses small input 
# Maximum query size 8 keywords.
# Each keyword contains 2 (hash) elements ->  FP rate ~ 2^-40
# Maximum server set size 128 keywords
for i in 1 2 4 8 16 # 32 64 128 256 512 1024 2048 4096 8192
do
    # Aggregation CA-MS
    # The query ciphertext contains 4 replicates of the base query (2^13 / (8 * 2 * 128))
    echo "$(date)" 1>&2
    echo "Running document search (CA) with $i server sets:" 1>&2
    ii=$((2*i))

    $exec_dir/doc_search/doc_search -logn 13 -ns ${i} -max-q 8 -hash-per-kw 2 -max-doc 128 -agg "ca-ms" -r "${repeats}" -o "$bench_dir/doc_ca_ms_ns${i}.json"  -v
done
for i in 1 2 4 8 16 # 32 64 128 256 512 1024 2048 4096 8192
do
    # Aggregation X-MS
    # The query ciphertext contains 16 replicates of the base query (2^15 / (8 * 2 * 128))
    echo "$(date)" 1>&2
    echo "Running document search (X-MS) with $i server sets:" 1>&2
    ii=$((2*i))

    $exec_dir/doc_search/doc_search -logn 15 -ns ${i} -max-q 8 -hash-per-kw 2 -max-doc 128  -agg "x-ms" -r "${repeats}"  -o "$bench_dir/doc_x_ms_ns${i}.json" -v
done



#################################################################
# Running chemical search
# Matching Tversky
# Uses small domain
# Compound finger print size |D| = 167 
# Use compounds from ChEMBL dataset (The mini file only contains 8k compounds)
for i in 1000 2000 4000 8000 
# 16000 32000 64000 128000 256000 512000 1000000 2000000 (requires full dataset)
do
    echo "Running chemical search with $i server sets:" 1>&2

    # BUG Aggregation CA-MS
    echo "$(date)" 1>&2
    $exec_dir/chem_search/chem_search  -logn 15 -ns $i -agg "" -chemdb-path "$data_dir/raw_chem/fps-mini.txt" -o "$bench_dir/chem_p15_ns$i.json" -r $repeats -v

    # Aggregation X-MS
    echo "$(date)" 1>&2
    $exec_dir/chem_search/chem_search  -logn 15  -ns $i -agg "x-ms" -chemdb-path "$data_dir/raw_chem/fps-mini.txt" -o "$bench_dir/chem_p15_ns${i}_xms.json" -r $repeats
done



#################################################################
# # Running small domain comparison
# # No matching or aggregation
for i in  4 8 16 32 64 128 # 256 512 1024
do
    echo "$(date)" 1>&2

    # Domain size = 256
    echo "Running small domain (CA-256) with $i server sets:" 1>&2
    $exec_dir/small_domain_bench/small_domain_bench -logn 13 -ns $i -r $repeats -sd-domain-size 256 -o "$bench_dir/sm_ca_256_ns$i.json"
   
    # Domain size = 4096 
    echo "Running small domain (CA-4096) with $i server sets:" 1>&2
    $exec_dir/small_domain_bench/small_domain_bench -logn 13 -ns $i  -r $repeats -sd-domain-size 4096 -o "$bench_dir/sm_ca_4k_ns$i.json"
    sleep 1
done
