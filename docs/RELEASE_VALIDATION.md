# Release validation

## 1.1.2 security release — October 30, 2026

Grafana review ticket 240360 requests removal of the two binary dependency warnings in 1.1.1. This release updates `golang.org/x/net` to 0.62.0, which fixes GO-2026-6860 and GO-2026-6866, and resolves its required `golang.org/x/sys` 0.49.0 and `golang.org/x/text` 0.43.0 dependencies. Go 1.26.10, Grafana Go SDK 0.296.5, application source, and npm dependencies are unchanged. The panel remains at its signed 1.1.1 release.

CI and the tagged release now run `govulncheck@v1.8.0 -mode=binary -scan=module` in text mode against all six backend targets. The release check scans executables extracted from the exact signed ZIP; `set -euo pipefail` makes module findings fail the workflow even if a later validator would classify them as warnings. Source reachability analysis remains an additional check, not a substitute for binary dependency remediation. JSON scanner output is not used as an exit-code gate.

Confirm the new source scan, backend regression tests, six binary scans, and full signed-archive validator results before publication. Preserve 1.1.1 evidence as historical; its absence of reachable HTML calls did not satisfy the reviewer's request to update the embedded module. Use the newly generated ZIP and SHA1 for the existing datasource submission; do not resubmit the old 1.1.1 ZIP or change the panel submission solely for this fix.

## 1.1.1 signing release — October 30, 2026

Grafana has requested signed builds of both companion plugins. Version 1.1.1 changes release metadata and signing safeguards only; the reviewed 1.1.0 plugin behavior and dependency graph are unchanged. The release workflow requires the public-signing gate and token, checks for a signed manifest in the downloaded archive, and retains exact-archive provenance and full validator checks. Verify the actual signature classification and archive hashes before publication and submission; signing is not itself catalog publication.

Grafana's recommended SDK HTTP-client integration and configuration-default persistence remain separate future work, explicitly non-blocking in the review. They are not included in this signing release.

