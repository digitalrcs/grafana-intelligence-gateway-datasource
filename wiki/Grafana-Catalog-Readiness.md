# Grafana Catalog Readiness

The repository contains catalog metadata and screenshots, Apache-2.0 licensing, CI and multi-version E2E workflows, a deterministic reviewer environment, and release automation configured to produce provenance attestations. These are preparation mechanisms, not evidence that every check has passed for a new release. The full checklist is maintained in [`CERTIFICATION.md`](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/CERTIFICATION.md).

For 1.1.0, validate the exact downloadable ZIP, matching source tag, checksum, backend executable modes, signature state, and build provenance. Record final results in the [release validation record](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/docs/RELEASE_VALIDATION.md). Publish the companion panel independently and verify its required data-source dependency is available; the two plugins remain separate archives and submissions. See [Release 1.1.0](Release-1.1.0) for feature and upgrade scope.

Initial review archives remain unsigned until Grafana grants a public signature level. Grafana Labs makes the final approval and signature-classification decision. DigitalRCS should answer the submission questionnaire literally and confirm whether the plugin requires a Commercial Plugin Subscription.

The HTTP override is default-off and intended for controlled organizational networks where address classification cannot reliably establish locality. Because Grafana policy prefers TLS, disclose this behavior in the testing guidance and allow the reviewer to assess it.

Describe Copilot Studio as a bounded Direct Line preview, not enterprise-SSO certification or proof of compatibility with a particular tenant. Also disclose untested providers, Grafana versions, and unresolved dependency advisories. A passing validator, successful build attestation, or public GitHub release does not replace Grafana's manual review.
