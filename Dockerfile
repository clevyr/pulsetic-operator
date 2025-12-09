#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache \
  GOARCH="$TARGETARCH" CGO_ENABLED=0 go build -ldflags='-w -s' -tags grpcnotrace -trimpath -o manager cmd/main.go


FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/manager /

ENTRYPOINT ["/manager"]
