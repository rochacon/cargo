#!/bin/bash
# This script will setup your server from scratch
set -e

# FIXME configure these variables
BASE_DOMAIN="localhost"
BUCKET="my-slug-bucket-container" # XXX right now slugs will be public, remember, pre-alpha software

# Setup Docker
curl -sL http://get.docker.io/ | bash
docker pull flynn/slugbuilder
docker pull flynn/slugrunner

# Setup gitreceived
curl -sL https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz | tar xzC /usr/local

export PATH=/usr/local/go/bin:$PATH
export GOROOT=/usr/local/go
export GOPATH=/go

apt-get install -yq git
go get -v github.com/flynn/gitreceive-next/gitreceived
go get -v github.com/rochacon/cargo

useradd -G docker -m -s /bin/bash git || true
su -c "mkdir -p /home/git/.ssh && ssh-keygen -t rsa -f /home/git/.ssh/id_rsa -N ''" git
su -c "mkdir -p /home/git/{keys,hosts,repositories}" git

cat >/etc/init/gitreceived.conf <<EOF
start on runlevel [2345]
stop on starting rc RUNLEVEL=[016]

setuid git
setgid docker

respawn

console log
chdir /home/git

exec /go/bin/gitreceived -p 2222 -k /home/git/keys -r /home/git/repositories /home/git/.ssh/id_rsa "/go/bin/cargo -bucket $BUCKET -d $BASE_DOMAIN"
EOF

start gitreceived

apt-get install -y nginx
cat >/etc/nginx/nginx.conf <<EOF
daemon on;
error_log error.log;
pid /var/run/nginx/nginx.pid;

events {
    use epoll;
    worker_connections 4096;
}

http {
    server_names_hash_bucket_size 256;
    include /home/git/hosts/*.conf;
}
EOF
nginx -s reload
