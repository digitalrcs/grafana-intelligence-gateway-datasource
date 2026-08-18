# Configuration

## Provider policy

| Setting               | Purpose                                                                              |
| --------------------- | ------------------------------------------------------------------------------------ |
| Provider              | `openai`, `lmstudio`, or an OpenAI-compatible `custom` endpoint.                     |
| Base URL              | Administrator-controlled URL up to `/v1`. Dashboard queries cannot replace it.       |
| Default model         | Used when a panel or query omits a model.                                            |
| Allowed models        | Administrator allow-list. Include the default model; when empty, only the default is permitted. |
| Timeout               | Server-side provider deadline, 1-600 seconds.                                        |
| Maximum output tokens | Administrator ceiling; the lower panel request cap wins.                             |
| Allow insecure HTTP   | Explicitly permits non-TLS provider traffic. Default is off.                         |

![Explicit insecure HTTP override](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/insecure-http-override.png)

HTTPS is required by default. Enable **Allow insecure HTTP** only when an administrator has verified that an organizational endpoint is protected by trusted network controls and TLS is unavailable. The plugin deliberately does not guess whether an arbitrary corporate hostname or address is local. The override can expose credentials and prompts in transit; OpenAI remains restricted to `api.openai.com`.

## Credentials

Enter only the credential required by the provider. A bearer token takes precedence over an API key. Unsupported or unused secret fields are not stored.

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
