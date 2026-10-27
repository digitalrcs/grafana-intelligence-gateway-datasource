# Release 1.1.0

The 1.1.0 companion pair adds **data-only**, **images-only**, and **combined data and image** assessment through the required secure AI data source. The companion panel captures the rendered dashboard or selected panels, or accepts PNG/JPEG uploads, and requires preview plus an explicit **Analyze images** action. Capturing a chart does not retrieve its underlying rows: configure the panel's separate **Queries** connection for exact numeric input.

## Data-source changes

- Six provider profiles: OpenAI, LM Studio, Open WebUI, Grafana LLM, Copilot Studio, and Custom.
- Opt-in image policy with a separate approved vision-model list, capability discovery, and bounded PNG/JPEG validation.
- Backend limits for count, decoded size, dimensions, pixels, encoded request size, output tokens, timeouts, and concurrency.
- Separate secure credentials for Grafana LLM service-account access and Copilot Studio Direct Line channel access.
- Copilot Studio **preview**, limited to an approved published agent and correlated text replies with completion signals. Delegated end-user sign-in and enterprise SSO are not supported or certified by this adapter.

## Upgrade and test

Install both 1.1.0 packages side by side in the same Grafana installation, restart Grafana, verify the installed versions, and reload the browser page. Existing text-only use does not require enabling vision or recreating credentials. Enable images only after verifying the selected model/agent supports them, then use the [input-mode walkthrough](Panel-Integration) and [testing checklist](Testing-and-Review).

For reasoning models, review both panel and administrator output limits. The lower cap wins; the soft visible-answer instruction does not reserve an answer budget. Response timeout is a separate limit, and the earliest panel, backend, proxy, or provider deadline still applies. See [Troubleshooting](Troubleshooting).

## Publication and review

The [full release notes](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/docs/RELEASE_NOTES_1.1.0.md) describe scope and upgrade requirements. Check the [release assets](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/releases) and [validation record](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/docs/RELEASE_VALIDATION.md) for actual publication, checksums, source tag, signature state, provenance, test results, and remaining limitations. Documentation, a passing build, or a GitHub release does not establish Grafana catalog approval.
