# Configuration

## Provider policy

| Setting               | Purpose                                                                               |
| --------------------- | ------------------------------------------------------------------------------------- |
| Provider              | `openai`, `lmstudio`, or an OpenAI-compatible `custom` endpoint.                      |
| Base URL              | Administrator-controlled URL up to `/v1`. Dashboard queries cannot replace it.        |
| Default model         | Used when a panel or query omits a model.                                             |
| Allowed models        | Exact permitted model IDs; include the default. Empty permits only the default model. |
| Timeout               | Server-side provider deadline, 1-600 seconds.                                         |
| Maximum output tokens | Administrator ceiling; the lower panel request cap wins.                              |
| Allow insecure HTTP   | Explicitly permits non-TLS provider traffic. Default is off.                          |

![Explicit insecure HTTP override](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/insecure-http-override.png)

HTTPS is required by default. Enable **Allow insecure HTTP** only when an administrator has verified that an organizational endpoint is protected by trusted network controls and TLS is unavailable. The plugin deliberately does not guess whether an arbitrary corporate hostname or address is local. The override can expose credentials and prompts in transit; OpenAI remains restricted to `api.openai.com`.

## Credentials

Enter only the API key or bearer token required by the provider. A bearer token takes precedence over an API key. OAuth client-credentials exchange and streaming are not supported; responses are buffered.

After save, Grafana returns only configured/reset flags through `secureJsonFields`. To replace a secret, select its reset control, enter the new value, and save again.

## Provisioning

```yaml
apiVersion: 1

datasources:
  - name: Intelligence Gateway Secure AI
    uid: intelligence-gateway-secure
    type: digitalrcs-intelligencegateway-datasource
    access: proxy
    jsonData:
      provider: openai
      baseUrl: https://api.openai.com/v1
      defaultModel: gpt-4.1-mini
      timeoutSeconds: 300
      allowedModels:
        - gpt-4.1-mini
      maxOutputTokens: 256000
      allowInsecureHttp: false
    secureJsonData:
      apiKey: ${OPENAI_API_KEY}
```

Set `OPENAI_API_KEY` in the Grafana server or container secret environment, not in a dashboard, committed YAML, or browser panel option.

## Upgrading from 1.0.0

An empty allowed-model list now permits only the configured default, not every provider model. Add any additional model IDs used by your panels and include the default model. Model discovery returns only permitted IDs and removes provider-specific metadata.

The nonfunctional OAuth client-secret control and streaming switch have been removed. Legacy `allowStreaming` and `clientSecret` configuration values are ignored; streaming requests are always rejected. Remove unused stored client secrets through provisioning or Grafana's data-source API. The explicit **Allow insecure HTTP** override is unchanged.
