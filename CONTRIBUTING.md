# Contributing

Issues and pull requests are welcome. Keep changes focused, avoid including credentials or real prompts, and preserve the server-side secret boundary.

Before submitting a change, run:

```bash
npm ci
npm run typecheck
npm run lint
npm run test:ci
npm run build
mage -v test
mage -v build:linux
```

For browser tests, build the frontend and Linux backend, then run `docker compose up -d` followed by `npm run e2e`. The provisioned mock provider is deterministic and requires no external credential.

Security reports belong in private GitHub Security Advisories, not public issues.
