#!/usr/bin/env bash

set -e

pushd frontend
npm install
npm run build
popd
