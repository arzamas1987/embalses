# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# API binary
FROM builder AS api-build
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

# MCP binary
FROM builder AS mcp-build
RUN CGO_ENABLED=0 go build -o /bin/mcp ./cmd/mcp

# Ingest binary
FROM builder AS ingest-build
RUN CGO_ENABLED=0 go build -o /bin/ingest ./cmd/ingest

# Admin binary
FROM builder AS admin-build
RUN CGO_ENABLED=0 go build -o /bin/admin ./cmd/admin

# Runtime targets
FROM gcr.io/distroless/static-debian12:nonroot AS api
COPY --from=api-build /bin/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]

FROM gcr.io/distroless/static-debian12:nonroot AS mcp
COPY --from=mcp-build /bin/mcp /mcp
EXPOSE 8081
ENTRYPOINT ["/mcp"]
