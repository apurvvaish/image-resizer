# Image Resizer (Go + Gin)

Resize uploaded images into multiple widths in parallel. Returns URLs to access original + resized variants.

## Run locally

```bash
go mod tidy
go run ./cmd
# or
PORT=8080 UPLOAD_DIR=./uploads SIZES=100,500,1000 go run ./cmd
