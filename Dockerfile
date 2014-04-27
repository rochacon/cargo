# Cargo Dockerfile
FROM ubuntu:14.04
RUN apt-get update -yq
RUN apt-get install -yq git bzr nginx-full

# ENV AWS_ACCESS_KEY_ID ""
# ENV AWS_SECRET_ACCESS_KEY ""
# ENV BASE_DOMAIN ""
# ENV BUCKET ""
# ENV DOCKER_HOSTS ""
# ENV S3_ENDPOINT ""

ADD gitreceived /usr/local/bin/gitreceived
ADD cargo /usr/local/bin/cargo
RUN mkdir -p /etc/cargo/{keys,repositories}
# ADD cargo.json /etc/cargo/config.json

ADD run.sh /run.sh
EXPOSE 22
EXPOSE 80
CMD "/run.sh"
