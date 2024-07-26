#!/bin/bash

for container in $(docker ps --format '{{.Names}}'); do
    iflink=`docker exec -it $container sh -c 'cat /sys/class/net/eth*/iflink'`
    for net in $iflink;do
        net=`echo $net|tr -d '\r'`
        veth=`grep -l $net /sys/class/net/veth*/ifindex`
        veth=`echo $veth|sed -e 's;^.*net/\(.*\)/ifindex$;\1;'`
        echo $container:$veth
    done
done