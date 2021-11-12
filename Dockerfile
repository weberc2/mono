FROM golang

WORKDIR /workspace

COPY go.* ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -o /bin/comments

CMD /bin/comments
