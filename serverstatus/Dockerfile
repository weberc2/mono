FROM golang

WORKDIR /work

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -v -o /app

FROM gcr.io/distroless/static-debian12
COPY --from=0 /app /app

CMD ["/app"]