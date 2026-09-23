FROM mcr.microsoft.com/devcontainers/go:1.26-bookworm

# Install golangci-lint via official binary installer (go install takes 5-10 min to compile)
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b /usr/local/bin v1.59.1

# Install Go dev tools
RUN go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest

ENV TZ=Europe/Berlin
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && echo "${TZ}" > /etc/timezone

WORKDIR /workspace
