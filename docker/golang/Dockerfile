FROM golang:1.23

WORKDIR /workspace

COPY go.mod go.sum ./

COPY ./pkg ./pkg

ARG TARGET

COPY ./cmd/${TARGET} ./cmd/${TARGET}

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" \
    --mount=type=cache,target="/go/pkg/mod" \
    CGO_ENABLED=0 go build -ldflags '-s' -v -o /bin/app ./cmd/${TARGET}

# `debug-nonroot` variant contains a shell and weighs about 3.3MB vs 1.9MB for
# the `nonroot`.
FROM gcr.io/distroless/static-debian12:debug-nonroot

COPY --from=0 /bin/app /bin/app

USER nobody

CMD ["/bin/app"]
