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
2. Extract each ZIP's plugin-ID directory into Grafana's plugin directory: `digitalrcs-intelligencegateway-panel` and `digitalrcs-intelligencegateway-datasource`, side by side as shown above.
3. Keep every packaged backend executable executable on Linux (`0755`).
4. Restart Grafana.
5. Open **Connections > Data sources > Add new data source** and select **Intelligence Gateway Secure AI**.
6. Configure the provider and credential, then select **Save & test**. This checks model endpoint access, not inference quality.
7. In an Intelligence Gateway panel, select this configured instance under **AI provider → Secure AI data source**. Keep **Queries** connected to the metrics/table data to analyze. See [Panel Integration](Panel-Integration).

Installing either package does not create a data-source instance, supply a credential, or select it in existing dashboards. Grafana's installer can resolve declared plugin dependencies from its plugin repository, but the dependent package must be available there. During review, install both provided archives explicitly. The panel declares this data source as a required dependency; this data source does not declare a reverse dependency on the panel.

### Upgrade the pair to 1.1.0

For the full image workflow, install version 1.1.0 of both plugins. Back up the existing plugin directories and Grafana configuration, replace each complete package in its own directory, and restart Grafana. Verify the installed versions on the intended server and reload the whole browser page before testing; a dashboard query refresh does not reload plugin code. Existing data-source credentials and dashboard UIDs are not meant to be recreated during this upgrade.

Image analysis is opt-in: existing text-only instances do not gain an image-model allow-list automatically. Enable it only after confirming the provider/model accepts images. Follow [Panel Integration](Panel-Integration) for data-only, images-only, and combined assessments, and [Release 1.1.0](Release-1.1.0) for validation boundaries. Do not treat the presence of a source branch or this guide as proof that a particular release ZIP has been published or approved.

### Unsigned review builds

An archive prepared for initial catalog review may be unsigned until Grafana grants a public signature level. Check the exact release's validation record rather than assuming its signature status. Use an unsigned archive only in a controlled self-managed development/review environment that explicitly permits `digitalrcs-intelligencegateway-datasource`. If using an unsigned panel build too, permit `digitalrcs-intelligencegateway-panel` as well. Do not disable signature verification globally or assume unsigned archives are installable in Grafana Cloud. Follow the applicable signed-release and deployment requirements for production.

## Phase 1: credential-free mock connection check

> **DEMO ONLY - Mock AI (no inference)** returns a fixed delivery receipt. No AI model runs in the supplied demo, and the receipt does not understand or assess query data or images. No cloud credentials or hosted AI access are shipped. The mock is only a transport/rendering check, not an inference test.

Clone the repository and run:

```bash
npm ci
npm run build
mage -v build:linux
docker compose up --build
```

Open <http://localhost:3000>, sign in with `admin` / `admin`, and open **Dashboards > Intelligence Gateway Data Source Review**. The stack provisions **DEMO ONLY - Mock AI (no inference)** without a real provider key. Its settings show the demo warning and real-provider checklist. Select **Save & test**, then run the supplied query to see the labeled mock receipt. Set `GRAFANA_PORT` before startup if port 3000 is unavailable.

Keep this mock instance configured for repeatable tests. Do not replace it with a private provider or add real credentials to the mock.

## Phase 2: real assessment with your own provider

Genuine inference requires an approved reachable provider, a compatible model/agent, and any required credential. For catalog review, use an existing approved provider or coordinate suitable limited test API access privately with DigitalRCS through the existing Grafana review conversation before Phase 2. No hosted service or credentials are currently bundled or promised; any access must be arranged and confirmed separately. Never publish credentials or private endpoint details, and do not count the mock check as completed real-provider review.

Local LM Studio may be configured without authentication; hosted providers normally require an authorized account and credential. Confirm model availability, retention policy, and any usage limits or costs for the selected test connection.

1. In **Connections → Data sources → Add new data source**, add a **separate Intelligence Gateway Secure AI** instance. Give it a descriptive name such as `Operations AI` and leave the mock instance unchanged.
2. Select the real provider and enter its **Base URL** as reached from the **Grafana server/container**, not the browser. For LM Studio on a Docker Desktop host, use `http://host.docker.internal:1234/v1` when its server accepts connections from Docker. For a remote host, use an address reachable from the container. `localhost` in a container points back to that container; it is not your desktop. See [Configuration](Configuration) for each provider's path.
3. Enter the exact **Default model** and include it in **Allowed models**. Verify it is available on that provider. Store the API key, bearer token, or Direct Line secret in the appropriate secure field only. Leave credentials blank only when the provider intentionally does not require them. Prefer HTTPS; permit HTTP only for an approved trusted-network connection.
4. Set a supported timeout and output budget, then select **Save & test**. This checks connectivity/model access or Direct Line token exchange; it does not run inference or prove image understanding.
5. For image assessment, enable **Enable image analysis** and list a verified image-capable model in **Allowed vision models**. It must also be allowed by the model policy. Text-only compatibility is not sufficient.
6. Edit the Intelligence Gateway panel and select the **new real instance** under **AI provider → Secure AI data source**. Load permitted models and choose the intended one. Do not put this AI connection in **Queries** when analyzing another panel: **Queries** must still select the metric/table source, such as **Dashboard** plus the source chart.
7. Run **Analyze** with synthetic query data, or preview synthetic images and explicitly select **Analyze images**. Check that the result assesses the supplied evidence rather than displaying a mock receipt. Test image-only inference separately from combined images and query data.

The mock notice is a fixture notice, not an inference detector or security control: it requires the supplied `reviewMock: true` marker, Custom provider, and the exact mock base URL (ignoring surrounding whitespace and trailing slashes). Changing the provider or URL removes it; there is no editable mock toggle. The backend transport and authentication are unchanged.

Continue with [Configuration](Configuration), [Security and Secrets](Security-and-Secrets), and [Panel Integration](Panel-Integration) before production use.

## If the panel cannot find the data source

- Check that both plugin IDs are installed and loaded in the same Grafana server. Restart after installation or metadata changes.
- Confirm the correct backend executable and permissions; inspect Grafana's plugin startup errors.
- Create and save a data-source **instance** in the dashboard's organization. An installed plugin by itself is not a selectable connection.
- Check the user's permissions and the panel's **AI provider** selector. Its **Queries** selector serves a different purpose.
- From Docker, `localhost` means the container itself. Use a provider address reachable from that container, with HTTPS or an explicitly approved HTTP override.

See Grafana's [packaging guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin) and [plugin metadata reference](https://grafana.com/developers/plugin-tools/reference/plugin-json) for the underlying layout and dependency contract.
