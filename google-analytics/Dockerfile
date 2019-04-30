FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/google-analytics

ADD main.go .
ADD analytics.go .
ADD vendor vendor

RUN go build -o /usr/bin/google-analytics .

FROM alpine:3.9

RUN apk add --no-cache ca-certificates

# Setting the group prevented access to /tmp at runtime
# lchown started to fail
# -G app 
RUN addgroup -S app && adduser app -S \
 && mkdir -p /home/app

WORKDIR /home/app

COPY --from=builder /usr/bin/google-analytics /home/app/

RUN chown -R app /home/app
USER app

EXPOSE 8080
VOLUME /tmp/

ENTRYPOINT ["/home/app/google-analytics"]
