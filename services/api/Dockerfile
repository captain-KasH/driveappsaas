# Build context: repository root (see deploy/compose.yml).
FROM node:22-alpine AS web
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /web
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY apps/web/ ./
RUN npx next build

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download
COPY services/api/ ./
COPY --from=web /web/out/ ./internal/web/dist/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /taxios-api ./cmd/taxios-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /taxios-api /taxios-api
USER nonroot
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD ["/taxios-api", "healthcheck"]
ENTRYPOINT ["/taxios-api"]
CMD ["serve"]
