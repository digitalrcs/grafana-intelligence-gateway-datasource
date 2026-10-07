# Grafana catalog readiness

This repository is prepared for Grafana's automated validation and manual plugin review. Approval and signature classification remain decisions made by Grafana Labs.

## Evidence checklist

- Public source repository with Apache-2.0 license and `Copyright 2026 DigitalRCS`.
- Official `@grafana/create-plugin` structure with React frontend and Go backend.
- Complete `plugin.json` metadata, logos, catalog screenshots, minimum Grafana version, source, issues, and project links.
- Credentials stored only in `secureJsonData`; no analytics or tracking code.
- Server-side URL, model, token, timeout, body-size, redirect, concurrency, rate, and error-redaction policies.
- Typecheck, lint, frontend build, Go tests/build, and multi-version Grafana E2E in GitHub Actions.
- Provisioned Docker review environment with a deterministic mock provider and sample dashboard.
- Official Grafana release workflow with provenance attestation. Initial review archives remain unsigned until Grafana grants a public signature level.
- ZIP packaging preserves the top-level plugin directory and executable `0755` backend binaries.

## Submission fields

After publishing a version tag and its generated GitHub release:

- **OS & Architecture:** Single archive containing all generated backend binaries.
- **URL:** Direct URL to the release ZIP.
- **Source code URL:** Version tag URL, not a moving branch.
- **SHA1:** Value in the release `.sha1` asset.
- **Provisioning provided:** Yes.
- **Testing guidance:** Use the steps in [Testing and Review](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Testing-and-Review).

The first review submission may be unsigned. Configure `GRAFANA_PUBLIC_SIGNING_ENABLED=true` and the `GRAFANA_ACCESS_POLICY_TOKEN` repository secret only after Grafana grants the public signature level.
