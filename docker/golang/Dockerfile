# syntax=docker/dockerfile:1.4
FROM golang:1.25

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./

ENV GOMODCACHE=/go-cache/mod
ENV GOCACHE=/go-cache/build
RUN --mount=type=bind,target=/go-cache \
    go mod download

COPY *.go ./

COPY pkg pkg

ARG TARGET

COPY cmd/${TARGET} cmd/${TARGET}

RUN --mount=type=bind,target=/go-cache \
    CGO_ENABLED=0 go build -ldflags '-s' -v -o /bin/app ./cmd/${TARGET}

FROM scratch

COPY --from=0 /etc/passwd /etc/passwd

COPY --from=0 /bin/app /bin/app

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER nobody

CMD ["/bin/app"]
