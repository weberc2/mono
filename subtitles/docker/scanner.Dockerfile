FROM golang

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/app ./cmd/scanner

FROM alpine

COPY --from=0 /bin/app /bin/app

CMD ["/bin/app", "."]