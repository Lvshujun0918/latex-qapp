# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM ubuntu:24.04
WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive

# Base runtime dependencies.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      ca-certificates \
      tzdata \
      fontconfig \
      libsqlite3-0 \
      dumb-init \
      curl \
    && rm -rf /var/lib/apt/lists/*

# Install custom fonts before TeX Live.
COPY font/ /usr/local/share/fonts/custom/
RUN fc-cache -f -v

# TeX Live runtime for xelatex rendering.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      texlive-xetex \
      texlive-latex-extra \
      texlive-lang-chinese \
      texlive-science \
      texlive-fonts-recommended \
    && rm -rf /var/lib/apt/lists/*

# Create unprivileged runtime user.
RUN useradd --system --uid 10001 --create-home --shell /usr/sbin/nologin appuser

COPY --from=builder /out/server /app/server
COPY template.tex /app/template.tex

RUN mkdir -p /app/public/pdfs /app/data \
    && chown -R appuser:appuser /app \
    && chmod +x /app/server

ENV TEMPLATE_TEX_PATH=/app/template.tex \
    PDF_OUTPUT_DIR=/app/public/pdfs \
    DATABASE_DSN=file:/app/data/app.db?cache=shared&mode=rwc \
    GIN_MODE=release \
    PORT=8080

VOLUME ["/app/data", "/app/public"]

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD curl -fsS http://127.0.0.1:${PORT}/health || exit 1

USER appuser
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/app/server"]
