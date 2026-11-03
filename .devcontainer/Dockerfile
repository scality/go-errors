FROM mcr.microsoft.com/devcontainers/go:1.26-trixie@sha256:52c89857fb443d53b896918a3b2ed968801f0b667ad98e56e337662cfb9d6ba6

USER vscode

# Same version as the CI lint job.
# renovate: datasource=github-releases depName=golangci/golangci-lint extractVersion=^v(?<version>.+)$
ARG GOLANGCI_VERSION=2.15.0
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sudo sh -s -- v${GOLANGCI_VERSION}
