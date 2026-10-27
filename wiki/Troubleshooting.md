# Troubleshooting

## The model reasons but returns no final answer

An empty visible answer accompanied by reasoning and a `length` finish reason means generation stopped at a length limit while still reasoning. This is distinct from a network failure or a response timeout.

1. Check **Maximum output tokens** in the panel and on the secure data-source instance. The smaller value is sent to the provider; the provider may impose a smaller limit again.
2. Increase the panel budget within the administrator ceiling and model's supported limits. For example, 8,192 is a bounded starting point for testing a reasoning model, not a guarantee that it will finish.
3. Use **Requested answer max tokens (soft)** for a concise visible answer, understanding that this is a prompt instruction rather than a separate answer budget.
4. If necessary, use a provider-supported non-reasoning configuration or a different model. The gateway does not currently expose a reasoning-effort control.

The panel's provider-default output option removes only its own cap. It does not bypass the administrator ceiling. A model advertised with a 256K context window does not necessarily support a 256K output. Do not repeatedly retry with an unlimited budget or treat partial reasoning as a completed operational assessment.

## The browser still shows controls from an older build

Confirm the full browser URL and installed plugin version on the exact Grafana instance being tested. Replace both complete plugin packages and restart Grafana after metadata changes; replacing only `module.js` under an unchanged version can leave an already-loaded browser module in use. Reload the whole browser page, not just Grafana's dashboard query-refresh button. Brave's reload-ignoring-cache shortcuts are **Ctrl+Shift+R** or **Shift+F5**.

If a hard reload does not help, compare a fresh private window with the normal window before assuming a server defect or deleting browser data. A long paragraph appearing as a generated image name, rather than the panel heading, indicates older capture code or a different installation; record the full URL and both plugin versions. The 1.1.0 panel bounds generated capture labels and excludes the text-panel body. Do not change secrets or provider settings to repair an image-name validation error.

## Data source is busy

Only four requests may run concurrently per data-source instance. Additional requests fail immediately with HTTP 429;
they are not queued with their image payloads. Wait for an active request to finish and retry. The separate 30-requests-per-minute
limit still applies, including rejected attempts.

## Save & test still says HTTP is restricted

Confirm **Allow insecure HTTP** is checked on the same data-source instance and select **Save & test**. If the checkbox is missing or the old error remains, rebuild both plugin halves and restart Grafana:

```bash
npm run build
mage -v build:linux
docker compose restart grafana
```

A backend executable from an older build can continue enforcing the previous locality-only rule even when the frontend displays the newer setting.

## Model is rejected

The requested model must match `allowedModels` exactly. If the list is empty, only `defaultModel` is permitted. A nonempty list must include the default. Add the model administratively or use the configured default; secure model discovery shows only permitted IDs.

## Image analysis is disabled or rejected

The administrator must enable **Enable image analysis** and approve the selected model in **Allowed vision models**, which must also satisfy the ordinary model allow-list. A text model is not automatically a vision model. If the provider rejects a validated image request, confirm actual vision support and reduce image count/dimensions or prompt size. PNG and JPEG are supported; remote image URLs, SVG, WebP, GIF, malformed base64, truncated images, and excessive sizes are rejected before sending.

Data-only requests do not require vision. Images-only requests still require a prompt and the secure AI connection; turn off **Include query data with images** in the panel if query rows should be omitted. Combined requests send only the Intelligence panel's query results, not every pictured panel's raw data.

## Copilot Studio needs sign-in or never completes

The preview adapter uses an approved published agent's Direct Line channel secret; a normal Microsoft/enterprise SSO login is not that credential. It does not implement delegated end-user authentication, OAuth-card interaction, or SSO token exchange. Do not disable agent authentication to make a test pass. Ask the tenant administrator for an approved integration and authentication design.

The agent must return a correlated text reply and a supported completion signal (`acceptingInput`, `expectingInput`, or `endOfConversation`). Progress-only or non-text replies are not treated as final assessments. For images, the published agent must permit file uploads. A successful **Save & test** checks token exchange only, not the agent's inference, file understanding, or tenant-specific authorization.

## Grafana LLM connection fails

Use the fixed resource base ending `/api/plugins/grafana-llm-app/resources/llm/v1`, preserve your Grafana subpath, and select the `base` or `large` alias. Store a Grafana service-account token with LLM app access in **Bearer token**. The configured app requires Grafana 13+; its two-minute upstream timeout can be shorter than the data-source timeout. Interactive enterprise SSO or browser login does not provision a backend service-account token.

## Provider times out

The panel, data source, proxy, provider, and Grafana LLM app can each have their own deadlines. The first reached deadline ends the request. Reduce input/image size or use an appropriate model; review configured timeouts with the administrator. Increasing a panel timeout cannot extend an upstream provider deadline.

## Provider cannot be reached

Remember that the backend runs from the Grafana server/container. Test DNS and connectivity from that network namespace. Redirects are intentionally rejected; configure the final provider URL directly.

## Credential does not update

Saved secrets are write-only. Select the field's reset control, enter the replacement, and save. Grafana will display only configured state afterward.
