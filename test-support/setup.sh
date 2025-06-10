#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

modal deploy libmodal_test_support.py

echo "Deploying libmodal-test-secret..."
modal secret create --force libmodal-test-secret \
  a=1 b=2 c="hello world" >/dev/null

# Must be signed into AWS CLI for Modal Labs
echo "Deploying libmodal-ghcr-test..."
ecr_test_secret=$(aws secretsmanager get-secret-value \
  --secret-id test/libmodal/GhcrTest --query 'SecretString' --output text)
modal secret create --force libmodal-ghcr-test \
  USERNAME="$(echo "$ecr_test_secret" | jq -r '.GHCR_USERNAME')" \
  PASSWORD="$(echo "$ecr_test_secret" | jq -r '.GHCR_PASSWORD')" \
  >/dev/null

echo "Deploying libmodal-aws-ecr-test..."
ecr_test_secret=$(aws secretsmanager get-secret-value \
  --secret-id test/libmodal/AwsEcrTest --query 'SecretString' --output text)
modal secret create --force libmodal-aws-ecr-test \
  AWS_ACCESS_KEY_ID="$(echo "$ecr_test_secret" | jq -r '.AWS_ACCESS_KEY_ID')" \
  AWS_SECRET_ACCESS_KEY="$(echo "$ecr_test_secret" | jq -r '.AWS_SECRET_ACCESS_KEY')" \
  AWS_REGION=us-east-1 \
  >/dev/null

echo "Deploying libmodal-gcp-artifact-registry-test..."
ecr_test_secret=$(aws secretsmanager get-secret-value \
  --secret-id test/libmodal/GcpArtifactRegistryTest --query 'SecretString' --output text)
modal secret create --force libmodal-gcp-artifact-registry-test \
  SERVICE_ACCOUNT_JSON="$(echo "$ecr_test_secret" | jq -r '.SERVICE_ACCOUNT_JSON')" \
  >/dev/null
