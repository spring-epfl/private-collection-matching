#!/bin/bash
out=$1

echo "start aggregating measures into $out.\n"

rm $out
echo $'[\n' > $out
cat $2 >> $out

set -- "${@:3}"

for i in "$@"; do
    echo "Concat file $i to $out"
    echo $',\n' >> $out
    cat $i >> $out
done

echo $'\n]' >> $out
