# API image. Build context is the repository root (see docker-compose.yml).
# The fixtures are not copied in: compose mounts ./fixtures read-only at /data.
FROM golang:1.27.2-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY backend/ ./
RUN go build -trimpath -ldflags="-s -w" -o /out/sessions ./cmd/sessions

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sessions /sessions
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/sessions"]
# --bind-any lets nginx reach the API on the compose network; the API still accepts only a
# loopback Host header, which nginx sets after its own Host and Origin checks.
CMD ["serve", "--input", "/data/claude-code", "--labels", "/data/labels/heldout.json", "--listen", "0.0.0.0:8080", "--bind-any"]
