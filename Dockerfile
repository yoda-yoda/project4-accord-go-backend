# Step 1: 빌드 스테이지
FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o go-server .

# Step 2: 런타임 스테이지
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/go-server .

EXPOSE 4000
ENTRYPOINT ["./go-server"]
