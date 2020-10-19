FROM golang:1.15.2-buster

RUN apt-get update \
    && go get golang.org/dl/go1.9.7 \
    && /go/bin/go1.9.7 download

# install Snyk
RUN apt install -y nodejs npm
RUN npm install -g snyk