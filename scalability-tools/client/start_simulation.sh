#!/bin/bash

# Only starts the device simulator. No server, network simulation or setup/trigger scripts

if [[ $# -lt 1 ]]; then
  echo "Usage: ./start_simulation.sh <config_file>"
  exit 1
fi

if ! command -v yq &> /dev/null; then
    echo "yq is not installed. Installing..."
    sudo wget https://github.com/mikefarah/yq/releases/download/v4.13.2/yq_linux_amd64 -O /usr/local/bin/yq
    sudo chmod +x /usr/local/bin/yq
fi

config_path=$(realpath $1)
mode=$(yq eval '.client.mode.type' "$1")
template_path=$(realpath $(yq eval '.client.template' "$1"))
client_number=$(yq eval '.client.number' "$1")
if [[ $client_number -eq 0 ]]; then 
    echo "The number of clients must be greater than zero"
    exit 1
fi

if [[ "$mode" = "threads" ]]; then
  echo "Start simulation in threads mode"
  go build -o simulator .
  ./simulator --config $1
elif [[ "$mode" = "container" ]]; then
  docker run --name "device-simulator" -it \
            --network host \
            -v $config_path:/app/config.yaml \
            -v $template_path:/app/plugin \
            device-simulator
elif [[ "$mode" = "hybrid" ]]; then
  container_num=$(yq eval '.client.mode.containerNum' "$1")
  if [[ $container_num -eq 0 ]]; then 
    echo "The number of containers must be greater than zero"
    exit 1
  fi

  avg_client_per_container=$(( client_number / container_num + 1))
  remainder=$(( client_number % container_num ))
  
  for ((i=0; i<container_num; i++)); do
    if [[ $i -eq $remainder ]]; then
      avg_client_per_container=$((avg_client_per_container - 1))
    fi
    docker run --name $(echo "device-simulator-$i") -d \
            --network host \
            -v $config_path:/app/config.yaml \
            -v $template_path:/app/plugin \
            -e CLIENT_NUM=$avg_client_per_container \
            device-simulator
  done
else
  echo "Unrecognized simulation mode"
  exit 1
fi