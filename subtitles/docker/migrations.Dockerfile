FROM golang AS builder

WORKDIR /work

COPY go.mod go.sum tools.go ./

RUN CGO_ENABLED=0 go build -tags postgres -o /bin/migrate github.com/golang-migrate/migrate/v4/cmd/migrate

FROM gcr.io/distroless/static-debian12

WORKDIR /work

COPY --from=builder /bin/migrate /bin/migrate

COPY db/migrations ./db/migrations