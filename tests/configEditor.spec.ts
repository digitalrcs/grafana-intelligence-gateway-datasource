import { expect, test } from '@grafana/plugin-e2e';

test('renders the provider policy and secure credential editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });

  await expect(page.getByLabel('Base URL')).toBeVisible();
  await expect(page.getByLabel('Default model')).toBeVisible();
  await expect(page.getByLabel('API key')).toBeVisible();
  await expect(page.getByText('Credentials are encrypted by Grafana')).toBeVisible();
});

test('rejects OpenAI configuration without a server-side credential', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByLabel('Base URL').fill('https://api.openai.com/v1');

  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'API key or bearer token is required' });
});
