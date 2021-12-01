FROM golang

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /workspace

COPY go.* ./

RUN go mod download

# the `cmd/buildhack` package references a bunch of dependencies that we want
# to precompile for caching purposes so the final `go build` only builds the
# packages in this repo.
COPY cmd/buildhack/ cmd/buildhack/
RUN CGO_ENABLED=0 go build -v -ldflags '-s' -o /tmp/buildhack ./cmd/buildhack/

COPY *.go ./

COPY ./pkg/ ./pkg/

RUN CGO_ENABLED=0 go build -v -ldflags '-s' -o /bin/comments

FROM scratch

COPY --from=0 /etc/passwd /etc/passwd

COPY --from=0 /bin/comments /bin/comments

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER nobody

CMD ["/bin/comments"]
