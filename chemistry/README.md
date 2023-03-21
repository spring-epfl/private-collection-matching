# Chemistry tools: preparing molecular datasets

One of our  use cases is privately searching a chemical compound database, where a client has a single compound and wants to know if the server has one or more similar compounds in its database. Before we can use our PCM framework, we must process the descriptions of chemical elements to compute their fixed-length fingerprints. This folder contains the tools to do so.

If you are not interested in testing the framework with a large (2 million) dataset, you can also use the smaller dataset provided in `data/raw_chem/` which contains 8,000 compounds and is provided in preprocessed form.

**Warning:** While we provide scripts to manage working with the full 2 million compound dataset, doing so requires downloading a **3GB** compressed dataset and storing a **> 20GB** dataset during the processing.


## Obtaining a Molecular Dataset

In our experiments, we used fingerprints extracted from the [ChEMBL dataset](https://chembl.gitbook.io/chembl-interface-documentation/downloads) [1]. In particular, we have tested these scripts against the [v28 version of this dataset](http://doi.org/10.6019/CHEMBL.database.28).

To directly download the dataset, run:

```
wget https://ftp.ebi.ac.uk/pub/databases/chembl/ChEMBLdb/releases/chembl_28/chembl_28_sqlite.tar.gz
```

You can extract this file to get the `chembl_28.db` file. This dataset v28 with more than 2 million compounds for evaluating our system.

You can run the following command to extract fingerprints from this db:

```
$ python -m venv venv
$ source venv/bin/activate
$ pip install -r requirements.txt
$ python chem.py -D chembl_28/chembl_28_sqlite/chembl_28.db -o fps.txt
```

The outputs are stored in the file `fps.txt`.


## Files

 - `chem.py` allows reading smiles from the dataset and uses [rdkit](https://www.rdkit.org/) to convert them into MACCS keys.
 - `fp_stats.py` compute general statistic over fingerprints

## References

[1] Mendez D, Gaulton A, Bento AP, Chambers J, De Veij M, Félix E, Magariños MP, Mosquera JF, Mutowo P, Nowotka M, Gordillo-Marañón M, Hunter F, Junco L, Mugumbate G, Rodriguez-Lopez M, Atkinson F, Bosc N, Radoux CJ, Segura-Cabrera A, Hersey A, Leach AR. ChEMBL: towards direct deposition of bioassay data. Nucleic Acids Res. 2019 Jan 8;47(D1):D930-D940. doi: 10.1093/nar/gky1075. PMID: 30398643; PMCID: PMC6323927.
