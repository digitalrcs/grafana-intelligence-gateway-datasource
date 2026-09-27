# 1.0.1 release validation

This maintenance release reconciles the backend-hardening work with the packaged catalog overview. It does not represent Grafana approval, certification, or a published catalog listing.

## Security dependency assessment — September 27, 2026

The npm lockfile has no high or critical findings after aligning the Grafana packages to 13.1.1 and patching DOMPurify, react-use/js-cookie, js-yaml, fast-uri, qs, and nanoid. React 18 is retained for compatibility.

Four moderate affected package entries remain in the inherited React Router dependency chain. They represent two advisories, not four separate vulnerabilities:

- [GHSA-wrjc-x8rr-h8h6](https://github.com/remix-run/react-router/security/advisories/GHSA-wrjc-x8rr-h8h6): untrusted navigation paths can cause external redirects.
- [GHSA-337j-9hxr-rhxg](https://github.com/remix-run/react-router/security/advisories/GHSA-337j-9hxr-rhxg): constructor injection in SSR error hydration.

The plugin does not import routing APIs, implement SSR hydration, or bundle these libraries. Grafana/React modules are supplied by the Grafana host. This is an assessment of the plugin archive, not a claim that every supported Grafana host is free of those issues: administrators must keep Grafana patched. Forcing Router 7 into the host's Router 6 compatibility package is not a safe maintenance fix. Reassess this finding when upgrading Grafana dependencies.

CI checks production npm advisories at high severity and above. Full npm audit still reports the documented moderate findings; no advisory suppression or forced breaking override is used.

The baseline Go scan found eight reachable vulnerabilities: six in the Go standard library and two in gRPC. The backend build moves from Go 1.26.5 to 1.26.8 and from Grafana's Go SDK 0.296.2 to 0.296.5, including gRPC 1.83.2. The module minimum and release workflow must stay aligned so published binaries use the patched toolchain.

The final `govulncheck@v1.8.0 -show verbose ./pkg/...` scan reported **No vulnerabilities found** across all three production root packages, 69 modules, and the Go 1.26.8 standard library. `go mod verify` passed. CI repeats the pinned source scan. This is a dated scan result, not a guarantee about future advisories.

## Upgrade and behavior checks

- An empty model allow-list permits only the default model; a nonempty list must include it.
- Model discovery returns approved IDs only, with no upstream model metadata or headers.
- OAuth client-secret and streaming controls are removed. Legacy values are ignored; streaming is rejected even with a legacy enabling value.
- The explicit insecure-HTTP override remains supported, defaults off, and displays its transport warning.
- API key/bearer credentials remain in Grafana secure settings. Mock receipts remain explicitly labeled as not AI inference.

See [Configuration](../wiki/Configuration.md#upgrading-from-100) before upgrading an instance that used an unrestricted empty model list.

## Release verification

The release workflow builds all supported backend targets, packages the catalog README/assets, and produces the ZIP, SHA1 asset, and GitHub build provenance. Validate the exact downloadable archive and source tag before submitting it to Grafana. An initial-review unsigned warning is expected until Grafana grants a public signature level; do not bypass signing checks for a production-signed release.

The local preflight and final release results are recorded in the release pull request and GitHub release. Passing automated validation does not replace Grafana's manual review or coordinate publication of the companion panel dependency automatically.
