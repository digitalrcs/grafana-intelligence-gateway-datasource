# Security policy

Report suspected vulnerabilities privately through this repository's GitHub Security Advisories. Do not open a public issue containing credentials, prompts, provider responses, private endpoint details, or an exploitable payload.

The plugin stores provider credentials only in Grafana `secureJsonData`. The Go backend decrypts them server-side, enforces configured provider/model/token/timeout/body/rate policies, rejects redirects, and returns sanitized provider errors. Prompts, responses, and credentials are not logged by default.

HTTPS is required by default. **Allow insecure HTTP** is an explicit administrator override for trusted organizational networks whose endpoint locality cannot be classified reliably. It can expose credentials and prompts in transit; leave it disabled unless the network risk is understood and controlled.

Security fixes are supported on the latest released version. Include the plugin version, Grafana version, deployment model, and safe reproduction details in a private report.
