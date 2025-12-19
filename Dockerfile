FROM golang:1.25.5 AS builder

ARG TARGETARCH
ARG TARGETOS

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-X main.version=$(git describe --abbrev=0 --tags || echo dev)" \
    -o clispot

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    ffmpeg \
    python3 \
    python3-pip \
 && pip3 install --break-system-packages yt-dlp \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/clispot /usr/bin/clispot

RUN mkdir -p /root/.cache/clispot /root/.clispot 

ENTRYPOINT ["/usr/bin/clispot"]
