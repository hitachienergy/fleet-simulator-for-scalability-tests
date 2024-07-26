#!/bin/bash

# $1    string configuration in yaml format or path to yaml configuration
# $2    path to main output folder 

if [[ -f $1 ]]; then
    config=$(cat "$1")
else
    config=$1
fi

mkdir -p $2

for ((i=1; i<=${TIMES}; i+=1)); do
    echo -n "Running test ${i}/${TIMES}"

    # create the result logs and data paths
    result_dir="$2/${i}"
    mkdir -p ${result_dir}
    rm -rf "$result_dir"/*

    # start a simulation and redirect std out and error to logs file
    python3 -u ../../simulator/src/fistsim/main.py --config "$config" > "${result_dir}/simulation_logs.log" 2>&1
    result_code=$?
    output_path=$(echo "$config" | yq -r ".output.path")
    mv -f $output_path/* $result_dir

    
    # # start a simulation and redirect std out and error to logs file
    # echo "TEST" > "${result_dir}/simulation_logs.log"       # TEST 
    # output_path=$(echo "$config" | yq -r ".output.path")
    # mkdir -p $output_path                                   # TEST 
    # echo "$config" > $output_path/config.yaml               # TEST 
    # mv -f $output_path/* $result_dir                        
    # result_code=0                                           # TEST

    # capture user CTRL-C 
    if [ "$result_code" -eq 2 ] # custom python script code for CTRL-C
    then
        echo ""
        echo "Detected CTRL-C while a simulation was running. Aborting all following simulations..."
        exit 2
    elif [ "$result_code" -eq 3 ] # custom python script code for simulation timeout
    then 
        echo "TIMED OUT"
    else
        echo "COMPLETED"
    fi
    
    sleep $DELAY

done