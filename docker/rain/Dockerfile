FROM golang:1.22

WORKDIR /workspace

RUN go install github.com/cenkalti/rain@2.2.0 && cp $GOPATH/bin/rain /bin/rain

FROM tailscale/tailscale:latest

COPY --from=0 /bin/rain /bin/rain

CMD ["/bin/rain", "server"]
