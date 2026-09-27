# Catalog content and companion integration

## Which README is published?

Maintain the customer-facing overview in [`src/README.md`](../src/README.md). The unmodified Grafana build configuration prefers that file over the repository-root README and copies it to `dist/README.md`. The release ZIP must contain it at `digitalrcs-intelligencegateway-datasource/README.md` alongside `plugin.json` and `module.js`.

The root README covers development and the backend contract. The wiki covers setup and troubleshooting. Neither replaces the packaged catalog README. Use absolute links in the catalog copy so they work outside GitHub.

Grafana extracts catalog page content from the submitted archive. A GitHub merge alone does not update an existing catalog submission or release. A new versioned build containing the changed README/assets must be submitted; do not replace an already-published archive. See the [catalog update FAQ](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-faqs#how-can-i-update-my-plugins-catalog-page).

## Companion relationship

- The panel is `digitalrcs-intelligencegateway-panel`, displayed as **Grafana Intelligence Gateway**.
- This data source is `digitalrcs-intelligencegateway-datasource`, displayed as **Intelligence Gateway Secure AI**.
- The panel's `dependencies.plugins` already declares this data source with its exact ID, `type: datasource`, and display name. No change to that declaration is needed for this documentation update.
- The data source does not require panel code to perform provider requests, and can return an answer through its own query editor. Keep its `dependencies.plugins` empty; adding a reverse dependency would create an unnecessary cycle.
- Keep the existing plugin IDs/types. Describe the relationship through the catalog overview, short description, documentation links, and installation steps—not by nesting plugin packages or inventing a parent-plugin field.

Grafana's current [installer](https://github.com/grafana/grafana/blob/main/pkg/plugins/manager/installer.go) can fetch declared dependencies from the plugin repository. The declaration therefore must not be treated as merely descriptive. Coordinate catalog availability of the data source before or alongside the panel; an unavailable dependency can prevent catalog installation. Installation does not create an instance, configure its credentials, or select it in dashboards.

During review, use the separate supplied archives or the provisioned sibling-source Docker environment. The [installation guide](../wiki/Installation-and-Quick-Start.md) shows correct sibling plugin directories and the distinction between the panel's metric/query connection and its AI provider connection.

## What is included

| Item                     | Source / verification                                                                                                                                                                                                                                                 |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| User-facing catalog copy | `src/README.md`: purpose, both-plugin setup, provider options, privacy, limits, support, and license.                                                                                                                                                                 |
| Discoverability          | `src/plugin.json`: companion description, keywords, documentation/repository/issues/license links, and exact plugin identity.                                                                                                                                         |
| Screenshot               | `src/img/intelligence-gateway-assessment.png`: existing genuine synthetic-data LM Studio capture, copied unchanged from the companion panel. [Capture provenance](https://github.com/digitalrcs/grafana-intelligence-gateway/blob/main/docs/LIVE_REVIEW_EVIDENCE.md). |
| Brand asset              | `docs/images/digitalrcs-logo.webp`: existing DigitalRCS asset reused from the panel documentation.                                                                                                                                                                    |
| Packaging                | `scripts/package-plugin.py`: derives filename from built metadata, checks required content, and preserves executable permissions.                                                                                                                                     |
| Review environment       | Provisioned mock data source/dashboard; [Testing and Review](../wiki/Testing-and-Review.md) distinguishes connectivity from actual inference.                                                                                                                         |
| Build/release automation | Existing CI builds/tests frontend and backend; release workflow creates versioned archives and provenance.                                                                                                                                                            |

The older provider-policy screenshot shows an HTTP URL without the override enabled. It is retained as a historical repository asset but is no longer listed as a catalog screenshot. The enabled-override and companion-selector screenshots remain in the catalog gallery.

## Readiness status — September 27, 2026

The published data-source `v1.0.0` archive has the correct plugin-ID root, README/license/changelog, frontend assets, three original screenshots, and six backend binaries with executable permissions. Its SHA1 asset and GitHub provenance are present. It still contains the previous README, not this catalog update, and is unsigned as an initial review build.

The 1.0.1 release work resolves the technical findings from the catalog audit:

1. The changes proposed in the older draft [PR #11](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/pull/11) are reconciled with current catalog content and broader backend regression tests, rather than merging its obsolete packaging and dependency changes.
2. The editor no longer offers a nonfunctional OAuth client-secret field or streaming switch. Legacy settings are ignored and streaming requests rejected. Model discovery exposes only approved IDs; an empty list allows only the default, and a nonempty list must include it. See the [upgrade notes](../wiki/Configuration.md#upgrading-from-100).
3. High-severity npm dependency findings are patched. The remaining moderate host-external router findings and release verification boundaries are documented in [Release validation](RELEASE_VALIDATION.md). Do not describe a nonzero full npm audit as completely clean.

Release/submission gates remain:

- Publish a **new versioned release**, then validate its exact ZIP, source revision, SHA1, and provenance before submission. The local packaging helper is not a substitute for the release workflow's multi-platform build, signing, or attestation.
- Coordinate the two catalog submissions/dependency availability with Grafana. Public catalog lookups did not resolve either ID during this audit. No listing, approval, signature classification, or production certification is inferred from successful builds.

Before submitting, run the full plugin validator against the **new release archive**, not only the `metadatavalid` analyzer used in ordinary CI. Verify the provider configuration and intended panel workflow on the build being submitted. Existing release and screenshot evidence is useful provenance, not proof of every future build.

Official references: [metadata/dependencies](https://grafana.com/developers/plugin-tools/reference/plugin-json), [packaging](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin), [publishing best practices](https://grafana.com/developers/plugin-tools/publish-a-plugin/publishing-best-practices), and [submission requirements](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin).
