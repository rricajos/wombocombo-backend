FROM golang:1.22-alpine AS build

WORKDIR /app

# Copy everything first, then resolve deps and build
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api-server main.go

# Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget

COPY --from=build /api-server /api-server

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:3000/api/health || exit 1

ENTRYPOINT ["/api-server"]