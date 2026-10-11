FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /taxios-api ./cmd/taxios-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /taxios-api /taxios-api
USER nonroot
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD ["/taxios-api", "healthcheck"]
ENTRYPOINT ["/taxios-api"]
CMD ["serve"]
