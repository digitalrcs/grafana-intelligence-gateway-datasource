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
  await expect(page.getByText('DEMO ONLY - Mock AI (no inference)', { exact: true })).toHaveCount(0);
  await expect(page.getByLabel('OAuth client secret')).toHaveCount(0);
  await expect(page.getByLabel('Permit streaming')).toHaveCount(0);
  await page.getByLabel('Allow insecure HTTP').check({ force: true });
  await expect(page.getByText('Insecure HTTP override enabled')).toBeVisible();
  await page.getByLabel('Allow insecure HTTP').uncheck({ force: true });
  await expect(page.getByText('Insecure HTTP override enabled')).toHaveCount(0);
  await expect(page.getByText('Credentials are encrypted by Grafana')).toBeVisible();
});

test('labels only the supplied mock and clears its notice when the connection changes', async ({
  gotoDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<IntelligenceGatewayDataSourceOptions>({ fileName: 'datasources.yml' });
  const configPage = await gotoDataSourceConfigPage(ds.uid);
  const receiptNotice = page.getByText(/This demo returns a fixed delivery receipt/);

  await expect(configPage).toHaveAlert('warning', { hasText: 'DEMO ONLY - Mock AI (no inference)' });
  await expect(receiptNotice).toBeVisible();
  await expect(page.getByText('Create a separate Intelligence Gateway Secure AI data source.')).toBeVisible();
  await expect(page.getByText(/not its Queries selector/)).toBeVisible();

  // Do not save: leave the supplied mock unchanged for other tests and reviewers.
  await page.getByRole('radio', { name: 'LM Studio', exact: true }).check({ force: true });
  await expect(receiptNotice).toHaveCount(0);
  await page.getByRole('radio', { name: 'Custom', exact: true }).check({ force: true });
  await expect(receiptNotice).toBeVisible();
  // Grafana includes the required-field marker in this input's accessible name.
  await page.getByRole('textbox', { name: /^Base URL(?: \*)?$/ }).fill('https://ai.example.com/v1');
  await expect(receiptNotice).toHaveCount(0);
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
  await page.getByRole('radio', { name: 'Custom', exact: true }).check({ force: true });
  await page.getByLabel('Base URL').fill(ds.jsonData.baseUrl!);
  await page.getByLabel('Default model').fill(ds.jsonData.defaultModel!);
  await page.getByLabel('Allowed models').fill('');

  await expect(configPage.saveAndTest()).not.toBeOK();
  await page.getByLabel('Allow insecure HTTP').check({ force: true });
  await expect(page.getByText('Insecure HTTP override enabled')).toBeVisible();
  await expect(configPage.saveAndTest()).toBeOK();
});
