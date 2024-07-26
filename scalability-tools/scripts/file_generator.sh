#!/bin/bash

output_folder=$1

# this option generates alphanumeric files but it is quiet slow 
function generate_alphanumeric() {
    size=$1
    output_file=$2
    file_size=$(( ((size/12)+1)*12 ))
    read_size=$((file_size*3/4))

    echo "wanted=$size file=$file_size read=$read_size"

    dd if=/dev/urandom bs=$read_size count=1 | base64 > $output_file

    truncate -s "$size" $output_file
}

generate_alphanumeric 1048576 "$output_folder/1MB.txt"
generate_alphanumeric 2097152 "$output_folder/2MB.txt"
generate_alphanumeric 4194304 "$output_folder/4MB.txt"
generate_alphanumeric 8388608 "$output_folder/8MB.txt"
generate_alphanumeric 16777216 "$output_folder/16MB.txt"
generate_alphanumeric 33554432 "$output_folder/32MB.txt"

generate_alphanumeric 73400320 "$output_folder/70MB.txt"
generate_alphanumeric 58720256 "$output_folder/56MB.txt"
generate_alphanumeric 44040192 "$output_folder/42MB.txt"
generate_alphanumeric 29360128 "$output_folder/28MB.txt"
generate_alphanumeric 14680064 "$output_folder/14MB.txt"

# # this option generates alphanumeric files but it is quiet slow 
# function generate_alphanumeric() {
#     size=$1
#     output_file=$2
#     dd if=/dev/urandom bs=1 count=$size | tr -dc 'a-zA-Z0-9' > $output_file
# }

# # this option uses an ascii charset and it's much faster 
# function generate() {
#     size=$1
#     output_file=$2
#     head -c $size < /dev/urandom > $output_file
# }

# generate 1048576 "$output_folder/1MB.txt"
# generate 2097152 "$output_folder/2MB.txt"
# generate 4194304 "$output_folder/4MB.txt"
# generate 8388608 "$output_folder/8MB.txt"
# generate 16777216 "$output_folder/16MB.txt"
# generate 33554432 "$output_folder/32MB.txt"