# Intelligence Gateway Secure AI data source

This backend data source is the secure provider boundary for the [Grafana Intelligence Gateway panel](https://github.com/digitalrcs/grafana-intelligence-gateway). Grafana encrypts provider credentials in `secureJsonData`; only the Go backend decrypts them and adds them to outbound provider requests. Dashboard JSON and browser code receive neither the stored value nor provider error bodies.

![Secure provider settings](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/secure-provider-settings.png)

## Start here

- [Installation and Quick Start](Installation-and-Quick-Start)
- [Configuration](Configuration)
- [Security and Secrets](Security-and-Secrets)
- [Panel Integration](Panel-Integration)
- [Testing and Review](Testing-and-Review)
- [Grafana Catalog Readiness](Grafana-Catalog-Readiness)
- [Troubleshooting](Troubleshooting)

Plugin ID: `digitalrcs-intelligencegateway-datasource`  
Minimum Grafana version: `11.6.0`
