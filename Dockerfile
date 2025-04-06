FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o displateBot .

FROM alpine

WORKDIR /app
COPY --from=builder /app/displateBot .

ENTRYPOINT ["./displateBot"]