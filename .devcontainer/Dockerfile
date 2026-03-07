FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

ARG GO_VERSION="1.26.0"
ARG NODE_VERSION="20.19.0"
ARG NPM_VERSION="10.8.2"
ARG ANGULAR_CLI_VERSION="21.1.0"
ARG GOLANGCI_LINT_VERSION="v2.10.1"
ARG GOFUMPT_VERSION="v0.9.2"
ARG GOPLS_VERSION="v0.21.1"
ARG DLV_VERSION="v1.26.0"

RUN apt-get update \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        ca-certificates \
        curl \
        git \
        gnupg \
        build-essential \
        postgresql-client \
        xz-utils \
    && ARCH="$(dpkg --print-architecture)" \
    && case "${ARCH}" in \
        amd64) GO_ARCH="amd64"; NODE_ARCH="x64" ;; \
        arm64) GO_ARCH="arm64"; NODE_ARCH="arm64" ;; \
        *) echo "Unsupported architecture: ${ARCH}" && exit 1 ;; \
    esac \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm -f /tmp/go.tar.gz \
    && curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-${NODE_ARCH}.tar.xz" -o /tmp/node.tar.xz \
    && rm -f /usr/local/bin/node /usr/local/bin/npm /usr/local/bin/npx \
    && tar -xJf /tmp/node.tar.xz -C /usr/local --strip-components=1 \
    && rm -f /tmp/node.tar.xz \
    && npm install -g "npm@${NPM_VERSION}" "@angular/cli@${ANGULAR_CLI_VERSION}" \
    && curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/${GOLANGCI_LINT_VERSION}/install.sh" \
        | sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}" \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV GOPATH=/home/vscode/go
ENV PATH=/usr/local/go/bin:${GOPATH}/bin:${PATH}

USER vscode

RUN CGO_ENABLED=0 go install mvdan.cc/gofumpt@${GOFUMPT_VERSION} \
    && CGO_ENABLED=0 go install golang.org/x/tools/gopls@${GOPLS_VERSION} \
    && CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@${DLV_VERSION}
