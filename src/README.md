# Intelligence Gateway Secure AI

Secure server-side AI provider access for the DigitalRCS Grafana Intelligence Gateway panel.

Configure the provider URL, allowed models, timeout, and output-token ceiling in `jsonData`. Store
credentials only in the API key or bearer-token fields; Grafana encrypts these `secureJsonData`
values and never returns them to browser code.

The backend exposes `GET /models` plus `POST /chat/completions` and `POST /analyze` data-source resources. Model discovery returns only administrator-approved IDs. It enforces the
administrator's model and token policies, fixed upstream paths, TLS/host rules, redirect blocking, bounded bodies,
timeouts, concurrency/rate limits, and sanitized errors. The ordinary query editor can also submit a prompt and returns
the assessment as a one-row data frame.

For OpenAI, use `https://api.openai.com/v1`. HTTPS is required by default. If an organizational provider endpoint is
available only over HTTP and its locality cannot be classified reliably, an administrator can explicitly enable
**Allow insecure HTTP**. This override permits HTTP regardless of hostname/IP classification and can expose credentials
and prompts in transit, so use it only with appropriate trusted-network and firewall controls.

Limit who may edit and query this data source. Put provisioning secrets in the Grafana server/container environment,
never in a dashboard or committed YAML file.

Project documentation: <https://github.com/DigitalRCS/grafana-intelligence-gateway-datasource>
