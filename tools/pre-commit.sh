#!/bin/sh

docker run --rm -v "$(pwd):/app" go-dev-tools:latest -c 'task lint'
