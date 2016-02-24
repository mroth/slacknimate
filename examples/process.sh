#!/bin/bash
N=${1:-10}
for i in $(seq 0 $N); do
  echo "Processing items: $i/$N"
  sleep 1
done
