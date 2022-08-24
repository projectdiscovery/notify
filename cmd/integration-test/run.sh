#!/bin/bash

rm integration-test notify 2>/dev/null

go build ../notify
go build

./integration-test
if [ $? -eq 0 ]
then
  exit 0
else
  exit 1
fi
