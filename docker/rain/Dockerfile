FROM golang:1.22

WORKDIR /workspace

RUN go install github.com/cenkalti/rain@latest && cp $GOPATH/bin/rain /bin/rain

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates curl

USER nobody

COPY --from=0 /bin/rain /bin/rain

CMD ["/bin/rain", "server"]
