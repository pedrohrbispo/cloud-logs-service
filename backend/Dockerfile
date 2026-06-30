# syntax=docker/dockerfile:1

# ---- build stage ----------------------------------------------------------
# Go 1.25: required by aws-sdk-go-v2 v1.42+ (see go.mod `go` directive).
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache module downloads separately from source.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH

RUN go build -trimpath -ldflags="-s -w" -o /out/server      ./cmd/server
RUN go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck
RUN go build -trimpath -ldflags="-s -w" -o /out/seed        ./cmd/seed

# ---- server image (default) ----------------------------------------------
# distroless static: no shell, no package manager, runs as nonroot.
FROM gcr.io/distroless/static-debian12:nonroot AS server
WORKDIR /
COPY --from=build /out/server      /server
COPY --from=build /out/healthcheck /healthcheck
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=4s --start-period=5s --retries=3 \
  CMD ["/healthcheck"]
ENTRYPOINT ["/server"]

# ---- seed image (one-shot init job) --------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS seed
WORKDIR /
COPY --from=build /out/seed /seed
USER nonroot:nonroot
ENTRYPOINT ["/seed"]
