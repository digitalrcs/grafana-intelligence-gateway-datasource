# Changelog

## 1.0.1 - 2026-09-27

- Reworked the packaged catalog overview around the Intelligence Gateway companion workflow, with real assessment imagery and documentation/support links.
- Clarified side-by-side plugin installation, same-organization instance setup, and separate query-data and AI-provider connections.
- Made local ZIP filenames follow built plugin metadata and added preflight validation plus automated packaging tests.
- Restricted model discovery to permitted IDs and removed provider-specific metadata. Empty allow-lists now permit only the configured default; nonempty lists must include it.
- Removed unused OAuth client-secret and streaming controls. Legacy values are ignored and streaming requests are rejected. The explicit insecure-HTTP override remains supported.
- Replaced the provider dropdown with a three-option radio group to avoid a configuration-editor crash when changing providers on older supported Grafana versions.
- Patched frontend dependencies while retaining React 18 compatibility. See `docs/RELEASE_VALIDATION.md` for the remaining host-external router advisory assessment.
- Updated the backend build toolchain to Go 1.26.8 and Grafana's Go SDK to 0.296.5 to address reachable standard-library and gRPC vulnerabilities.

Upgrade note: list additional model IDs explicitly if your 1.0.0 installation relied on an empty allow-list accepting any model. Authentication remains API key or bearer token. This is an unsigned catalog-review build, not Grafana approval or certification.

## 1.0.0 - 2026-08-12

- Initial secure backend data-source implementation.
- Added encrypted API key, bearer token, and reserved OAuth client-secret configuration.
- Added OpenAI, LM Studio, and custom OpenAI-compatible provider policies.
- Added model discovery, chat-completions/analyze resources, and a query-to-answer data frame path.
- Added server-side host, TLS, redirect, model, token, timeout, body-size, concurrency, rate, and error-redaction controls.
- Added an administrator-controlled `allowInsecureHttp` override for organizational provider networks that cannot be classified by hostname or IP range.
- Added a deterministic provisioned review environment, catalog screenshots, release provenance, and submission-readiness documentation.
