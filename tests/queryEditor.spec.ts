import { expect, test } from '@grafana/plugin-e2e';

test('renders the secure gateway query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  const row = panelEditPage.getQueryEditorRow('A');
  await expect(row.getByRole('textbox', { name: 'User prompt' })).toBeVisible();
  await expect(row.getByRole('textbox', { name: 'Model override' })).toBeVisible();
  await expect(row.getByRole('spinbutton', { name: 'Temperature' })).toBeVisible();
  await expect(row.getByRole('spinbutton', { name: 'Output token cap' })).toBeVisible();
});

test('runs a new query when the user prompt changes', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const row = panelEditPage.getQueryEditorRow('A');
  const queryReq = panelEditPage.waitForQueryDataRequest();
  await row.getByRole('textbox', { name: 'User prompt' }).fill('Summarize the current panel data.');
  await row.getByRole('textbox', { name: 'User prompt' }).blur();
  await expect(await queryReq).toBeTruthy();
});
