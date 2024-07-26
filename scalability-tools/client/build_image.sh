#!/bin/bash

# go clean -cache 

# go build -o simulator .

docker build --no-cache -t device-simulator -f docker/Dockerfile .
