#!/usr/bin/env bash
source .env

curl \
    --header "X-Vault-Token: $VAULT_TOKEN" \
    --request GET \
    --data '{"type": "userpass"}' \
    $VAULT_ADDR/v1/sys/auth