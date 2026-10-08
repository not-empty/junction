# Production image: builds the API and the migrate tool (migrations are embedded).
# Runs on the build machine's own arch and cross-compiles for the target, so
# building for another arch needs no emulation.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
# Cache mounts keep modules and compiled packages between builds, so only what
# changed is recompiled. One build for both binaries shares the dependencies.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/api ./cmd/migrate

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/api /out/migrate /

USER 10001:10001
EXPOSE 8080

ENTRYPOINT ["/api"]
