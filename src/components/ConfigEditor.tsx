import React, { ChangeEvent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import {
  Alert,
  Combobox,
  type ComboboxOption,
  FieldSet,
  InlineField,
  Input,
  SecretInput,
  Stack,
  Switch,
  TextArea,
} from '@grafana/ui';
import {
  DEFAULT_CONFIG,
  IntelligenceGatewayDataSourceOptions,
  IntelligenceGatewaySecureJsonData,
  Provider,
} from '../types';

type Props = DataSourcePluginOptionsEditorProps<
  IntelligenceGatewayDataSourceOptions,
  IntelligenceGatewaySecureJsonData
>;

const providerOptions: Array<ComboboxOption<Provider>> = [
  { label: 'OpenAI', value: 'openai', description: 'OpenAI API' },
  { label: 'LM Studio', value: 'lmstudio', description: 'Local OpenAI-compatible server' },
  { label: 'Custom', value: 'custom', description: 'Remote OpenAI-compatible API' },
];

export function ConfigEditor({ options, onOptionsChange }: Props) {
  const jsonData = { ...DEFAULT_CONFIG, ...options.jsonData };

  const updateJson = <K extends keyof IntelligenceGatewayDataSourceOptions>(
    key: K,
    value: IntelligenceGatewayDataSourceOptions[K]
  ) => onOptionsChange({ ...options, jsonData: { ...options.jsonData, [key]: value } });

  const updateSecret = (key: keyof IntelligenceGatewaySecureJsonData, value: string) =>
    onOptionsChange({
      ...options,
      secureJsonData: { ...options.secureJsonData, [key]: value },
    });

  const resetSecret = (key: keyof IntelligenceGatewaySecureJsonData) =>
    onOptionsChange({
      ...options,
      secureJsonFields: { ...options.secureJsonFields, [key]: false },
      secureJsonData: { ...options.secureJsonData, [key]: '' },
    });

  return (
    <Stack direction="column" gap={2}>
      <Alert title="Server-side secret boundary" severity="info">
        Credentials are encrypted by Grafana and used only by this backend. They are never returned to dashboards or
        browser code. Restrict who can edit and query this data source.
      </Alert>

      <FieldSet label="Provider policy">
        <InlineField label="Provider" labelWidth={24} required>
          <Combobox
            id="config-provider"
            width={40}
            options={providerOptions}
            value={jsonData.provider}
            onChange={(choice) => updateJson('provider', choice.value ?? 'openai')}
          />
        </InlineField>
        <InlineField
          label="Base URL"
          labelWidth={24}
          required
          tooltip="Administrator-controlled URL. HTTPS is required unless Allow insecure HTTP is explicitly enabled."
        >
          <Input
            id="config-base-url"
            width={60}
            value={jsonData.baseUrl}
            placeholder="https://api.openai.com/v1"
            onChange={(event: ChangeEvent<HTMLInputElement>) => updateJson('baseUrl', event.target.value)}
          />
        </InlineField>
        <InlineField
          label="Allow insecure HTTP"
          labelWidth={24}
          tooltip="Disables the HTTPS requirement for this data source. Credentials and prompts can be intercepted in transit."
        >
          <Switch
            id="config-allow-insecure-http"
            value={jsonData.allowInsecureHttp}
            onChange={(event) => updateJson('allowInsecureHttp', event.currentTarget.checked)}
          />
        </InlineField>
        {jsonData.allowInsecureHttp ? (
          <Alert title="Insecure HTTP override enabled" severity="warning">
            Provider credentials and prompts may cross the network without transport encryption. Enable this only on a
            network you trust.
          </Alert>
        ) : null}
        <InlineField label="Default model" labelWidth={24} required>
          <Input
            id="config-default-model"
            width={40}
            value={jsonData.defaultModel}
            onChange={(event: ChangeEvent<HTMLInputElement>) => updateJson('defaultModel', event.target.value)}
          />
        </InlineField>
        <InlineField
          label="Allowed models"
          labelWidth={24}
          tooltip="One model ID per line. The default model must be included. When empty, only the default model is permitted."
        >
          <TextArea
            id="config-allowed-models"
            rows={4}
            cols={60}
            value={(jsonData.allowedModels ?? []).join('\n')}
            onChange={(event: ChangeEvent<HTMLTextAreaElement>) =>
              updateJson(
                'allowedModels',
                event.target.value
                  .split('\n')
                  .map((model) => model.trim())
                  .filter(Boolean)
              )
            }
          />
        </InlineField>
        <InlineField label="Timeout (seconds)" labelWidth={24} required>
          <Input
            id="config-timeout-seconds"
            type="number"
            min={1}
            max={600}
            width={20}
            value={jsonData.timeoutSeconds}
            onChange={(event: ChangeEvent<HTMLInputElement>) =>
              updateJson('timeoutSeconds', Number(event.target.value))
            }
          />
        </InlineField>
        <InlineField label="Maximum output tokens" labelWidth={24} required>
          <Input
            id="config-max-output-tokens"
            type="number"
            min={1}
            max={1048576}
            width={24}
            value={jsonData.maxOutputTokens}
            onChange={(event: ChangeEvent<HTMLInputElement>) =>
              updateJson('maxOutputTokens', Number(event.target.value))
            }
          />
        </InlineField>
      </FieldSet>

      <FieldSet label="Credentials">
        <SecretField
          id="config-api-key"
          label="API key"
          configured={Boolean(options.secureJsonFields.apiKey)}
          value={options.secureJsonData?.apiKey}
          onChange={(value) => updateSecret('apiKey', value)}
          onReset={() => resetSecret('apiKey')}
        />
        <SecretField
          id="config-bearer-token"
          label="Bearer token"
          configured={Boolean(options.secureJsonFields.bearerToken)}
          value={options.secureJsonData?.bearerToken}
          onChange={(value) => updateSecret('bearerToken', value)}
          onReset={() => resetSecret('bearerToken')}
        />
      </FieldSet>
    </Stack>
  );
}

interface SecretFieldProps {
  id: string;
  label: string;
  configured: boolean;
  value?: string;
  onChange: (value: string) => void;
  onReset: () => void;
}

function SecretField({ id, label, configured, value, onChange, onReset }: SecretFieldProps) {
  return (
    <InlineField label={label} labelWidth={24}>
      <SecretInput
        id={id}
        width={60}
        isConfigured={configured}
        value={value ?? ''}
        placeholder={`Enter ${label.toLowerCase()}`}
        onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.target.value)}
        onReset={onReset}
      />
    </InlineField>
  );
}
