# Security and Secrets

## Request boundary

1. An administrator stores a provider credential in Grafana `secureJsonData`.
2. The panel stores only the data-source UID and non-secret generation choices.
3. Grafana resolves the data-source instance server-side.
4. The backend validates the fixed provider host, model, token cap, timeout, body, and rate policies.
5. The backend adds the decrypted credential and sends the provider request.
6. The browser receives only the answer or sanitized error metadata.

## Enforced controls

- Fixed `/models` and `/chat/completions` upstream paths.
- HTTPS by default; the administrator-only HTTP override is disabled by default and visibly warns about exposure.
- OpenAI host restriction, redirect rejection, and prohibited-address checks.
- 1 MiB request and 16 MiB buffered response ceilings.
- Four concurrent calls and 30 calls per minute per data-source instance.
- Model allow-list, administrator token ceiling, supported roles, and temperature validation.
- No prompt, answer, credential, or upstream error-body logging by default.

Use Grafana data-source permissions to restrict edit and query access. Keep provider budgets and organization billing limits enabled, and use egress/firewall policy as a second SSRF boundary.

Grafana's plugin policy expects credentials to be stored securely and transmitted over secure transport. The HTTP option is an explicit operational escape hatch for controlled internal networks, not the recommended production default; its acceptability remains subject to Grafana's catalog review.
