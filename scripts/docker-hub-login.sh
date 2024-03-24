#!/usr/bin/env bash
source .env

if [[ $# -ne 3 ]]; then
    echo "must supply arguments username password email"
    exit 1
fi

kubectl create secret docker-registry regcred --docker-server=https://index.docker.io/v1/ --docker-username=$1 --docker-password=$2 --docker-email=$3
