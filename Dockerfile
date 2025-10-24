# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/app ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /root
COPY --from=builder /app/build/app ./app
EXPOSE 8080
ENTRYPOINT ["./app"]
