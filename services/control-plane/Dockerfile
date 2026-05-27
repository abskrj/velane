FROM golang:1.22-alpine AS builder
WORKDIR /app

# Install git so go mod can fetch private modules if needed.
RUN apk add --no-cache git

COPY go.mod go.sum ./
# Download dependencies using the exact versions in go.sum.
# If go.sum is incomplete, run "cd services/control-plane && go mod tidy" locally first.
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /runeforge ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /runeforge /runeforge
EXPOSE 8080
ENTRYPOINT ["/runeforge"]
