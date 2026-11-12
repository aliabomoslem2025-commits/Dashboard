FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mx-core ./cmd/server
FROM alpine:3.21
RUN adduser -D -H -u 10001 app
COPY --from=build /out/mx-core /mx-core
USER 10001
EXPOSE 8081
ENTRYPOINT ["/mx-core"]
