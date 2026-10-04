FROM golang:1.26-bookworm AS build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/dedhed ./cmd/dedhed
FROM debian:bookworm-slim
RUN apt-get update && apt-get install --yes --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/dedhed /usr/local/bin/dedhed
COPY migrations/ /app/migrations/
ENV API_ADDR=0.0.0.0:8080 DEDHED_MIGRATIONS=/app/migrations DEDHED_MEDIA=/media
USER 1000:1000
EXPOSE 8080
ENTRYPOINT ["dedhed"]
CMD ["serve"]
