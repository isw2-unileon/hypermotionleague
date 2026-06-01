# syntax=docker/dockerfile:1
#
# Production image for the HyperMotion League Go backend (Render / Docker).
#
# IMPORTANT — build context must be the REPO ROOT, not ./backend.
# This monorepo is a single Go module rooted at the repo root
# (module github.com/isw2-unileon/proyect-scaffolding); go.mod/go.sum live
# there, while the source lives under ./backend. Build with:
#
#     docker build -t backend-test -f backend/Dockerfile .
#
# On Render: set "Dockerfile Path" = backend/Dockerfile and leave the
# build/root context at the repository root.

# ---- builder ----------------------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache dependencies: copy only the module files first.
COPY go.mod go.sum ./
RUN go mod download

# Copy the backend source (the only Go code in the module).
COPY backend ./backend

# Static binary: no cgo so it runs on a bare Alpine runtime.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./backend/cmd/server

# ---- runtime ----------------------------------------------------------------
FROM alpine:3.20

# TLS roots so outbound HTTPS (Supabase / Postgres SSL) works.
RUN apk add --no-cache ca-certificates

# Run as an unprivileged user.
RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /server /server

# Documentation only — the app binds to $PORT (fallback 8080) at runtime,
# and Render injects PORT.
EXPOSE 8080

CMD ["/server"]
