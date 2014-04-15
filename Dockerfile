# Cargo Dockerfile
FROM dockerfile/nginx

# ENV AWS_ACCESS_KEY_ID ""
# ENV AWS_SECRET_ACCESS_KEY ""
# ENV BASE_DOMAIN ""
# ENV BUCKET ""
# ENV DOCKER_HOSTS ""

ADD gitreceived /usr/local/bin/gitreceived
ADD cargo /usr/local/bin/cargo
# VOLUME /etc/cargo/keys
RUN mkdir -p /etc/cargo/repositories

ADD run.sh /run
EXPOSE 22
ENTRYPOINT ["/run"]
