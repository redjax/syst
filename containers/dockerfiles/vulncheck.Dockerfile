ARG GOLANG_IMG_VERSION=1.20
FROM golang:${GOLANG_IMG_VERSION}-trixie AS base

ARG GOLANG_IMG_VERSION

ENV DEBIAN_FRONTEND=noninteractive

# Install curl and unzip via apt, then remove apt cache to keep image small
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Download official osv-scanner binary and make it executable
RUN curl -L -o /usr/local/bin/osv-scanner https://github.com/google/osv-scanner/releases/latest/download/osv-scanner-linux-amd64 \
    && chmod +x /usr/local/bin/osv-scanner

FROM base AS deps

WORKDIR /tmp/deps

# Copy only go.mod and go.sum first to leverage Docker cache for dependencies install
COPY go.mod go.sum ./

# Install Go vulnerability tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install golang.org/x/vuln/cmd/govulncheck@latest

FROM base AS scanner

WORKDIR /repo

# Copy entire source tree
COPY . /repo

# Add Go tools directory to PATH
ENV PATH="/go/bin:${PATH}"

# Entrypoint command to run your scanning script and output to /scans
CMD ["bash", "/repo/scripts/vuln-scan.sh", "--output-dir", "/scans"]
