# Testing and Review

## Deterministic Grafana review

```bash
npm ci
npm run typecheck
npm run lint
npm run test:ci
npm run build
mage -v test
mage -v build:linux
docker compose up --build
```

Open <http://localhost:3000>, sign in with `admin` / `admin`, and verify:

1. **Connections > Data sources > Intelligence Gateway Review AI > Save & test** reports success.
2. **Dashboards > Intelligence Gateway Data Source Review** loads the provisioned data-source query.
3. Running the query returns **MOCK MODE — no AI inference performed** and explains that the fixed receipt does not analyze the input.

The mock provider is isolated to Docker Compose, accepts no credential, logs no request content, and produces a deterministic response. It exists so reviewers do not need DigitalRCS or OpenAI secrets.

## Production-provider check

For an end-to-end operational use case, follow the [traffic-shift reviewer walkthrough](https://github.com/digitalrcs/grafana-intelligence-gateway/wiki/Reviewer-Walkthrough). It explains the two plugins' responsibilities, supplies synthetic DC1/DC2 request rates, and shows how to connect LM Studio and check that the response changes with the input. The default mock proves connectivity and rendering only, not inference quality.

Configure a separate instance with a real provider and credential. Confirm model discovery and a small test request, then verify that dashboard JSON contains the data-source UID but no credential. Do not paste secrets into issues, screenshots, logs, or exported dashboards.

## 1.1.0 companion-panel checks

Use version 1.1.0 of both plugins and synthetic, non-sensitive data. The integrated panel repository contains the image UI tests and the dashboard fixture. Keep mock transport tests separate from live-provider inference checks; a receipt proving image delivery does not prove that a model understood an image.

| Check                    | Expected result                                                                                                                               |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Data only                | **Queries** returns the intended rows; **Analyze** sends no images and renders the final answer/explicitly labeled mock receipt.              |
| Selected-panel image     | **Images** lists panel headings rather than text-panel bodies. The selected capture previews before any analysis request.                     |
| Dashboard image          | Capture previews the rendered dashboard, excluding Intelligence Gateway panels; review warnings about unrendered panels.                      |
| Uploaded images          | Valid PNG/JPEG files preview, can be removed, and obey count/size/dimension limits.                                                           |
| Images only              | **Include query data with images** is off; serialized query rows are omitted, while the configured prompt and image evidence remain.          |
| Combined input           | The same request contains images and this Intelligence panel's query results. Other pictured panels' raw values are not implicitly collected. |
| Explicit send and clear  | No image inference request occurs before **Analyze images**. **Clear analysis** clears pending images/errors/answer and cancels active work.  |
| Policy rejection         | Unapproved image models, remote URLs, malformed files, excessive payloads, and unsupported streaming fail before upstream inference.          |
| Token and timeout limits | Lower panel/admin token cap is enforced. Reasoning-only length failures are distinct from timeout failures.                                   |

For a real image test, verify model capabilities with the provider rather than guessing from its name. Ask about an observable property in the synthetic image and verify the answer against the source. Check data-only and combined inference separately, because a correct numeric answer with query rows present is not by itself proof of image understanding. Keep private endpoints, credentials, and real operational data out of public screenshots and release evidence.

Open WebUI and Grafana LLM require their own configured deployments. Copilot Studio preview tests must distinguish protocol/unit coverage from a real tenant-approved published agent test: token exchange, image upload permission, reply correlation, and completion signals must all work. An SSO-requiring agent is outside the preview's supported authentication path. Do not label unrun providers or Grafana versions as passing; record exact commands, versions, results, and limitations in the [release validation record](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/docs/RELEASE_VALIDATION.md).
