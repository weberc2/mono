FROM golang

WORKDIR /work

COPY . .

ENV GOCACHE=/root/.cache/go-build
ENV CGO_ENABLED=0
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o /dedup ./cmd/dedup

FROM gcr.io/distroless/static-debian12:debug
COPY --from=0 /dedup /bin/dedup
CMD ["/bin/dedup"]