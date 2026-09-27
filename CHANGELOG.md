# Changelog

## Unreleased

- Reworked the packaged catalog overview around the Intelligence Gateway companion workflow, with real assessment imagery and documentation/support links.
- Clarified side-by-side plugin installation, same-organization instance setup, and separate query-data and AI-provider connections.
- Made local ZIP filenames follow built plugin metadata and added preflight validation plus automated packaging tests.

## 1.0.0 - 2026-08-12

- Initial secure backend data-source implementation.
- Added encrypted API key, bearer token, and reserved OAuth client-secret configuration.
- Added OpenAI, LM Studio, and custom OpenAI-compatible provider policies.
- Added model discovery, chat-completions/analyze resources, and a query-to-answer data frame path.
- Added server-side host, TLS, redirect, model, token, timeout, body-size, concurrency, rate, and error-redaction controls.
- Added an administrator-controlled `allowInsecureHttp` override for organizational provider networks that cannot be classified by hostname or IP range.
- Added a deterministic provisioned review environment, catalog screenshots, release provenance, and submission-readiness documentation.
