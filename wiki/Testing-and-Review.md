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
