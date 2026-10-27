# Security and Secrets

## Request boundary

1. An administrator stores a provider credential in Grafana `secureJsonData`.
2. The panel stores only the data-source UID and non-secret generation choices.
3. Grafana resolves the data-source instance server-side.
4. The backend validates the fixed provider host, model, token cap, timeout, body, and rate policies.
5. The backend adds the decrypted credential and sends the provider request.
6. The browser receives the provider's bounded JSON response or a sanitized error, never the stored credential. Provider responses can include usage metadata and reasoning fields; the panel renders the visible final answer rather than presenting reasoning as an assessment.

## Enforced controls

- Fixed Chat Completions resource paths; provider-specific Open WebUI, Grafana LLM, and Direct Line routes remain administrator-bound. Browser requests cannot choose a host or path.
- HTTPS by default; the administrator-only HTTP override is disabled by default and visibly warns about exposure.
- OpenAI host restriction, redirect rejection, and prohibited-address checks.
- 1 MiB text-request/text-portion ceiling, 18 MiB encoded image-request ceiling, and 16 MiB buffered response ceiling.
- Image sending is disabled by default and requires a separate explicit subset of administrator-approved models. Only inline base64 PNG/JPEG images are accepted, with format, full-decode, count, dimension, pixel, and byte checks. The backend creates image content parts itself; arbitrary remote image URLs and unknown fields are rejected.
- Four concurrent calls and 30 calls per minute per data-source instance; further active requests fail immediately with HTTP 429 instead of queuing image bodies.
- Model allow-list enforced for generation and discovery; an empty list permits only the administrator default. Discovery exposes permitted IDs, not provider-specific metadata.
- Administrator token ceiling and temperature validation for Chat Completions providers. Copilot Studio declares these controls unsupported because response policy belongs to the agent. Supported roles and request bounds remain enforced; streaming requests are rejected.
- No prompt, answer, credential, or upstream error-body logging by default.
- Non-cacheable resource responses, JSON-only chat responses, disabled content-type sniffing, and sanitized transport/body-read errors.

Use Grafana data-source permissions to restrict edit and query access. Keep provider budgets and organization billing limits enabled, and use egress/firewall policy as a second SSRF boundary.

Grafana LLM uses a separately configured service-account bearer token restricted to the LLM app resource API. It does not forward browser cookies or impersonate the requesting dashboard viewer. Open WebUI uses the configured WebUI API credential. Copilot Studio uses a separate approved Direct Line channel secret, a fresh conversation per assessment, and a scoped ephemeral channel token. It does not perform end-user SSO or bypass agent authentication; sign-in responses are errors. Its temperature/output policies belong to the agent. Images, prompts, and responses are held in memory during the request and are not written to plugin storage or logs. Provider retention policies still apply.

## Input selection and consent

The panel's data-only, images-only, and combined workflows use the same backend boundary. Images are previewed in browser memory and are sent only by an explicit **Analyze images** action; enabling image support is not an automatic capture or sending permission. Turning off **Include query data with images** omits serialized query rows but does not remove the configured prompt, visible image contents, or non-secret dashboard context.

Images may expose chart labels, annotations, names, or other sensitive information that is absent from query data. Treat image text and attachment names as untrusted evidence, not instructions. Captured Intelligence Gateway panels are excluded to avoid recursively sending assessments, but that is not a general redaction system. Review every preview and apply the provider's retention, access, and organizational data-handling rules. Clearing browser previews cannot retract data already delivered to a provider.

Plain JSON data sources, dashboard JSON, and uploaded configuration files are not secret stores. Do not load an API key from an ordinary query result or return it through a frame. Use Grafana's secure fields, or provision `secureJsonData` from a server-side secret environment; never commit the resulting credential.

Grafana's plugin policy expects credentials to be stored securely and transmitted over secure transport. The HTTP option is an explicit operational escape hatch for controlled internal networks, not the recommended production default; its acceptability remains subject to Grafana's catalog review.
