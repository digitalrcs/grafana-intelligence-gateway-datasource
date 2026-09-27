# Installation and Quick Start

## Install the companion pair

**Intelligence Gateway Secure AI** is the server-side connection for the **Grafana Intelligence Gateway** panel. Install both plugins in the same Grafana installation. Configure the data-source instance in the same Grafana organization as the dashboard.

| Plugin                         | Plugin ID                                   | Where you use it                                                                         |
| ------------------------------ | ------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Grafana Intelligence Gateway   | `digitalrcs-intelligencegateway-panel`      | Add/edit a dashboard visualization.                                                      |
| Intelligence Gateway Secure AI | `digitalrcs-intelligencegateway-datasource` | Connections → Data sources; then select the instance in the panel's AI provider options. |

They are separate packages. For a typical Docker deployment, the extracted layout is:

```text
/var/lib/grafana/plugins/
  digitalrcs-intelligencegateway-panel/
    plugin.json
    module.js
    README.md
    img/
  digitalrcs-intelligencegateway-datasource/
    plugin.json
    module.js
    README.md
    gpx_intelligencegateway_linux_amd64
    img/
```

Use your installation's configured `paths.plugins` / `GF_PATHS_PLUGINS` directory if it differs. Choose the backend executable appropriate to your Grafana server's OS/architecture; release archives can include several. The browser's OS is not the deciding factor.

Do **not** place the data source inside the panel directory or combine both plugins into one ZIP. When mounting source builds, mount each repository's `dist` directory at its own plugin-ID path. Source repositories may be siblings for the development Compose setup; production installs need the built packages, not the source tree or `node_modules`.

## Install a release

1. Obtain compatible release ZIPs for the [data source](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/releases) and [panel](https://github.com/digitalrcs/grafana-intelligence-gateway/releases). Once both are available in the Grafana catalog, your administrator can use the catalog installation flow instead.
2. Extract its single `digitalrcs-intelligencegateway-datasource` directory into Grafana's plugin directory.
3. Keep every packaged backend executable executable on Linux (`0755`).
4. Restart Grafana.
5. Open **Connections > Data sources > Add new data source** and select **Intelligence Gateway Secure AI**.
6. Configure the provider and credential, then select **Save & test**. This checks model endpoint access, not inference quality.
7. In an Intelligence Gateway panel, select this configured instance under **AI provider → Secure AI data source**. Keep **Queries** connected to the metrics/table data to analyze. See [Panel Integration](Panel-Integration).

Installing either package does not create a data-source instance, supply a credential, or select it in existing dashboards. Grafana's installer can resolve declared plugin dependencies from its plugin repository, but the dependent package must be available there. During review, install both provided archives explicitly. The panel declares this data source as a required dependency; this data source does not declare a reverse dependency on the panel.

### Unsigned review builds

The initial data-source catalog-review build is unsigned. Use it only in a controlled self-managed development/review environment that explicitly permits `digitalrcs-intelligencegateway-datasource`. If using an unsigned panel build too, permit `digitalrcs-intelligencegateway-panel` as well. Do not disable signature verification globally or assume unsigned archives are installable in Grafana Cloud. Follow the applicable signed-release and deployment requirements for production.

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

## If the panel cannot find the data source

- Check that both plugin IDs are installed and loaded in the same Grafana server. Restart after installation or metadata changes.
- Confirm the correct backend executable and permissions; inspect Grafana's plugin startup errors.
- Create and save a data-source **instance** in the dashboard's organization. An installed plugin by itself is not a selectable connection.
- Check the user's permissions and the panel's **AI provider** selector. Its **Queries** selector serves a different purpose.
- From Docker, `localhost` means the container itself. Use a provider address reachable from that container, with HTTPS or an explicitly approved HTTP override.

See Grafana's [packaging guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin) and [plugin metadata reference](https://grafana.com/developers/plugin-tools/reference/plugin-json) for the underlying layout and dependency contract.
