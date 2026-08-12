# Panel Integration

Install the [Grafana Intelligence Gateway panel](https://github.com/digitalrcs/grafana-intelligence-gateway) alongside this data source.

1. Configure and test an **Intelligence Gateway Secure AI** data-source instance.
2. Edit the Intelligence Gateway panel.
3. Under **AI provider**, select the instance in **Secure AI data source**.
4. Select **Load models securely**, or enter an administrator-allowed model.
5. Choose temperature and an output request cap, then select **Analyze**.

![The panel using the secure data source](https://raw.githubusercontent.com/digitalrcs/grafana-intelligence-gateway-datasource/main/src/img/panel-secure-datasource.png)

The panel sends the constructed prompt and non-secret generation choices through Grafana's authenticated data-source resource API. The effective hard output limit is the lower of the panel cap and data-source `maxOutputTokens`. If the panel uses the provider-default option, the administrator ceiling and provider limit still apply.

Direct provider fields in the panel are intended for migration or restricted development. Values entered there are dashboard JSON and are not secret storage.
