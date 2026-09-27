import { expect, test } from '@grafana/plugin-e2e';
import { IntelligenceGatewayDataSourceOptions } from '../src/types';

test('renders the provider policy and secure credential editor', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });

  await expect(page.getByLabel('Base URL')).toBeVisible();
  await expect(page.getByLabel('Default model')).toBeVisible();
  await expect(page.getByLabel('Allow insecure HTTP')).toBeVisible();
  await expect(page.getByLabel('API key')).toBeVisible();
  await expect(page.getByLabel('OAuth client secret')).toHaveCount(0);
  await expect(page.getByLabel('Permit streaming')).toHaveCount(0);
  await page.getByLabel('Allow insecure HTTP').check({ force: true });
  await expect(page.getByText('Insecure HTTP override enabled')).toBeVisible();
  await page.getByLabel('Allow insecure HTTP').uncheck({ force: true });
  await expect(page.getByText('Insecure HTTP override enabled')).toHaveCount(0);
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
});

test('requires the explicit HTTP override for a provider on the Docker network', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<IntelligenceGatewayDataSourceOptions>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type, deleteDataSourceAfterTest: false });
  await page.getByRole('combobox', { name: 'Provider', exact: true }).click();
  await page.getByRole('option', { name: /^Custom/ }).click();
  await page.getByLabel('Base URL').fill(ds.jsonData.baseUrl!);
  await page.getByLabel('Default model').fill(ds.jsonData.defaultModel!);
  await page.getByLabel('Allowed models').fill('');

  await expect(configPage.saveAndTest()).not.toBeOK();
  await page.getByLabel('Allow insecure HTTP').check({ force: true });
  await expect(page.getByText('Insecure HTTP override enabled')).toBeVisible();
  await expect(configPage.saveAndTest()).toBeOK();
});
