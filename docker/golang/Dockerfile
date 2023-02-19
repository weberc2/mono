FROM golang:1.18.3

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

COPY ./pkg ./pkg

ARG TARGET

COPY ./cmd/${TARGET} ./cmd/${TARGET}

RUN CGO_ENABLED=0 go build -ldflags '-s' -v -o /bin/app ./cmd/${TARGET}

FROM scratch

COPY --from=0 /etc/passwd /etc/passwd

COPY --from=0 /bin/app /bin/app

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER nobody

CMD ["/bin/app"]
