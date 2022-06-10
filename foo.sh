#!/bin/bash

keycloakBaseURL=$(yq e .keycloak.production.baseurl ~/.cx-cli.yaml)
keycloakRealm=$(yq e .keycloak.production.realm ~/.cx-cli.yaml)
keycloakClientID=$(yq e '.keycloak.production.clients.[] | select(.id == "admin-service").name' ~/.cx-cli.yaml)
keycloakClientSecret=$(yq e '.keycloak.production.clients.[] | select(.id == "admin-service").secret' ~/.cx-cli.yaml)

token=$(curl -q \
    --data-urlencode "grant_type=client_credentials" \
    --data-urlencode "client_id=${keycloakClientID}" \
    --data-urlencode "client_secret=${keycloakClientSecret}" \
    "${keycloakBaseURL}/auth/realms/${keycloakRealm}/protocol/openid-connect/token" 2>/dev/null \
    | jq -r .access_token)

set -x
curl -H "Authorization: Bearer ${token}" ${keycloakBaseURL}/auth/admin/realms/${keycloakRealm}/authentication/required-actions
