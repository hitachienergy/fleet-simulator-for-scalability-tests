#!/bin/bash

# $1            Initial yaml configuration
# $2            Simulation output directory 
# $3            yaml config variable field
# $4 ... $n     values to use for the field 
                # NOTE: if you want to use string values, enclose them in \" (e.g. "abc" becomes \"abc\")

# ===== Process input =====

echo "Yaml configuration:   $1"
echo "Output directory:  $2"
echo "Target field:  $3"

# convert trailing values to list
args=("${@:4}")
RANGE="${args[*]}"
RANGE=($RANGE)
if [ ${#RANGE[@]} -eq 0 ];
then
    echo "RANGE is not defined. Aborting process..."
fi

# load yaml configuration into a variable
main_config=$(cat "$1")

# ===== Run scenario variants =====

for value in "${RANGE[@]}"; do
    echo "Running test with value ${value} for field $3"
    
    # update simulation configuration
    config=$(yq e "$3 = $value" <<< "$main_config")
    result_dir="$2/$value"

    # run simulation
    ./run-simulation.sh "$config" $result_dir
    result_code=$?

    # capture user CTRL-C 
    if [ "$result_code" -eq 2 ] # (custom python script code)
    then
        exit 2
    fi
    
    echo "========================"
done