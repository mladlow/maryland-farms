#!/bin/bash

echo "[" > ./data/stables.json
for jsonFile in ./data/json/*.json
do
  cat $jsonFile >> ./data/stables.json
  echo "," >> ./data/stables.json
done

echo "]" >> ./data/stables.json
