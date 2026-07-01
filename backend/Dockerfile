# FAIRRIDE — Multi-stage service Dockerfile
# Build context must be the backend/ directory.
# Usage:
#   docker build \
#     --build-arg SERVICE=identity \
#     --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#     -t fairride/identity:$(git rev-parse --short HEAD) .

ARG SERVICE=identity
ARG GO_VERSION=1.22
ARG GIT_COMMIT=unknown

# ─── Stage 1: Builder ────────────────────────────────────────────────────────
FROM golang:${GO_VERSION}-alpine AS builder

ARG SERVICE
ARG GIT_COMMIT

RUN apk --no-cache add git ca-certificates tzdata

WORKDIR /workspace

# Copy workspace and module files first to maximise Docker layer cache hits.
COPY go.work ./
COPY shared/go.mod shared/go.sum ./shared/
COPY services/${SERVICE}/go.mod ./services/${SERVICE}/

# Download dependencies before copying source.
RUN cd shared && go mod download

# Copy full source.
COPY shared/ ./shared/
COPY services/${SERVICE}/ ./services/${SERVICE}/

# Build a fully static binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w -X main.buildVersion=${GIT_COMMIT}" \
      -o /bin/server \
      ./services/${SERVICE}/cmd/server

# ─── Stage 2: Runtime ────────────────────────────────────────────────────────
FROM scratch AS runtime

ARG SERVICE

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bin/server /bin/server

# gRPC port (default :50051), HTTP health port (default :8080)
EXPOSE 50051 8080

LABEL org.opencontainers.image.title="fairride-${SERVICE}"
LABEL org.opencontainers.image.source="https://github.com/fairride/fairride"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"

ENTRYPOINT ["/bin/server"]
