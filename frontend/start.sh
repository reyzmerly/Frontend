#!/bin/bash
cd "$(dirname "$0")"
PORT=8080 node .output/server/index.mjs
