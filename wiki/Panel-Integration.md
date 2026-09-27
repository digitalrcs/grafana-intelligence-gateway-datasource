# Panel Integration

Install the [Grafana Intelligence Gateway panel](https://github.com/digitalrcs/grafana-intelligence-gateway) alongside this data source.

Both are separately installed plugins in the same Grafana installation; the instance and dashboard must share a Grafana organization. Follow [Installation and Quick Start](Installation-and-Quick-Start) for plugin-directory placement. Do not copy this data source into the panel's folder.

The panel has two independent connections:

| Setting                                 | Select                                                                                                                          |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **Queries**                             | Your metric/table data source, or Grafana's **Dashboard** data source and the existing panel whose results you want to analyze. |
| **AI provider → Secure AI data source** | A configured **Intelligence Gateway Secure AI** instance for model access.                                                      |

1. Configure and test an **Intelligence Gateway Secure AI** data-source instance.
2. Edit the Intelligence Gateway panel.
   Under **Queries**, connect the data to analyze and confirm it returns results for the current time range.
3. Under **AI provider**, select the instance in **Secure AI data source**.
4. Select **Load models securely**, or enter an administrator-allowed model.
5. Choose temperature and an output request cap, then select **Analyze**.

![The panel using the secure data source](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/panel-secure-datasource.png)

The panel sends the constructed prompt and non-secret generation choices through Grafana's authenticated data-source resource API. The effective hard output limit is the lower of the panel cap and data-source `maxOutputTokens`. If the panel uses the provider-default option, the administrator ceiling and provider limit still apply.

The current panel has no direct provider credential fields. Configure credentials only on the data-source instance. Installing the plugin or importing a dashboard does not create a provider connection automatically. Multiple panels can reuse one instance, but each panel must select its UID.
