#!/usr/bin/env fish
source .env

for cred in (vault list -format=json sys/leases/lookup/database/creds/readonly | jq -r .[])
    vault lease revoke database/creds/readonly/$cred
end