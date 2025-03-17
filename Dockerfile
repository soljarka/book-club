FROM golang:1.23

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/ .

CMD ["./out/book-club"]
