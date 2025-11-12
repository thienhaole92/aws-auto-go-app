# syntax=docker/dockerfile:1.4

ARG GO_VERSION=1.25.1
ARG ALPINE_VERSION=3.22

########################
# --- Build Stage --- #
########################
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Install build dependencies
RUN apk add --no-cache upx

WORKDIR /src

# Pre-copy go mod files for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Optional: Set build-time variables
ARG TARGETOS
ARG TARGETARCH

# Build the Go application
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-s -w" \
    -trimpath \
    -o aws-auto-go-app && \
    upx --lzma -q aws-auto-go-app

########################
# --- Final Stage --- #
########################
FROM gcr.io/distroless/static:latest AS final

ARG COMMIT_SHA=unknown

# Environment variables
ENV COMMIT_SHA=${COMMIT_SHA}

# Set working directory
WORKDIR /app

# Copy the compiled binary
COPY --from=builder --chown=nonroot:nonroot /src/aws-auto-go-app /app/aws-auto-go-app
COPY --chown=nonroot:nonroot db/migrations/ /app/db/migrations/

# Use ENTRYPOINT for better shell signal handling
ENTRYPOINT ["/app/aws-auto-go-app"]
