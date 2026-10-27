# Testing and Review

> **The supplied demo is not an AI service.** **DEMO ONLY - Mock AI (no inference)** returns a fixed delivery receipt. It cannot understand data or images, and no real-provider credentials or hosted model access are included. Complete the two phases below separately.

## Phase 1: deterministic connection review (no credentials)

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

1. **Connections > Data sources > DEMO ONLY - Mock AI (no inference)** shows a prominent demo notice and a real-provider checklist; **Save & test** reports success.
2. **Dashboards > Intelligence Gateway Data Source Review** loads the provisioned data-source query.
3. Running the query returns **MOCK MODE — no AI inference performed** and explains that the fixed receipt does not analyze the input.

The mock provider is isolated to Docker Compose, accepts no credential, logs no request content, and produces a deterministic response. It exists so reviewers do not need DigitalRCS or OpenAI secrets.

## Phase 2: real-provider assessment

For an end-to-end operational use case, follow the [traffic-shift reviewer walkthrough](https://github.com/digitalrcs/grafana-intelligence-gateway/wiki/Reviewer-Walkthrough). It explains the two plugins' responsibilities, supplies synthetic DC1/DC2 request rates, and shows how to connect LM Studio and check that the response changes with the input. The default mock proves connectivity and rendering only, not inference quality.

Use an existing approved provider, or coordinate suitable limited test API access privately with DigitalRCS through the existing Grafana review conversation before this phase. No hosted service or credentials are currently bundled or promised; any access must be arranged and confirmed separately. Do not publish credentials or private endpoint details. The mock check alone does not complete real-provider review.

Keep the mock instance configured; create a **separate Intelligence Gateway Secure AI** instance using the [provider checklist](Installation-and-Quick-Start#phase-2-real-assessment-with-your-own-provider). Configure the agreed reachable provider URL, compatible model, and secure credential if required, then **Save & test**. Select this new instance under the panel's **AI provider → Secure AI data source**, not **Queries**. For images, explicitly enable image analysis and approve an image-capable model.

URLs are reached from the Grafana server/container. LM Studio on the Docker Desktop host can use `http://host.docker.internal:1234/v1` if its server accepts that connection; a remote provider needs its reachable network address. `localhost` is not the Docker host. Hosted access and credentials are not bundled with the demo.

Confirm permitted model discovery and a small synthetic inference request, then verify that dashboard JSON contains the data-source UID but no credential. A successful **Save & test** alone is not an inference result. Do not paste secrets into issues, screenshots, logs, or exported dashboards.

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
