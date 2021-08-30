#!/bin/bash

# getting vault token
if [ "${VAULT_TOKEN}" == "" ]; then
  VAULT_TOKEN=$(./vault write -field=token auth/kubernetes/login role=${VAULT_ROLE} jwt=@${CWA_VAULT_TOKENFILE})
fi

DB_USER=$(VAULT_TOKEN=${VAULT_TOKEN} ./vault kv get -field=user ${CWA_MAP_VAULT_BACKEND}/database)
DB_PASSWORD=$(VAULT_TOKEN=${VAULT_TOKEN} ./vault kv get -field=password ${CWA_MAP_VAULT_BACKEND}/database)
DB_HOST=$(VAULT_TOKEN=${VAULT_TOKEN} ./vault kv get -field=host ${CWA_MAP_VAULT_BACKEND}/database)
DB_PORT=$(VAULT_TOKEN=${VAULT_TOKEN} ./vault kv get -field=port ${CWA_MAP_VAULT_BACKEND}/database)
DB_DATABASE=$(VAULT_TOKEN=${VAULT_TOKEN} ./vault kv get -field=database ${CWA_MAP_VAULT_BACKEND}/database)
DB_URL=jdbc:postgresql://${DB_HOST}:${DB_PORT}/${DB_DATABASE}

./flyway/flyway -user=${DB_USER} -password=${DB_PASSWORD} -url=${DB_URL} ${CWA_DB_CLEAN} migrate
if [ $? -ne 0 ]; then
  exit 1
fi

./app/backend