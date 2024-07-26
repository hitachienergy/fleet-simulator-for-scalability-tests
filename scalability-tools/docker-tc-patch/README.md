# Docker Traffic Control
This folder contains a patch for a Docker-based Linux traffic control system named docker-tc.
The original project is visible on (GitHub)[https://github.com/lukaszlach/docker-tc].

## What is docker-tc?
Docker Traffic Control allows you to set a rate limit on the container network and can emulate network conditions like delay, packet loss, duplication, and corruption for Docker containers, all based only on labels.

## Why do we need a patch?

We expanded docker-tc functionalities by fixing 2 hard-coded fields (latency and burst) for the rate limiting API.

These 2 variables are linked to some linux-tc configuration that shapes and limits incoming traffic. Under normal circumstances, these hard-coded configurations are good enough, but errors may occur when different transport protocols are used or when simulations create high traffic intensities.

## How to apply the patch?

Clone the docker-tc repo:

`
git clone https://github.com/lukaszlach/docker-tc.git
cd docker-tc
`

Apply the provided patch:

`
git am ../0001-customization.patch
`

Build a patched image of docker-tc:

`
docker build --no-cache -t lukaszlach/docker-tc .
`
