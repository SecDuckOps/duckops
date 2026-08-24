FROM golang:1.27-bookworm AS server-builder

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /src
COPY go.mod go.sum* ./duck/
WORKDIR /src/duck
RUN go mod download
COPY . /src/duck
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/duckops ./main.go

FROM alpine:3.20 AS go-tool-builder

RUN apk add --no-cache curl unzip

RUN mkdir -p /out && \
    curl -sSfL https://github.com/aquasecurity/tfsec/releases/download/v1.28.14/tfsec-linux-amd64 -o /out/tfsec && chmod +x /out/tfsec

RUN curl -sSfL https://github.com/securego/gosec/releases/download/v2.25.0/gosec_2.25.0_linux_amd64.tar.gz | tar -xzf - -C /out

RUN curl -sSfL https://github.com/trufflesecurity/trufflehog/releases/download/v3.92.3/trufflehog_3.92.3_linux_amd64.tar.gz | tar -xzf - -C /out

RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /out

RUN curl -sSfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /out

RUN curl -sSfL https://github.com/gitleaks/gitleaks/releases/download/v8.24.2/gitleaks_8.24.2_linux_x64.tar.gz | tar -xzf - -C /out gitleaks

RUN curl -sSfL https://github.com/projectdiscovery/subfinder/releases/download/v2.6.5/subfinder_2.6.5_linux_amd64.zip -o /tmp/subfinder.zip && unzip /tmp/subfinder.zip -d /out

RUN curl -sSfL https://github.com/projectdiscovery/httpx/releases/download/v1.6.10/httpx_1.6.10_linux_amd64.zip -o /tmp/httpx.zip && unzip /tmp/httpx.zip -d /out httpx

RUN curl -sSfL https://github.com/projectdiscovery/katana/releases/download/v1.1.2/katana_1.1.2_linux_amd64.zip -o /tmp/katana.zip && unzip /tmp/katana.zip -d /out katana

RUN curl -sSfL https://github.com/projectdiscovery/nuclei/releases/download/v3.4.4/nuclei_3.4.4_linux_amd64.zip -o /tmp/nuclei.zip && unzip /tmp/nuclei.zip -d /out nuclei

FROM debian:bookworm-slim AS cert-builder

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates openssl && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /out && \
    openssl ecparam -name prime256v1 -genkey -noout -out /out/ca.key && \
    openssl req -x509 -new -key /out/ca.key -out /out/ca.crt -days 3650 \
    -subj "/C=US/ST=CA/O=DuckOps/CN=DuckOps Testing Root CA" \
    -addext "basicConstraints=critical,CA:TRUE" \
    -addext "keyUsage=critical,digitalSignature,keyEncipherment,keyCertSign" && \
    openssl pkcs12 -export -out /out/ca.p12 -inkey /out/ca.key -in /out/ca.crt \
    -passout pass:"" -name "DuckOps Testing Root CA" && \
    chmod 644 /out/ca.crt && \
    chmod 600 /out/ca.key /out/ca.p12

FROM debian:bookworm-slim AS extras-builder

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates git nodejs npm python3 python3-venv && \
    python3 -m venv /app/venv && \
    /app/venv/bin/pip install --no-cache-dir arjun bandit checkov packaging semgrep && \
    mkdir -p /home/duckops/.npm-global /home/duckops/tools /opt && \
    npm install --global --prefix /home/duckops/.npm-global eslint retire && \
    git clone --depth 1 https://github.com/docker/docker-bench-security.git /opt/docker-bench-security && \
    git clone --depth 1 https://github.com/ticarpi/jwt_tool.git /home/duckops/tools/jwt_tool && \
    git clone --depth 1 https://github.com/projectdiscovery/nuclei-templates.git /opt/nuclei-templates && \
    rm -rf \
    /opt/docker-bench-security/.git \
    /home/duckops/tools/jwt_tool/.git \
    /opt/nuclei-templates/.git \
    /root/.cache \
    /root/.npm

FROM debian:bookworm-slim

