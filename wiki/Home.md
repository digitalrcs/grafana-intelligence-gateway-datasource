# Intelligence Gateway Secure AI data source

This backend data source is the secure provider boundary for the [Grafana Intelligence Gateway panel](https://github.com/digitalrcs/grafana-intelligence-gateway). Grafana encrypts provider credentials in `secureJsonData`; only the Go backend decrypts them and adds them to outbound provider requests. Dashboard JSON and browser code receive neither the stored value nor provider error bodies.

Designed specifically for the Intelligence Gateway dashboard experience, the data source supplies model access rather than the metrics being analyzed. Install the two plugins side by side, then select a configured instance in the panel's **AI provider** options. The panel's **Queries** connection remains your metric/table source or Grafana's Dashboard data source.

![Companion panel assessment using the secure AI connection with synthetic data](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/intelligence-gateway-assessment.png)

The screenshot shows the companion panel's output, not a visualization supplied by this data source. Read the [catalog overview](https://github.com/digitalrcs/grafana-intelligence-gateway-datasource/blob/main/src/README.md) for capabilities, setup, and limitations.

## Start here

- [Worked example: traffic assessment and real-provider evidence](https://github.com/digitalrcs/grafana-intelligence-gateway/wiki/Reviewer-Walkthrough)
- [Installation and Quick Start](Installation-and-Quick-Start)
- [Configuration](Configuration)
- [Security and Secrets](Security-and-Secrets)
- [Panel Integration](Panel-Integration)
- [Testing and Review](Testing-and-Review)
- [Grafana Catalog Readiness](Grafana-Catalog-Readiness)
- [Troubleshooting](Troubleshooting)

Plugin ID: `digitalrcs-intelligencegateway-datasource`  
Minimum Grafana version: `11.6.0`
