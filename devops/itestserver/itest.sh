#!/bin/bash

export GOPATH=$HOME/go
export PATH=$PATH:$HOME/go/bin
export GITHUB_TOKEN="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"
export GOOGLE_CLOUD_PROJECT="senecacam-staging"
export GOOGLE_APPLICATION_CREDENTIALS="/home/lucaloncar/credentials/app.json"
export GOOGLE_OAUTH_CREDENTIALS="/home/lucaloncar/credentials/oauth.json"

echo "Cloning seneca repo"
git clone -b staging https://${GITHUB_TOKEN}@github.com/Seneca-AI/seneca.git || exit 1
echo "Cloning common repo"
git clone https://${GITHUB_TOKEN}@github.com/Seneca-AI/common.git || exit 1
echo "Copying protos"
cp -r common/proto_out/go/api seneca || exit 1

echo "Running integration test"
cd seneca/test/integrationtest
sudo env "PATH=$PATH" "GOPATH=$GOPATH" "GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT" "GOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS" "GOOGLE_OAUTH_CREDENTIALS=$GOOGLE_OAUTH_CREDENTIALS" go run . || exit 1
cd -

