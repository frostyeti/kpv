FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY kpv /usr/local/bin/kpv
ENTRYPOINT ["/usr/local/bin/kpv"]
