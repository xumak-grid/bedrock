#!/bin/bash
kubectl create secret generic bedrock-api-secrets \
  -o yaml --dry-run \
  -n bedrock \
  --from-literal=vault-address="$VAULT_ADDR" \
  --from-literal=vault-token="$VAULT_TOKEN" \
  --from-literal=aws_access_key="$AWS_ACCESS_KEY" \
  --from-literal=aws_secret_key="$AWS_SECRET_KEY"
