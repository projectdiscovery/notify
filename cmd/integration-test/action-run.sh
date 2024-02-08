#!/bin/bash

# setup gotify server
bash gotify.sh
export $(echo "GOTIFY_APP_TOKEN=$(<gotify-app-token.txt)")

rm -f final-config.yaml temp.yaml gotify-app-token.txt
( echo "cat <<EOF >final-config.yaml";
  cat test-config.yaml;
  echo "EOF";
) >temp.yaml
. temp.yaml
rm integration-test notify 2>/dev/null

go build ../notify
go build

DEBUG=true ./integration-test --provider-config final-config.yaml
if [ $? -eq 0 ]
then
  exit 0
else
  exit 1
fi
