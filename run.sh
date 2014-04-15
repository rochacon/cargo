#!/bin/bash
set -e
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-""}
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-""}
BUCKET=${BUCKET:-""}
S3_ENDPOINT=${S3_ENDPOINT:-""}

CARGO_BASE_DOMAIN=${BASE_DOMAIN:-"localhost"}
CARGO_PRIVATE_KEY=${CARGO_PRIVATE_KEY:-""}
CARGO_PRIVATE_KEY_FILENAME="/etc/cargo/private_key"

DOCKER_HOSTS=${DOCKER_HOSTS:-"http://127.0.0.1:4243"}

if [ "$CARGO_PRIVATE_KEY" != "" ]; then
    echo "$CARGO_PRIVATE_KEY" > "$CARGO_PRIVATE_KEY_FILENAME"
else
    echo "Generating random private key"
    ssh-keygen -t rsa -f $CARGO_PRIVATE_KEY_FILENAME -N ''
fi

/usr/local/bin/gitreceived \
    -p 22 \
    -k /etc/cargo/keys \
    -r /etc/cargo/repositories \
    "$CARGO_PRIVATE_KEY_FILENAME" \
    "/usr/local/bin/cargo -bucket '$BUCKET' -domain '$CARGO_BASE_DOMAIN' -aws-key '$AWS_ACCESS_KEY_ID' -aws-secret '$AWS_SECRET_ACCESS_KEY' -dockers '$DOCKER_HOSTS'"
    # "/usr/local/bin/cargo -config /etc/cargo.json"