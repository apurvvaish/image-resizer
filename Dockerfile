# --- build stage ---
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o image-resizer ./cmd

# --- runtime stage ---
FROM gcr.io/distroless/base-debian12
WORKDIR /app
ENV PORT=8080 \
    UPLOAD_DIR=/app/uploads \
    SIZES=100,500,1000 \
    MAX_UPLOAD_MB=10
COPY --from=builder /app/image-resizer /app/image-resizer
# create uploads dir
USER 65532:65532
ENTRYPOINT ["/app/image-resizer"]
