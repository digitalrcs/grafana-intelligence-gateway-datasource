# Intelligence Gateway Secure AI

The server-side AI connection for **Grafana Intelligence Gateway**.

Intelligence Gateway Secure AI by **DigitalRCS** is purpose-built to accompany the [Grafana Intelligence Gateway panel](https://github.com/digitalrcs/grafana-intelligence-gateway). Configure your AI provider once, keep its credentials in Grafana's encrypted data-source settings, and let your Intelligence Gateway panels use that connection without storing provider keys in dashboards.

Use local models through **LM Studio**, hosted models through **OpenAI**, or another provider with a compatible **Chat Completions API**. Administrators choose the endpoint, allowed models, and request limits; dashboard authors choose the data and question to analyze.

## Built for Intelligence Gateway

The two plugins work together but have different responsibilities:

| Plugin                                         | Role                                                                                                                            |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **Grafana Intelligence Gateway** panel         | Builds the prompt from selected Grafana query results and operational context, then displays the assessment beside your charts. |
| **Intelligence Gateway Secure AI** data source | Connects to the model from the Grafana server, supplies the configured credential, and applies administrator request policies.  |

Together, they help operators compare traffic across sites, summarize an incident window, and prepare handovers with observations and follow-up checks. This data source is the AI connection, **not the source of the metrics being analyzed**. Keep using your existing metric, log, or table data sources for those queries.

![The companion Intelligence Gateway panel displaying a real local-model assessment of synthetic traffic data](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/intelligence-gateway-assessment.png)

_Example of the two plugins working together: the panel presents a local LM Studio assessment of synthetic data; this data source provides the server-side model connection. The chart and assessment are panel content, not a visualization supplied by the data source._

## What you can configure

- **A shared provider connection.** Reuse a configured instance across Intelligence Gateway panels, or create separate instances for different providers and teams.
- **Server-side credentials.** Store the provider API key or bearer token in Grafana's secure fields. Saved credentials show configured/reset state rather than their stored values.
- **Model selection.** Set a default model and an optional allow-list of models that panels may request.
- **Request limits.** Set a server-side timeout and output-token ceiling. A panel can request a lower output limit, but cannot raise the administrator ceiling.
- **Transport policy.** Use HTTPS by default. An explicit **Allow insecure HTTP** override supports controlled HTTP-only environments, with the transport risks described below.

The backend also limits request/response sizes and applies per-instance rate and concurrency controls. Provider billing limits remain important; output-token and request limits are not a monetary budget.

## Requirements and installation

Requires **Grafana 11.6 or later**, an AI provider reachable from the Grafana server, and the **Grafana Intelligence Gateway** panel for the intended dashboard experience. Provider access, available models, and any usage charges are managed separately.

Install both plugins into the **same Grafana installation**. They are separate plugin packages, installed side by side in Grafana's configured plugins directory—not inside one another. A configured data-source instance must be available in the **same Grafana organization** as the dashboard.

See the [installation and plugin-placement guide](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Installation-and-Quick-Start) for release installation, Docker paths, and unsigned review environments. Installing the plugin does not create or configure a provider connection automatically.

## Connect your first panel

1. In **Connections → Data sources**, add **Intelligence Gateway Secure AI**.
2. Choose **OpenAI**, **LM Studio**, or **Custom**. Enter the provider base URL, normally ending in `/v1`, and the exact default model ID. The URL must be reachable from the Grafana server or container, not just your browser.
3. Enter an API key or bearer token only if your provider requires it. Set an allowed-model list, a timeout, and an output ceiling appropriate to the model and your budget. Leave streaming disabled for the companion-panel workflow.
4. Select **Save & test**. This checks the provider's model endpoint; it does not generate an assessment or prove model quality.
5. Add or edit a **Grafana Intelligence Gateway** panel. Under **Queries**, select the data to analyze—for example, **Dashboard** and an existing source panel.
6. Under **AI provider → Secure AI data source**, select the instance you just configured. Choose **Load models securely**, select an allowed model, set the analysis instructions, and select **Analyze**.

Do not select this AI data source as the panel's metric source when your intention is to analyze another panel's data. The metric/query connection and the AI provider connection are separate. The [panel integration guide](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Panel-Integration) explains both selectors.

## Security and compatibility

Provider requests run on the Grafana server. Selected query data, prompts, and operational context are sent to the configured provider. Choose a local or hosted endpoint according to your organization's data-handling requirements, and restrict who can edit and query the data-source instance.

Prefer HTTPS. **Allow insecure HTTP** disables the HTTPS requirement for that instance regardless of hostname or IP classification. Use it only with appropriate trusted-network and firewall controls: prompts, responses, and credentials can otherwise be intercepted in transit. It is not a way to bypass certificate validation for an HTTPS endpoint.

Supported authentication is a provider API key or bearer token. OAuth client-credentials exchange is not implemented; the reserved client-secret field is not an operational OAuth setup. The companion panel uses buffered, non-streaming responses. Custom providers and models must support the request format used by the gateway; OpenAI compatibility is not a guarantee that every model or feature is supported.

The query editor can also submit a supplied prompt and return an answer as a Grafana data frame, but it does not automatically collect other panels' data. Generated assessments require human verification and do not establish root cause or perform remediation.

## Documentation and support

- [Installation and quick start](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Installation-and-Quick-Start)
- [Provider configuration and provisioning](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Configuration)
- [Security and secret storage](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Security-and-Secrets)
- [Companion panel documentation](https://github.com/digitalrcs/grafana-intelligence-gateway/wiki)
- [Worked example with real-provider screenshots](https://github.com/digitalrcs/grafana-intelligence-gateway/wiki/Reviewer-Walkthrough)
- [Report an issue](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/issues)
- [Source code and releases](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource)

When reporting a problem, include Grafana and plugin versions, provider type, and the sanitized error. Do not include credentials or sensitive query data. The credential-free reviewer environment returns an explicitly labeled mock receipt; it does not perform AI inference.

Apache-2.0. Developed by [DigitalRCS](https://digitalrcs.com).

<a href="https://digitalrcs.com"><img src="https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/docs/images/digitalrcs-logo.webp" alt="DigitalRCS — Solutions that connect. Inside and out." width="240" /></a>
