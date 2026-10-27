# Grafana catalog readiness

This checklist supports preparation of the 1.1.0 companion-plugin update for Grafana's automated validation and manual review. It is not an assertion that 1.1.0 has been published, signed, submitted, or approved. Record actual results in [release validation](docs/RELEASE_VALIDATION.md). Approval and signature classification remain decisions made by Grafana Labs.

## Evidence checklist

- Public source repository with Apache-2.0 license and `Copyright 2026 DigitalRCS`.
- Official `@grafana/create-plugin` structure with React frontend and Go backend.
- Complete `plugin.json` metadata, logos, catalog screenshots, minimum Grafana version, source, issues, and project links.
- Credentials stored only in `secureJsonData`; no analytics or tracking code.
- Server-side URL, model, token, timeout, body-size, redirect, concurrency, rate, and error-redaction policies.
- Typecheck, lint, frontend build, Go tests/build, and multi-version Grafana E2E in GitHub Actions.
- Provisioned Docker review environment with a deterministic mock provider and sample dashboard.
- Official Grafana release workflow configured for provenance attestation. Verify the resulting attestation against the exact downloaded archive and immutable source tag; a workflow setting alone is not evidence of a successful attestation. Initial review archives may remain unsigned until Grafana grants a public signature level.
- ZIP packaging preserves the top-level plugin directory and executable `0755` backend binaries.

## 1.1.0 verification gates

- Build and test the exact release commit, including backend security tests and the companion panel's data-only, images-only, and combined-input workflows. Record the actual Grafana versions tested, including the declared minimum; do not infer a compatibility pass from metadata.
- Re-run dependency and Go vulnerability scans and document unresolved advisories and their scope. Do not carry forward a historical clean-scan statement as a current result.
- Run Grafana's plugin validator on the final ZIP. Distinguish expected initial-review unsigned status from packaging, metadata, security, or provenance failures; do not suppress unexplained failures.
- Verify the ZIP contains the packaged catalog README, license, changelog, referenced screenshots/logos, matching `plugin.json`, and platform-appropriate executable backends. Check the checksum after downloading the public asset, not only before upload.
- Verify build provenance identifies this repository's release workflow and the same immutable tag/commit as the submitted source. Confirm the signature state of that exact ZIP separately; provenance is not a Grafana signature.
- Publish updated documentation and the Wiki source to their respective GitHub repositories. A commit to the main repository's `wiki/` directory alone does not publish the GitHub Wiki.
- Include both plugin versions and their separate installation instructions in testing guidance. Image controls require the 1.1.0 companion panel and this data source's opt-in vision policy.
- Disclose the bounded Copilot Studio Direct Line preview, lack of delegated end-user SSO, and any unverified live-provider/tenant behavior. Do not represent protocol tests as Microsoft tenant certification.

## Policy classification to confirm

Grafana assigns Community or Commercial signatures based on the author, intended use, and dependent technologies. DigitalRCS is a business and the plugin can connect to commercially backed AI providers, so DigitalRCS should answer the submission questionnaire literally and ask Grafana whether a Commercial Plugin Subscription is required. Do not represent the plugin as Grafana-certified until review is complete.

## Submission fields

After publishing a version tag and its generated GitHub release:

- **OS & Architecture:** Single archive containing all generated backend binaries.
- **URL:** Direct URL to the release ZIP.
- **Source code URL:** Version tag URL, not a moving branch.
- **SHA1:** Value in the release `.sha1` asset.
- **Provisioning provided:** Yes.
- **Testing guidance:** Use the steps in [Testing and Review](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Testing-and-Review).

The first review submission may be unsigned. Configure `GRAFANA_PUBLIC_SIGNING_ENABLED=true` and the `GRAFANA_ACCESS_POLICY_TOKEN` repository secret only after Grafana grants the public signature level.

## Official references

- [Publish or update a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin): archive URL, source URL, SHA1, provisioning, and review process.
- [Package a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin): archive layout and executable backend permissions.
- [Automate builds and attest provenance](https://grafana.com/developers/plugin-tools/publish-a-plugin/build-automation#attest-provenance-for-plugin-builds): release workflow and attestation setup.
- [Publishing best practices](https://grafana.com/developers/plugin-tools/publish-a-plugin/publishing-best-practices): catalog material, compatibility testing, and review expectations.
