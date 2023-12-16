FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
COPY templates /app/templates
COPY assets /app/assets

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
