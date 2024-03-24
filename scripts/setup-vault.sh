#!/usr/bin/env bash
source .env

vault audit enable file file_path=stdout

vault policy write admin policies/admin.hcl
vault policy write global-reader policies/global_reader.hcl
vault policy write webviewer policies/webviewer.hcl

vault auth enable userpass
vault write auth/userpass/users/a \
    password=a \
    policies=admin
vault write auth/userpass/users/b \
    password=b \
    policies=global-reader
vault write auth/userpass/users/c \
    password=c

vault auth enable approle
vault write auth/approle/role/webviewer \
    token_policies="webviewer"
    # secret_id_ttl=10m \
    # token_num_uses=10 \
    # token_ttl=20m \
    # token_max_ttl=30m \
    # secret_id_num_uses=40 \
ROLE_ID=$(vault read auth/approle/role/webviewer/role-id -format=json | jq -r .data.role_id)
SECRET_ID=$(vault write -force auth/approle/role/webviewer/secret-id -format=json | jq -r .data.secret_id)
kubectl delete secrets webviewer-approle
kubectl create secret generic webviewer-approle \
    --from-literal="role-id=$ROLE_ID" \
    --from-literal="secret-id=$SECRET_ID"

vault secrets enable transit
vault secrets enable database

vault write database/config/postgresql \
    plugin_name=postgresql-database-plugin \
    connection_url="postgresql://{{username}}:{{password}}@$POSTGRES_URL/postgres?sslmode=disable" \
    allowed_roles=readonly \
    username="postgres" \
    password="rootpassword"

vault write database/roles/readonly \
    db_name=postgresql \
    creation_statements=@readonly.sql \
    default_ttl=1h \
    max_ttl=24h