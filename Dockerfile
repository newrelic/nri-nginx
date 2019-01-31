FROM golang:1.9 as builder
RUN go get -d github.com/newrelic/nri-nginx/... && \
    cd /go/src/github.com/newrelic/nri-nginx && \
    make && \
    strip ./bin/nr-nginx

FROM newrelic/infrastructure:latest
COPY --from=builder /go/src/github.com/newrelic/nri-nginx/bin/nr-nginx /var/db/newrelic-infra/newrelic-integrations/bin/nr-nginx
COPY --from=builder /go/src/github.com/newrelic/nri-nginx/nginx-definition.yml /var/db/newrelic-infra/newrelic-integrations/definition.yml
