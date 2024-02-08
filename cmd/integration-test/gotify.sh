#!/bin/bash

# If script running in local, then setup server with docker
# docker run -p 80:80 -v /var/gotify/data:/app/data ghcr.io/gotify/server

# Configuration variables
GOTIFY_URL="http://localhost"
AUTH_HEADER="Authorization: Basic YWRtaW46YWRtaW4=" # default username and password is admin:admin
APPLICATION_JSON='{
  "defaultPriority": 5,
  "description": "Gotify Test server",
  "name": "test-server"
}'

# Create a new application and get the response
APPLICATION_RESPONSE=$(curl --location "$GOTIFY_URL/application" \
--header "Content-Type: application/json" \
--header "$AUTH_HEADER" \
--data "$APPLICATION_JSON")

# Extract application token from the response
APP_TOKEN=$(echo "$APPLICATION_RESPONSE" | jq -r '.token')

if [ "$APP_TOKEN" == "null" ]; then
    echo "Failed to create application"
    exit 1
fi

echo "Gotify Application created successfully. Token: $APP_TOKEN"

# write the token to a file
echo "$APP_TOKEN" > gotify-app-token.txt