The October 30 full npm audit reports **five moderate and seven low affected package entries, zero high, and zero critical**. The production-only audit reports the same **four moderate React Router package entries** described below. The additional moderate [qs advisory](https://github.com/advisories/GHSA-6pq8-jhm2-2vcw) affects the development-only `tiny-lr` chain; the low [deepmerge advisory](https://github.com/advisories/GHSA-gcqx-w48m-w2h4) affects development test/type-check tooling. These tools are not shipped as plugin runtime implementations. Both high-severity audit gates pass; findings are not suppressed.

The fresh Go 1.26.10 scan with `govulncheck@v1.8.0` reports **zero reachable vulnerabilities and zero vulnerable imported packages**. Two `golang.org/x/net/html` module-only advisories, [GO-2026-6866](https://pkg.go.dev/vuln/GO-2026-6866) and [GO-2026-6860](https://pkg.go.dev/vuln/GO-2026-6860), are neither imported nor called by this plugin. This is not a claim that the entire dependency graph is advisory-free. The October 26 results below remain dated historical evidence, not today's scan totals.

The dated 1.1.0 validation record below remains historical evidence. The [1.1.0 submission handoff](GRAFANA_SUBMISSION.md) describes the older unsigned archives; use the new signed release's URLs and checksums for the signing update.

## Dependency assessment — October 26, 2026

Both repositories now pin the development signing tool's `get-uri > basic-ftp` dependency to 6.2.1 and the host-external `@grafana/data > moment` dependency to 2.31.0. This removes the current high-severity FTP directory-list parser advisory and the Moment findings without adopting sign-plugin 4 (which requires a newer Node runtime). The narrow FTP override retains the Client API used by get-uri; the signing-tool dependency import is smoke-tested. No forced broad audit fix is used.

The full npm audit reports **four moderate affected package entries, zero high, and zero critical**. The remaining React Router compatibility chain represents two advisories:

- [GHSA-wrjc-x8rr-h8h6](https://github.com/remix-run/react-router/security/advisories/GHSA-wrjc-x8rr-h8h6): external redirects from untrusted navigation paths.
- [GHSA-337j-9hxr-rhxg](https://github.com/remix-run/react-router/security/advisories/GHSA-337j-9hxr-rhxg): SSR error-hydration constructor injection.

The plugin does not import routing APIs or implement SSR hydration. Grafana and React are host-provided externals, not bundled router implementations. This is an archive-scope assessment, not a guarantee about Grafana hosts; administrators must keep Grafana patched. CI checks the full dependency graph at high severity and above, including development tools. The remaining findings are not suppressed.

Other narrow lockfile updates include DOMPurify, zod, fast-uri, source-map-js, ip-address, postcss-selector-parser, and serialize-javascript. No model/provider credentials are required for the checks.

## Preflight and scope

The feature includes tests for input bounds, explicit preview/send behavior, image-only versus combined requests, image-model approval, invalid image data, capture naming, and desktop/narrow checkbox layout. The secure companion also tests Direct Line protocol handling, JSON-only responses, fail-fast concurrency, and response-body slot ownership.

The initial fresh Go scan found 12 reachable vulnerabilities in the prior Go 1.26.8 toolchain and x/net 0.59.0. The release updates to **Go 1.26.10** and **golang.org/x/net 0.60.0**, retaining Grafana Go SDK 0.296.5. The patched `govulncheck@v1.8.0 ./pkg/...` scan reports **No vulnerabilities found**. This is a dated scan result, not a guarantee about future advisories.

Full native Go tests, Go vet, and repeated slow-body concurrency regressions passed before the toolchain patch. The complete patched-toolchain build/test/scan gate runs in CI, including race-enabled Mage tests. Check those results on the release commit.

Local build/unit checks are complemented by CI on released Grafana versions. Read the [Actions results](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/actions) for the exact release commit and tested matrix; do not infer a minimum-version test from plugin metadata. Native Windows uses Node 26, which exceeds plugin-e2e's supported Node range; CI uses Node 22 and is the clean-install browser gate.

The isolated Grafana 13.1.1 review environment previously verified capture/preview, explicit send, clear/remove, data-only inference, and checkbox wrapping at desktop and narrow widths. The current user's live LM Studio instance is not used for automatic mock tests or overwritten during release preparation. CI uses only synthetic data and the deterministic mock.

## Real-provider limitations

A real LM Studio text-only assessment of synthetic traffic completed in the local test environment after increasing the output budget. That does not prove image understanding. Mock vision tests verify transport/policy/UI, not real inference quality. Live image-understanding tests for all providers, Grafana LLM app deployments, and Copilot Studio tenants are not claimed. Copilot remains a bounded Direct Line preview, without delegated enterprise SSO.

Capture uses rendered dashboard content. Lazy panel loading or scheduled dashboard refresh may still query data; no universal zero-query guarantee is made. Review previews for omitted panels and sensitive content.

## Exact release archive gate

The default Docker review image now uses Grafana's official entrypoint in production application mode with only the specific review plugin ID allowed unsigned. It does not inject a missing development live-reload script. This local fixture's optional anonymous administrator access is not a production deployment recommendation.

The tagged Release workflow downloads its generated ZIP and SHA1 asset, verifies their agreement, verifies GitHub provenance against the repository, release workflow and exact tag commit, then runs the full default Grafana Plugin Validator Docker image with the matching source tag. It preserves the validator image digest and results as the `release-validation` Actions artifact.

Before publishing/submitting, inspect that evidence for packaging, source, security, dependency, provenance, and signature findings. An expected unsigned-review warning is not permission to overlook other failures. AI-assisted validator reviews require a separately configured provider and are not represented as performed without one.

The 1.1.0 archives were published unsigned for initial review. The 1.1.1 release requires public signing after Grafana's review; an unsigned archive is no longer an acceptable release result. Provenance and a Grafana signature are independent checks.

Verify public download/checksum and asset contents again after publication. Record the exact public asset hashes in the submission handoff rather than guessing them before the release is built. GitHub Wiki publication is a separate repository push; committing `wiki/` alone is insufficient.
