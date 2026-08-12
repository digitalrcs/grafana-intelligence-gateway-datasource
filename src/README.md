# Intelligence Gateway Secure AI

Secure server-side AI provider access for the DigitalRCS Grafana Intelligence Gateway panel.

Configure the provider URL, allowed models, timeout, output-token ceiling, and streaming policy in `jsonData`. Store
credentials only in the API key, bearer token, or OAuth client-secret fields; Grafana encrypts these `secureJsonData`
values and never returns them to browser code.

The backend exposes `GET /models` plus `POST /chat/completions` and `POST /analyze` data-source resources. It enforces the
administrator's model and token policies, fixed upstream paths, TLS/host rules, redirect blocking, bounded bodies,
timeouts, concurrency/rate limits, and sanitized errors. The ordinary query editor can also submit a prompt and returns
the assessment as a one-row data frame.

For OpenAI, use `https://api.openai.com/v1`. For a local LM Studio server in Docker, use
`http://host.docker.internal:1234/v1`. Remote custom providers must use HTTPS.

Limit who may edit and query this data source. Put provisioning secrets in the Grafana server/container environment,
never in a dashboard or committed YAML file.

Project documentation: <https://github.com/DigitalRCS/grafana-intelligence-gateway-datasource>
