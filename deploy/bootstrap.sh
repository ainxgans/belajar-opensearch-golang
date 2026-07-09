#!/bin/sh
set -e

OS_URL="${OPENSEARCH_URL:-http://opensearch:9200}"
INDEX="${OPENSEARCH_INDEX:-products}"
CONNECT_URL="${CONNECT_URL:-http://connect:8083}"
CONNECTOR_NAME="postgres-catalog-connector"

echo "Waiting for dependencies is handled by compose healthchecks; running bootstrap..."

# 1. Create OpenSearch index if it doesn't exist
status=$(curl -s -o /dev/null -w "%{http_code}" "$OS_URL/$INDEX")
if [ "$status" = "200" ]; then
  echo "Index '$INDEX' already exists, skipping."
else
  echo "Creating index '$INDEX'..."
  curl -s -f -X PUT "$OS_URL/$INDEX" \
    -H "Content-Type: application/json" \
    --data-binary @/deploy/opensearch/index-products.json
  echo "Index '$INDEX' created."
fi

# 2. Register Debezium connector if it doesn't exist
if curl -s -f -o /dev/null "$CONNECT_URL/connectors/$CONNECTOR_NAME"; then
  echo "Connector '$CONNECTOR_NAME' already exists, skipping."
else
  echo "Registering connector '$CONNECTOR_NAME'..."
  curl -s -f -X POST "$CONNECT_URL/connectors" \
    -H "Content-Type: application/json" \
    --data-binary @/deploy/debezium/pg-connector.json
  echo "Connector '$CONNECTOR_NAME' registered."
fi

echo "Bootstrap complete."
