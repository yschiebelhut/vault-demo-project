#!/usr/bin/env bash
source .env

if [[ -z $KUBECONFIG ]]; then
    echo "environment variable KUBECONFIG must be set"
    exit 1
fi

helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update

kubectl create namespace vault
helm install -n vault vault hashicorp/vault -f values-vault.yaml

kubectl create namespace postgre
helm install -n postgre postgre oci://registry-1.docker.io/bitnamicharts/postgresql -f values-postgre.yaml