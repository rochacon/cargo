# Cargo Dockerfile
FROM ubuntu:14.04
RUN apt-get update -yq && apt-get install -yq bzr git nginx-full wget && apt-get clean

# ENV AWS_ACCESS_KEY_ID ""
# ENV AWS_SECRET_ACCESS_KEY ""
# ENV BASE_DOMAIN ""
# ENV BUCKET ""
# ENV DOCKER_HOSTS ""
# ENV S3_ENDPOINT ""

# Install Go
RUN wget -qO /tmp/golang.tar.gz https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz \
    && test $(sha1sum /tmp/golang.tar.gz | awk '{print $1}') = "14068fbe349db34b838853a7878621bbd2b24646" \
    && tar -C /usr/local -xzf /tmp/golang.tar.gz \
    && rm /tmp/golang.tar.gz
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /go

# Get gitreceived
RUN go get -v github.com/flynn/flynn/gitreceived && go clean

# Build cargo
ADD . /go/src/github.com/rochacon/cargo
RUN cd /go/src/github.com/rochacon/cargo && go get -v ./... && go clean
RUN mkdir -p /etc/cargo/repositories

EXPOSE 22
EXPOSE 80
CMD "/go/src/github.com/rochacon/cargo/run.sh"
