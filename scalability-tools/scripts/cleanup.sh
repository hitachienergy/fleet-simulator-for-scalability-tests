#!/bin/bash

# Remove all Docker containers
docker rm -f $(docker ps -a -q)

# Remove all Docker volumes
docker volume rm -f $(docker volume ls -q)

# Remove all non-default Docker networks
docker network prune -f
