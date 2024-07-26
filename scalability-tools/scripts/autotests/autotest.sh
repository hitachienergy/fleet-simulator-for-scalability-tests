#!/bin/bash

export TIMES=1      # how many times will a single configuration be ran
export DELAY=180    # how much time to wait between each simulation

# ===== EXAMPLES =====

# main_dest_dir="/home/gismo/fist_workspace/scalability-report/data/dummy_works"

# dest_dir="results_1"
# original_yaml="example.yaml"
# field=".simulation.dummyWork.duration"
# ./autotest-variations.sh $original_yaml $dest_dir $field 10s 20s 30s

# dest_dir="results_2"
# original_yaml="example.yaml"
# ./run-simulation.sh $original_yaml $dest_dir

# ===== SIMULATIONS =====

# test all configurations

dest_dir="/home/gismo/fist_workspace/scalability-report/data/tests/ddi-default"
original_yaml="/home/gismo/fist_workspace/scalability-report/scalability-tools/examples/hawkbit/ddi-default.yaml"
./run-simulation.sh $original_yaml $dest_dir

dest_dir="/home/gismo/fist_workspace/scalability-report/data/tests/dmf-default"
original_yaml="/home/gismo/fist_workspace/scalability-report/scalability-tools/examples/hawkbit/dmf-default.yaml"
./run-simulation.sh $original_yaml $dest_dir

dest_dir="/home/gismo/fist_workspace/scalability-report/data/tests/http-default-kafka"
original_yaml="/home/gismo/fist_workspace/scalability-report/scalability-tools/examples/thingsboard/http-default-kafka.yaml"
./run-simulation.sh $original_yaml $dest_dir

dest_dir="/home/gismo/fist_workspace/scalability-report/data/tests/http-default-mem"
original_yaml="/home/gismo/fist_workspace/scalability-report/scalability-tools/examples/thingsboard/http-default-mem.yaml"
./run-simulation.sh $original_yaml $dest_dir
