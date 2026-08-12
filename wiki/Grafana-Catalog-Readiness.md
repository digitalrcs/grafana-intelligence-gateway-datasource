# Grafana Catalog Readiness

The repository includes catalog metadata and screenshots, Apache-2.0 licensing, CI and multi-version E2E checks, a deterministic reviewer environment, release provenance attestation, a correctly rooted multi-architecture ZIP, and an SHA1 release asset. The full evidence checklist is maintained in [`CERTIFICATION.md`](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/CERTIFICATION.md).

Initial review archives remain unsigned until Grafana grants a public signature level. Grafana Labs makes the final approval and signature-classification decision. DigitalRCS should answer the submission questionnaire literally and confirm whether the plugin requires a Commercial Plugin Subscription.

The HTTP override is default-off and intended for controlled organizational networks where address classification cannot reliably establish locality. Because Grafana policy prefers TLS, disclose this behavior in the testing guidance and allow the reviewer to assess it.
