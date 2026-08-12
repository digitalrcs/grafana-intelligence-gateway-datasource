# Troubleshooting

## Save & test still says HTTP is restricted

Confirm **Allow insecure HTTP** is checked on the same data-source instance and select **Save & test**. If the checkbox is missing or the old error remains, rebuild both plugin halves and restart Grafana:

```bash
npm run build
mage -v build:linux
docker compose restart grafana
```

A backend executable from an older build can continue enforcing the previous locality-only rule even when the frontend displays the newer setting.

## Model is rejected

The requested model must match `allowedModels` exactly when the allow-list is not empty. Add the model administratively or use the configured default.

## Provider cannot be reached

Remember that the backend runs from the Grafana server/container. Test DNS and connectivity from that network namespace. Redirects are intentionally rejected; configure the final provider URL directly.

## Credential does not update

Saved secrets are write-only. Select the field's reset control, enter the replacement, and save. Grafana will display only configured state afterward.
