#!/bin/sh

# Prepare 2PC-Circuit-PSI

cd /psi/2PC-Circuit-PSI/
mkdir build
cd build
cmake ..
cp ../aux_hash/* ../extern/HashingTables/cuckoo_hashing/
cp ../com_hash/* ../extern/HashingTables/common/
make

# Prepare emp

cd /psi/emp/
git clone https://github.com/PizzaWhisperer/emp-tool.git
cmake .
make
make install
