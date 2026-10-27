# Intelligence Gateway Secure AI 1.1.0

## Release scope

Version 1.1.0 expands the secure backend connection used by the Grafana Intelligence Gateway panel. Use the matching 1.1.0 companion panel to assess query data, images, or both. This document describes the version's intended package contents; the [release validation record](RELEASE_VALIDATION.md) and published release assets establish the actual build/publication status. Grafana approval and signature classification are separate decisions.

### Image assessment

- Accept explicitly attached PNG/JPEG images through the authenticated resource API. The companion panel supplies rendered dashboard captures, selected-panel captures, and uploaded images with preview/removal before manual sending.
- Support data-only, images-only, and combined input. The required secure data source handles provider access in all three cases; the panel's metric **Queries** connection remains separate.
- Default image sending to off. Require an administrator-approved vision-model subset and expose the effective image policy through `GET /capabilities`.
- Enforce 8 images, 4 MiB decoded bytes each, 12 MiB combined, 4096 pixels per side, 16 million pixels per image, 32 million pixels combined, and 18 MiB encoded request JSON. The text portion retains a 1 MiB limit and provider responses a 16 MiB limit.
- Validate names, standard base64, MIME/decoded format, dimensions, and complete PNG/JPEG decoding. Do not fetch remote image URLs or accept arbitrary multimodal message parts from dashboard clients.

### Provider profiles

The provider selector includes **OpenAI**, **LM Studio**, **Open WebUI**, **Grafana LLM**, **Copilot Studio**, and **Custom**. Open WebUI uses its `/api` routes. Grafana LLM uses the separately configured LLM app with a server-side service-account token and `base`/`large` aliases; it does not impersonate the dashboard viewer.

Copilot Studio is a **bounded Direct Line preview** for an organization-approved published agent. It uses a separate secure channel secret, an isolated conversation and scoped token, bounded image upload, correlated text replies, and explicit completion signals. Interactive sign-in, enterprise SSO token exchange, and delegated end-user authentication are not implemented. A normal SSO login is not a channel credential. Temperature and output-token policy belong to the agent, not this adapter. Protocol/unit coverage does not certify a particular tenant or agent.

### Security and operational behavior

- Clearly name the supplied fixture **DEMO ONLY - Mock AI (no inference)** and show a settings warning with an own-provider checklist only for its explicitly marked Custom/mock URL configuration. The fixed receipt verifies delivery, not AI understanding. Provider/URL changes clear the notice; backend transport and authentication are unaffected.
- Separate credential-free mock review from real assessment. Keep the mock instance and add a separate real Intelligence Gateway Secure AI instance with your own reachable provider, compatible model, and credential if needed. Select it under the panel's **AI provider**, not **Queries**. No cloud credentials or hosted model access are bundled.
- Provider credentials remain in Grafana `secureJsonData`; no provider secret is added to dashboard JSON, browser configuration, image metadata, or ordinary query results.
- Keep provider hosts, model/vision allow-lists, token ceilings, timeouts, redirects, payload sizes, and rate/concurrency policies on the backend.
- Fail additional concurrent requests promptly with HTTP 429 instead of retaining queued image payloads. Resource responses are JSON, non-cacheable, and protected against content-type sniffing; provider error bodies are not exposed.
- Preserve the lower-of-panel-and-administrator hard token limit. Reasoning and final output can share that budget; a short-answer prompt is a separate soft instruction. Timeout is a separate deadline and does not increase the token budget.
- Update provider, input-mode, image-policy, secret-storage, and troubleshooting documentation, including stale browser-module diagnosis and reasoning-only length failures.
- Update Go to 1.26.10 and x/net to 0.60.0 after a fresh vulnerability scan; patch build-tool and Moment dependencies. Remaining moderate host-external router findings are documented in the release validation record.

## Upgrade

1. Back up the installed plugin directories and Grafana configuration. Install both complete 1.1.0 packages side by side in the same Grafana installation; do not combine the plugins into one ZIP or nest one inside the other.
2. Restart Grafana after replacing packages/metadata, verify both installed versions, and reload the browser page. A dashboard data refresh does not reload its plugin module.
3. Keep existing data-source credentials, instance UIDs, and dashboard query configuration. Confirm the selected models still satisfy the administrator allow-list.
4. Leave vision disabled unless required. To enable it, verify the provider/model's actual image support and approve the exact model ID under **Allowed vision models**.
5. Test with synthetic input before operational use. Separate mock delivery receipts from genuine inference, and separately verify image understanding without relying on numeric rows supplied alongside the image.

The declared minimum remains Grafana 11.6.0; the optional Grafana LLM app connection requires Grafana 13 or later. Provider compatibility, authentication, retention, and provider-side limits must be checked for each deployment. HTTP is an explicit, default-off override for controlled environments, not the recommended production transport.

## Release evidence and limitations

Before updating the Grafana submission, verify the exact public ZIP, source tag, checksum, backend executable permissions, validator result, signature state, and GitHub provenance attestation. Keep the panel and data-source release fields separate. Publish Wiki pages to the Wiki repository in addition to committing their source under `wiki/`.

Do not claim that all providers, Microsoft tenants, or declared Grafana versions passed unless the release evidence records those checks. Dependency findings and skipped validation must remain visible. See [catalog readiness](../CERTIFICATION.md), [testing guidance](../wiki/Testing-and-Review.md), and [release validation](RELEASE_VALIDATION.md) for the evidence required to distinguish prepared work, published assets, and actual Grafana approval.
