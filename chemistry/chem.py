from argparse import ArgumentParser
import json
from pathlib import Path
import sqlite3
import sys
from typing import Any, List

from rdkit import Chem
from rdkit.Chem import MACCSkeys


def extract_data_from_db(db_path: Path, query: str) -> List[str]:
    if not db_path.is_file():
        raise Exception("Chembl database not found!")

    try:
        connection = sqlite3.connect(database=str(db_path))
        print("Connection to SQLite DB successful")

        cursor = connection.cursor()
        cursor.execute(query)
        result = cursor.fetchall()
        return result
    except Exception as e:
        print(f"The error '{e}' occurred")


def read_smiles_from_db(db_path: Path) -> List[str]:
    smiles = extract_data_from_db("select CANONICAL_SMILES from compound_structures;")
    with open('smiles.json', 'w') as fd:
        json.dump(smiles, fd)
    return smiles


def read_smiles_from_json(smiles: Path) -> List[str]:
    with open('smiles.json', 'r') as fd:
        return json.load(fd)


def extract_fingerprint(smiles: List[str]) -> Any:
    mol = Chem.MolFromSmiles(smiles)
    return MACCSkeys.GenMACCSKeys(mol)


def compute_fp_max_onbits(fps):
    fp_ons = [fp.GetNumOnBits() for fp in fps]
    return sorted(fp_ons)[-1]


def main():
    parser = ArgumentParser(sys.argv[0])
    parser.add_argument(
        "-D",
        "--database",
        help="ChEMBL database in SQLite format.",
        type=Path
    )
    parser.add_argument(
        "-o",
        "--output",
        help="File where to write the extracted fingerprints.",
        type=Path,
        default=(Path.cwd() / "fps.txt")
    )
    namespace = parser.parse_args(sys.argv[1:])

    db_path: Path = namespace.database.resolve()
    fps_path: Path = namespace.output.resolve()

    smiles = read_smiles_from_db(db_path)
    #test_fp = extract_fingerprint('CC(C)C1=C(C(=C(N1CC[C@H](C[C@H](CC(=O)O)O)O)C2=CC=C(C=C2)F)C3=CC=CC=C3)C(=O)NC4=CC=CC=C4')
    # fps = [extract_fingerprint(sm[0]) for sm in smiles]

    with fps_path.open("w") as fd:
        for idx, sm in enumerate(smiles):
            try:
                fp = extract_fingerprint(sm[0])
                fd.write(fp.ToBitString()+'\n')
                if (idx + 1) % 1000 == 0:
                    print(f"Proccessed {idx + 1}/{len(smiles)} compounds.")
            except Exception as err:
                print("warning: Failed to process a compound!")
                print(str(err))


if __name__ == "__main__":
    main()