LABEL description="DuckOps sandbox container"

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV DEBIAN_FRONTEND=noninteractive \
    DUCKOPS_CACHE_DIR=/var/lib/duckops \
    DUCKOPS_RUNTIME_HOME=/var/lib/duckops/home \
    HOME=/var/lib/duckops/home \
    XDG_CACHE_HOME=/var/lib/duckops/xdg-cache \
    XDG_CONFIG_HOME=/var/lib/duckops/xdg-config \
    TRIVY_CACHE_DIR=/var/lib/duckops/trivy \
    SEMGREP_SEND_METRICS=off \
    REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt \
    SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt \
    GOROOT=/usr/local/go \
    PATH=/app/venv/bin:/home/duckops/.local/bin:/home/duckops/.npm-global/bin:/usr/local/go/bin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    adduser \
    bash \
    ca-certificates \
    curl \
    git \
    jq \
    libcap2-bin \
    libpcap0.8 \
    libnss3-tools \
    nodejs \
    nmap \
    python3 \
    python3-pip \
    sqlmap \
    wapiti && \
    adduser --uid 1000 --disabled-password --gecos "" duckops && \
    mkdir -p \
    /app/certs \
    /home/duckops/.local/bin \
    /home/duckops/.local/pipx/venvs \
    /home/duckops/.npm-global/bin \
    /home/duckops/.npm-global/lib/node_modules \
    /home/duckops/tools \
    /opt/nuclei-templates \
    /var/lib/duckops \
    /workspace && \
    chown -R duckops:duckops /home/duckops /var/lib/duckops /workspace && \
    setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip /usr/bin/nmap && \
    apt-get purge -y --auto-remove libcap2-bin && \
    rm -rf /var/lib/apt/lists/*

COPY --from=server-builder /out/duckops /usr/local/bin/duckops
COPY --from=server-builder /usr/local/go /usr/local/go

# COPY --from=go-tool-builder /out/gobuster /usr/local/bin/gobuster
COPY --from=go-tool-builder /out/gitleaks /usr/local/bin/gitleaks
COPY --from=go-tool-builder /out/tfsec /usr/local/bin/tfsec
COPY --from=go-tool-builder /out/gosec /usr/local/bin/gosec
COPY --from=go-tool-builder /out/syft /usr/local/bin/syft
COPY --from=go-tool-builder /out/httpx /usr/local/bin/httpx
COPY --from=go-tool-builder /out/katana /usr/local/bin/katana
COPY --from=go-tool-builder /out/nuclei /usr/local/bin/nuclei
COPY --from=go-tool-builder /out/subfinder /usr/local/bin/subfinder
COPY --from=go-tool-builder /out/trivy /usr/local/bin/trivy
COPY --from=go-tool-builder /out/trufflehog /usr/local/bin/trufflehog

COPY --chown=duckops:duckops --from=cert-builder /out/ca.crt /app/certs/ca.crt
COPY --chown=duckops:duckops --from=cert-builder /out/ca.key /app/certs/ca.key
COPY --chown=duckops:duckops --from=cert-builder /out/ca.p12 /app/certs/ca.p12

COPY --chown=duckops:duckops --from=extras-builder /app/venv /app/venv
COPY --chown=duckops:duckops --from=extras-builder /home/duckops/.npm-global /home/duckops/.npm-global
COPY --chown=duckops:duckops --from=extras-builder /home/duckops/tools/jwt_tool /home/duckops/tools/jwt_tool
COPY --chown=duckops:duckops --from=extras-builder /opt/nuclei-templates /opt/nuclei-templates
COPY --chown=duckops:duckops --from=extras-builder /opt/docker-bench-security /opt/docker-bench-security

RUN cp /app/certs/ca.crt /usr/local/share/ca-certificates/duckops-ca.crt && \
    update-ca-certificates && \
    ln -sf /app/venv/bin/arjun /usr/local/bin/arjun && \
    ln -sf /app/venv/bin/bandit /usr/local/bin/bandit && \
    ln -sf /app/venv/bin/checkov /usr/local/bin/checkov && \
    ln -sf /app/venv/bin/semgrep /usr/local/bin/semgrep && \
    ln -sf /opt/docker-bench-security/docker-bench-security.sh /usr/local/bin/docker-bench-security && \
    ln -sf /home/duckops/.npm-global/lib/node_modules/eslint/bin/eslint.js /home/duckops/.npm-global/bin/eslint && \
    ln -sf /home/duckops/.npm-global/lib/node_modules/retire/lib/cli.js /home/duckops/.npm-global/bin/retire && \
    ln -sf /home/duckops/tools/jwt_tool/jwt_tool.py /usr/local/bin/jwt_tool && \
    chmod 644 /app/certs/ca.crt && \
    chmod 600 /app/certs/ca.key /app/certs/ca.p12

RUN trivy image --download-db-only --cache-dir /var/lib/duckops/trivy 2>/dev/null || true

# Pre-cache semgrep rules for offline use.
RUN mkdir -p /workspace/rules-cache && echo "package main" > /workspace/rules-cache/test.go && \
    SEMGREP_SEND_METRICS=off semgrep --config=p/default --json /workspace/rules-cache/test.go >/dev/null 2>&1 || true && \
    rm -rf /workspace/rules-cache

# Restore ownership after any root-run commands that may have touched /var/lib/duckops.
RUN chown -R duckops:duckops /var/lib/duckops /home/duckops

COPY containers/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/docker-entrypoint.sh && \
    chmod +x /usr/local/bin/docker-entrypoint.sh

USER duckops
WORKDIR /workspace
ENV DUCKOPS_SANDBOX_MODE=true
EXPOSE 48080 48081
ENTRYPOINT ["docker-entrypoint.sh"]
