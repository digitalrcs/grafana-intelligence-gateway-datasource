# Installation and Quick Start

## Install a release

1. Download the release ZIP from the [GitHub releases](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/releases).
2. Extract its single `digitalrcs-intelligencegateway-datasource` directory into Grafana's plugin directory.
3. Keep every packaged backend executable executable on Linux (`0755`).
4. Restart Grafana.
5. Open **Connections > Data sources > Add new data source** and select **Intelligence Gateway Secure AI**.
6. Configure the provider and credential, then select **Save & test**.

The first catalog-review build is intentionally unsigned. Development Grafana instances must allow the exact plugin ID until Grafana grants a public signature level.

## Credential-free reviewer environment

Clone the repository and run:

```bash
npm ci
npm run build
mage -v build:linux
docker compose up --build
```

Open <http://localhost:3000>, sign in with `admin` / `admin`, and open **Dashboards > Intelligence Gateway Data Source Review**. The stack provisions a deterministic local provider and never requires a real provider key. Set `GRAFANA_PORT` before startup if port 3000 is unavailable.

For production use, continue with [Configuration](Configuration), [Security and Secrets](Security-and-Secrets), and [Panel Integration](Panel-Integration).
