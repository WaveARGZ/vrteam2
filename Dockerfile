FROM golang:1.20

RUN apt-get update && apt-get install -y gcc

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o server main.go

CMD ["./server"]
