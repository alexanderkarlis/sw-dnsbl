#!/bin/bash

if [[ $# -eq 0 ]] ; then
    port=8080
else
    port=$1
fi

docker build -t akarlis/sw-dnsbl -f ./Dockerfile .
docker run --rm -d -p $port:8080 akarlis/sw-dnsbl:latest
# docker run --rm -p $port:8080 akarlis/sw-dnsbl:latest
