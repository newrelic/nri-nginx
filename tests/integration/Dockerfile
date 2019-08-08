FROM golang:1.9 as builder
ARG CGO_ENABLED=0
WORKDIR /go/src/github.com/newrelic/nri-nginx
COPY . .
RUN make clean compile

FROM alpine:latest
COPY --from=builder /go/src/github.com/newrelic/nri-nginx/bin /
CMD ["sleep", "1h"]