FROM golang:1.9 as builder
COPY . /go/src/github.com/newrelic/nri-nginx/
RUN cd /go/src/github.com/newrelic/nri-nginx && \
    make && \
    strip ./bin/nri-nginx

FROM newrelic/infrastructure:latest
ENV NRIA_IS_FORWARD_ONLY true
ENV NRIA_K8S_INTEGRATION true
COPY --from=builder /go/src/github.com/newrelic/nri-nginx/bin/nri-nginx /nri-sidecar/newrelic-infra/newrelic-integrations/bin/nri-nginx
COPY --from=builder /go/src/github.com/newrelic/nri-nginx/nginx-definition.yml /nri-sidecar/newrelic-infra/newrelic-integrations/definition.yml
USER 1000
