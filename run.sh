#!/bin/bash

docker run --rm -it \
	-w /app \
	-v "${PWD}":/app \
	--network firefly-iii_firefly_iii \
	-p 4000:4000 \
	--name firefly_webhooks \
	-e FIREFLY_BASE_URL="${FIREFLY_BASE_URL}" \
	-e FIREFLY_API_KEY="${FIREFLY_API_KEY}" \
	air \
	air -c .air.toml
