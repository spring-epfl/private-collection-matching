# Private Collection Matching Protocols
This repository accompanies the paper *"Private Collection Matching Protocols"* by Kasra
EdalatNejad (SPRING Lab, EPFL), Mathilde Raynal(SPRING Lab, EPFL), Wouter Lueks (CISPA Helmholtz Center for Information Security), and Carmela Troncoso (SPRING Lab, EPFL).

> **Abstract:**
> We introduce Private Collection Matching (PCM) problems, in which a client aims to determine whether a collection of sets owned by a server matches their interests.
> Existing privacy-preserving cryptographic primitives cannot solve PCM problems efficiently without harming privacy.
> We propose a modular framework that enables designers to build privacy-preserving PCM systems that output one bit: whether a collection of server sets matches the client's set.
> The communication cost of our protocols scales linearly with the size of the client's set and is independent of the number of server elements.
> We demonstrate the potential of our framework by designing and implementing novel solutions for two real-world PCM problems: determining whether a dataset has chemical compounds of interest, and determining whether a document collection has relevant documents.
> Our evaluation shows that we offer a privacy gain with respect to existing works at a reasonable communication and computation cost.

## This repository

This repository serves four goals:

 1. It contains the **implementation of the framework** in Go, as well as three example applications that use the framework to solve the problems in the paper. The implementations and the applications can be found in the `GoPSI` directory. Our implementations can be evaluated by running the `bench/bench.sh` script. Running the full benchmark set requires some extra data that you will need to retrieve and process, please check the next point to see how to prepare this data.

 2. We provide instructions for **building the chemical compound dataset** used in our benchmark in `chemistry/`. The full set of benchmarks requires a large database of fingerprints of chemical compounds. Preparing the full benchmark requires downloading a large compressed database (*~3GB*), decompressing it (requires 1-2 mins of processing and **~20 GB** of storage), and processing it with our provided scripts (takes *~20-30* minutes of processing). If you prefer to skip this step, we also provide a smaller subset of the database in `data/raw_chem/fps-mini.txt` to test the framework.

 3. To reproduce our **evaluation of related work** that was included in the paper. We compared the implementation based on our framework with two other approaches (1) based on a generic SMC compiler (EMP) and (2) an extensible Circuit PSI protocol. We implemented appropriate circuits using the SMC compiler. The `relatedwork/` directory contains the benchmarking infrastructure that we used to run these works and to evaluate their results.

 4. To enable **reproducing the graphs** in the paper. The scripts in `bench/` capture the full pipeline from benchmarking, processing raw measurements, and finally preparing the graphs in the paper.

 5. To **store our measurements** (both of our own implementations as well as those of related work). Raw measurement data can be found in `data/raw`. This directory contains both evaluations using our framework, as well as the related works. The `data/agg` directory contains the same data in post-processed form to simplify plotting. You can read the README in the `bench` directory for more information on how to load and use this data.
