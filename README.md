# Private Collection Matching Protocols
This repository accompanies the paper *"Private set matching protocols"* by Kasra
EdalatNejad (SPRING Lab, EPFL), Mathilde Raynal(SPRING Lab, EPFL), Wouter Lueks (SPRING Lab, EPFL), and Carmela Troncoso (SPRING Lab, EPFL).


This repository contains the benchmark code for evaluating our protocol and aims to facilitate the reproducibility of measurements in the paper.

You can find an early version of the paper at [arXiv](https://arxiv.org/abs/2206.07009).


> **Abstract:**
> We introduce Private Collection Matching (PCM) problems, in which a client aims to determine whether a collection of sets owned by a server matches their interests.
> Existing privacy-preserving cryptographic primitives cannot solve PCM problems efficiently without harming privacy.
> We propose a modular framework that enables designers to build privacy-preserving PCM systems that output one bit: whether a collection of server sets matches the client's set.
> The communication cost of our protocols scales linearly with the size of the client's set and is independent of the number of server elements.
> We demonstrate the potential of our framework by designing and implementing novel solutions for two real-world PCM problems: determining whether a dataset has chemical compounds of interest, and determining whether a document collection has relevant documents.
> Our evaluation shows that we offer a privacy gain with respect to existing works at a reasonable communication and computation cost.


## Structure of the repository
This repository is structured as follows:
 - `GoPSI`: this folder contains the framework introduced in the paper.
 - `chemistry`: python code to extract smiles from the ChEMBL compound database and compute their fingerprint.
 - `bench`: includes scripts to run the benchmark and the python code to generate figures from raw measurements.
 - `Data`: includes raw benchmark measurements and raw data used for chemical search.
 - `relatedwork`: includes codes and dockers to reproduce the related work measurement in the paper (Section 11.3) 

