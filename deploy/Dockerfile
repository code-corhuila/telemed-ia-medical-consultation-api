# syntax=docker/dockerfile:1.7

# ---------- build stage ----------
FROM golang:1.23-alpine AS build
WORKDIR /workspace

COPY go.mod go.sum* ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/consultation-api

# ---------- runtime stage ----------
FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/app /app/app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
