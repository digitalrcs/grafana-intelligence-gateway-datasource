# Grafana submission update — 1.1.0

Use **Submit Update** on the existing plugin submission. Submit the secure data source first, then the panel; they are separate packages. The fields below target the 1.1.0 release. Confirm publication and the exact archive verification before entering them in Grafana. Preparing these fields does not update the portal or grant approval.

| Field                 | Value                                                                                                                                                   |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Plugin ID             | `digitalrcs-intelligencegateway-datasource`                                                                                                             |
| Version               | `1.1.0`                                                                                                                                                 |
| Type                  | Data source with backend                                                                                                                                |
| OS & Architecture     | Single (one archive containing all generated backend architectures)                                                                                     |
| ZIP URL               | https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/releases/download/v1.1.0/digitalrcs-intelligencegateway-datasource-1.1.0.zip      |
| Source URL            | https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/tree/v1.1.0                                                                       |
| SHA1 asset            | https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/releases/download/v1.1.0/digitalrcs-intelligencegateway-datasource-1.1.0.zip.sha1 |
| Provisioning provided | Yes                                                                                                                                                     |
| Minimum Grafana       | 11.6.0; optional Grafana LLM app requires Grafana 13+                                                                                                   |
| License               | Apache-2.0                                                                                                                                              |

Copy the **hash value** from the SHA1 asset into Grafana's SHA1 field, not the checksum-file URL. Do not substitute GitHub's automatically generated source-code ZIP for the packaged plugin.

## Testing guidance

Install both Intelligence Gateway 1.1.0 packages side by side. Follow the [panel reviewer walkthrough](https://github.com/digitalrcs/grafana-intelligence-gateway/wiki/Reviewer-Walkthrough) and [data-source test guide](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Testing-and-Review). The provisioned Docker environment uses a deterministic mock and synthetic traffic data without external credentials. Its labeled receipt verifies transport, not AI understanding.

> **DEMO ONLY - Mock AI (no inference)** is the supplied test connection. It runs no model and cannot assess data or images. No cloud credentials or hosted AI access are shipped. A mock receipt must not be counted as genuine inference.

Review in two phases:

1. **Mock delivery check:** keep the provisioned mock instance, verify its prominent settings notice, select **Save & test**, and run the synthetic fixture. Expect a labeled receipt only.
2. **Real assessment:** the reviewer supplies a reachable provider and compatible model. Create a **separate Intelligence Gateway Secure AI** instance, configure the provider/base URL/model and required secure credential, then **Save & test**. Enable image analysis and approve a vision model for image tests. Select the new instance under the panel's **AI provider → Secure AI data source**, not **Queries**. Keep the mock available for repeatable transport tests.

The URL is accessed by the Grafana server/container, not the reviewer's browser. For LM Studio on the Docker Desktop host, `http://host.docker.internal:1234/v1` is an example when the server accepts container connections; use a reachable network address for a remote host. A trusted-network HTTP override is required for HTTP. **Save & test** checks connectivity/model access or Direct Line token exchange, not inference. See the [complete own-provider checklist](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/wiki/Installation-and-Quick-Start#phase-2-real-assessment-with-your-own-provider).

Verify all three input modes:

1. **Data only:** configure the panel's Queries using Grafana's Dashboard data source to reuse the source chart; select **Analyze** without attaching images.
2. **Images only:** enable vision on the secure AI data source and approve the model. In the panel, turn off **Include query data with images**; no source panel query is required. Capture or upload, review previews, then select **Analyze images**.
3. **Images and data:** enable **Include query data with images**, configure a source query, then explicitly send the reviewed images and bounded query results.

Confirm previews do not send inference requests, the Intelligence panels are excluded from capture, names wrap beside their checkboxes, Clear analysis removes queued images/results, and invalid images/model policies fail safely. Use a separate real provider with synthetic input to assess image understanding. Live validation of every provider and Microsoft tenant is not implied. Copilot Studio is a bounded Direct Line preview; delegated enterprise SSO is not implemented.

## Verify the exact release

```bash
gh release download v1.1.0 --repo digitalrcs/grafana-intelligence-gateway-datasource \
  --pattern 'digitalrcs-intelligencegateway-datasource-1.1.0.zip*'
sha1sum digitalrcs-intelligencegateway-datasource-1.1.0.zip
sha256sum digitalrcs-intelligencegateway-datasource-1.1.0.zip
gh attestation verify digitalrcs-intelligencegateway-datasource-1.1.0.zip \
  --repo digitalrcs/grafana-intelligence-gateway-datasource \
  --signer-workflow digitalrcs/grafana-intelligence-gateway-datasource/.github/workflows/release.yml \
  --source-ref refs/tags/v1.1.0
```

The Release workflow verifies the downloaded artifact, source-tag provenance, checksum, and full Grafana Plugin Validator output. Its evidence artifact records results and the validator image digest. Check those results rather than assuming the workflow configuration proves success.

The ZIP must contain one `digitalrcs-intelligencegateway-datasource/` root, version 1.1.0, production module, catalog README, changelog, license, and all referenced image assets. Verify backend binaries are included with portable executable permissions.

These repositories currently have no public-signing token: the release is intended as an **unsigned review archive**. Build provenance is not a Grafana signature. Public installation without an unsigned allow-list requires Grafana's signature assignment and a properly signed build. Do not represent this as certified or catalog-approved.

See [release notes](RELEASE_NOTES_1.1.0.md), [release validation](RELEASE_VALIDATION.md), and the [GitHub release](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/releases/tag/v1.1.0). Portal submission remains a separate administrator action.
