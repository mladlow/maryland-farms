#!/bin/bash

# This code takes individual stables and puts them in a large file in the form
# of a JSON array.

echo "[" > ./data/stables.json
for jsonFile in ./data/json/*.json
do
  cat $jsonFile >> ./data/stables.json
  echo "," >> ./data/stables.json
done

echo "]" >> ./data/stables.json
