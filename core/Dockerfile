FROM golang:1.22-alpine AS build
WORKDIR /src
# The Go toolchain and dependency resolution stay inside this build container.
# Generate go.sum from the checked-in go.mod and source imports when a checkout
# does not contain go.sum, then compile without requiring Go on the host.
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mx-core ./cmd/server
FROM alpine:3.21
RUN adduser -D -H -u 10001 app
COPY --from=build /out/mx-core /mx-core
USER 10001
EXPOSE 8081
ENTRYPOINT ["/mx-core"]
