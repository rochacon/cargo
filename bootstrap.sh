#!/bin/bash
# This script will setup your server from scratch
set -e

# FIXME configure these variables
AWS_ACCESS_KEY_ID=""
AWS_SECRET_ACCESS_KEY=""
BASE_DOMAIN="localhost"
BUCKET=""

# Setup Docker
curl -sL http://get.docker.io/ | bash
docker pull flynn/slugbuilder
docker pull flynn/slugrunner
echo 'DOCKER_OPTS="-H tcp://127.0.0.1:4243 -H unix:///var/run/docker.sock"' > /etc/default/docker
restart docker

# Setup Cargo
curl -sL https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz | tar xzC /usr/local

export PATH=/usr/local/go/bin:$PATH
export GOROOT=/usr/local/go
export GOPATH=/go

apt-get install -yq git bzr
go get -v github.com/flynn/gitreceive-next/gitreceived
go get -v github.com/rochacon/cargo

useradd -G docker -m -s /bin/bash git || true
echo 'git ALL=(ALL:ALL) NOPASSWD: /usr/sbin/nginx' > /etc/sudoers.d/nginx
su -c "mkdir -p /home/git/.ssh && ssh-keygen -t rsa -f /home/git/.ssh/id_rsa -N ''" git
su -c "mkdir -p /home/git/{keys,hosts,repositories}" git

cat >/etc/init/cargo.conf <<EOF
start on runlevel [2345]
stop on starting rc RUNLEVEL=[016]

setuid git
setgid docker

respawn

console log
chdir /home/git

exec /go/bin/gitreceived \\
        -p 2222 \\
        -k /home/git/keys \\
        -r /home/git/repositories \\
        /home/git/.ssh/id_rsa \\
        "/go/bin/cargo -bucket $BUCKET -d $BASE_DOMAIN -aws-key $AWS_ACCESS_KEY_ID -aws-secret $AWS_SECRET_ACCESS_KEY"
EOF
start cargo

# Setup NGINX
apt-get install -y nginx
cat >/etc/nginx/nginx.conf <<EOF
daemon on;
error_log error.log;
pid /var/run/nginx.pid;

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
