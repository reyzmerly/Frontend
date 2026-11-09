#!/bin/bash
cd "$(dirname "$0")"
PORT=5173 node .output/server/index.mjs
