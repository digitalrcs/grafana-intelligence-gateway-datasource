# Intelligence Gateway Secure AI data source

This DigitalRCS backend data-source plugin is the secure provider boundary for the companion
`digitalrcs-intelligencegateway-panel`. Grafana encrypts provider credentials stored in `secureJsonData`; only the Go
backend decrypts and adds them to provider requests. Dashboards and browser code never receive a credential.

Plugin ID: `digitalrcs-intelligencegateway-datasource`  
Grafana: `>=11.6.0`

## Configuration contract

Non-secret `jsonData`:

```json
{
  "provider": "openai",
  "baseUrl": "https://api.openai.com/v1",
  "defaultModel": "gpt-4.1-mini",
  "timeoutSeconds": 300,
  "allowedModels": ["gpt-4.1-mini", "gpt-4.1"],
  "maxOutputTokens": 256000,
  "allowStreaming": false
}
```

Write-only `secureJsonData` supports `apiKey`, `bearerToken`, and the reserved `clientSecret` field. A bearer token takes
precedence over an API key. OAuth client-credentials exchange is not enabled yet because the contract does not define a
token URL or client ID; storing the client secret now is forward-compatible but it is never sent as a provider bearer
token.

The included provisioning example reads `OPENAI_API_KEY` from the Grafana server/container environment. Never commit a
real key or put one in dashboard JSON.

## Backend API

The panel can access these data-source resources through Grafana's authenticated data-source resource API:

- `GET /models` proxies the configured provider model list.
- `POST /chat/completions` and `POST /analyze` accept a bounded gateway request and return the provider's
  OpenAI-compatible response.

Example request body:

```json
{
  "prompt": {
    "system": "You are a careful observability analyst.",
    "user": "Assess the supplied dashboard data."
  },
  "model": "gpt-4.1-mini",
  "temperature": 0.2,
  "maxOutputTokens": 1200,
  "stream": false
}
```

The effective output cap is `min(request.maxOutputTokens, jsonData.maxOutputTokens)`. If the request omits a cap, the
administrator ceiling is still sent to the provider. The ordinary data-source query editor exposes the same basic
request and returns a one-row `answer` frame, while the companion panel should use the resource API directly.

## Security policy

- Provider URLs are administrator-controlled and resource paths are fixed; dashboard requests cannot choose a host or
  arbitrary upstream path.
- OpenAI is restricted to `api.openai.com`. Custom remote providers require HTTPS; this permits an administrator to
  explicitly select a public or private compatible service. LM Studio may use HTTP only for `localhost`,
  `host.docker.internal`, or a loopback IP.
- Redirects are rejected. Link-local, multicast, and unspecified destinations are rejected.
- Request bodies are capped at 1 MiB and provider responses at 16 MiB.
- Each data-source instance permits four concurrent calls and 30 calls per minute. Provider/model-specific budget and
  organization billing limits should remain enabled as the authoritative cost control.
- Provider response bodies are never included in error messages. Prompts, responses, and credentials are not logged.
- Only `system`, `user`, and `assistant` message roles are accepted. The configured model allow-list and token ceiling are
  enforced server-side.

Grafana data-source permissions should restrict who may edit and query this instance. Review egress/firewall rules as a
second SSRF boundary.

## Development

Requirements: Node.js 22+, npm, Go version from `go.mod`, Mage, Docker, and Docker Compose. Grafana's scaffold tooling is
supported on Windows through WSL.

```bash
npm install
npm run typecheck
npm run lint
npm run test:ci
npm run build
mage -v test
mage -v build:linux
npm run package:plugin
docker compose up
```

Use `npm run package:plugin` instead of Windows `Compress-Archive`: it preserves the Linux backend binary's required
`0755` executable mode and the explicit top-level plugin directory in the ZIP.

Open <http://localhost:3000>, configure the provisioned data source, then use **Save & test**. The health check calls the
provider's `/models` route so it verifies network access and credentials without sending a prompt.

Restart Grafana after changing `src/plugin.json`. The generated GitHub Actions workflows build the frontend and backend,
run tests, package a ZIP with the required top-level plugin directory, and sign when
`GRAFANA_ACCESS_POLICY_TOKEN` is configured.

## Panel integration

In the companion panel, select this instance under **AI provider > Secure AI data source**. The panel stores this
data-source UID plus non-secret model, temperature, and output choices. **Load models securely** uses the `/models`
resource, and **Analyze** posts the constructed prompt to `/chat/completions`. Secure mode is buffered in the current
panel release. Existing direct provider fields remain only for local development and dashboard migration.

## Integrated Docker test

When both repositories are sibling directories, the panel repository's `docker-compose.yaml` mounts both `dist`
directories and provisions this data source with UID `intelligence-gateway-secure`. Build this frontend and Linux backend,
build the panel, set `OPENAI_API_KEY` in the shell, and run `docker compose up --build` from the panel repository. Open
<http://localhost:3004> and use the provisioned CSV assessment dashboard to exercise the intended end-to-end flow.

## License

Apache-2.0. Copyright 2026 DigitalRCS.
