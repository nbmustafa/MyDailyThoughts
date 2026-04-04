# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/k8sdiag \
    ./cmd/k8sdiag

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN addgroup -S k8sdiag && adduser -S k8sdiag -G k8sdiag

COPY --from=builder /bin/k8sdiag /usr/local/bin/k8sdiag

USER k8sdiag

ENTRYPOINT ["k8sdiag"]
CMD ["--help"]
