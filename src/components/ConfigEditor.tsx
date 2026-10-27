import React, { ChangeEvent } from 'react';
import { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import {
  Alert,
  FieldSet,
  InlineField,
  Input,
  RadioButtonGroup,
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

const providerOptions: Array<SelectableValue<Provider>> = [
  { label: 'OpenAI', value: 'openai', description: 'OpenAI API' },
  { label: 'LM Studio', value: 'lmstudio', description: 'Local OpenAI-compatible server' },
  { label: 'Open WebUI', value: 'openwebui', description: 'Open WebUI API' },
  { label: 'Grafana LLM', value: 'grafana-llm', description: 'Configured Grafana LLM app' },
  { label: 'Copilot Studio', value: 'copilot-studio', description: 'Approved Direct Line channel' },
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
          <RadioButtonGroup
            id="config-provider"
            aria-label="Provider"
            options={providerOptions}
            value={jsonData.provider}
            onChange={(provider) => updateJson('provider', provider)}
          />
        </InlineField>
        {jsonData.provider === 'openwebui' ? (
          <Alert title="Open WebUI connection" severity="info">
            Enter the Open WebUI root URL or its /api URL, including any deployment subpath. The backend uses
            /api/models and /api/chat/completions. Store the Open WebUI API key or bearer token below.
          </Alert>
        ) : null}
        {jsonData.provider === 'grafana-llm' ? (
          <Alert title="Grafana LLM app connection" severity="info">
            Requires the configured grafana-llm-app on Grafana 13 or later. Use a Base URL ending in
            /api/plugins/grafana-llm-app/resources/llm/v1, including any Grafana subpath. Set the model to base or
            large. Store a Grafana service-account token with access to the LLM app in Bearer token. Requests run as
            that service account. The app&apos;s own two-minute provider timeout also applies.
          </Alert>
        ) : null}
        {jsonData.provider === 'copilot-studio' ? (
          <Alert title="Copilot Studio Direct Line connection" severity="info">
            Use an approved published agent&apos;s Direct Line channel secret and HTTPS base URL ending in
            /v3/directline. Default model and allowed models are local agent aliases, not selectable underlying models.
            Temperature and output-token settings are controlled by the agent and do not apply to Direct Line. This
            connection does not perform interactive end-user sign-in or enterprise SSO; agents requesting sign-in return
            an explicit unsupported authentication error. This preview adapter requires correlated text replies and a
            completion signal. Enable file uploads in the agent before approving it for image analysis.
          </Alert>
        ) : null}
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
        <InlineField
          label={jsonData.provider === 'copilot-studio' ? 'Default agent alias' : 'Default model'}
          labelWidth={24}
          required
        >
          <Input
            id="config-default-model"
            width={40}
            value={jsonData.defaultModel}
            onChange={(event: ChangeEvent<HTMLInputElement>) => updateJson('defaultModel', event.target.value)}
          />
        </InlineField>
        <InlineField
          label={jsonData.provider === 'copilot-studio' ? 'Allowed agent aliases' : 'Allowed models'}
          labelWidth={24}
          tooltip="One model ID per line. Include the default model. When empty, only the default model is permitted."
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
        <InlineField
          label="Maximum output tokens"
          labelWidth={24}
          required
          disabled={jsonData.provider === 'copilot-studio'}
        >
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

      <FieldSet label="Image analysis policy">
        <InlineField
          label="Enable image analysis"
          labelWidth={24}
          tooltip="Permit explicitly attached PNG/JPEG images to be sent to approved vision models. Disabled by default."
        >
          <Switch
            id="config-enable-vision"
            value={jsonData.enableVision}
            onChange={(event) => updateJson('enableVision', event.currentTarget.checked)}
          />
        </InlineField>
        {jsonData.enableVision ? (
          <>
            <InlineField
              label="Allowed vision models"
              labelWidth={24}
              required
              tooltip="One exact image-capable model ID per line. Every entry must also be allowed by the model policy. Capabilities are never inferred from model names."
            >
              <TextArea
                id="config-allowed-vision-models"
                rows={4}
                cols={60}
                value={(jsonData.allowedVisionModels ?? []).join('\n')}
                onChange={(event: ChangeEvent<HTMLTextAreaElement>) =>
                  updateJson(
                    'allowedVisionModels',
                    event.target.value
                      .split('\n')
                      .map((model) => model.trim())
                      .filter(Boolean)
                  )
                }
              />
            </InlineField>
            <Alert title="Image limits and data handling" severity="info">
              Captured panels can contain sensitive visible information. Users preview and explicitly send images to the
              selected provider. The backend permits up to 8 PNG/JPEG images, 4 MiB each and 12 MiB total, at most 4096
              pixels per side, 16 million pixels per image and 32 million pixels total. Encoded requests are limited to
              18 MiB; text remains limited to 1 MiB. Provider limits may be lower. Verify each allowed model supports
              images; for Grafana LLM, verify the model behind the base or large alias.
            </Alert>
          </>
        ) : null}
      </FieldSet>

      <FieldSet label="Credentials">
        {jsonData.provider === 'copilot-studio' ? (
          <SecretField
            id="config-direct-line-secret"
            label="Direct Line secret"
            configured={Boolean(options.secureJsonFields.directLineSecret)}
            value={options.secureJsonData?.directLineSecret}
            onChange={(value) => updateSecret('directLineSecret', value)}
            onReset={() => resetSecret('directLineSecret')}
          />
        ) : null}
        {jsonData.provider !== 'copilot-studio' ? (
          <>
            {jsonData.provider !== 'grafana-llm' ? (
              <SecretField
                id="config-api-key"
                label="API key"
                configured={Boolean(options.secureJsonFields.apiKey)}
                value={options.secureJsonData?.apiKey}
                onChange={(value) => updateSecret('apiKey', value)}
                onReset={() => resetSecret('apiKey')}
              />
            ) : null}
            <SecretField
              id="config-bearer-token"
              label="Bearer token"
              configured={Boolean(options.secureJsonFields.bearerToken)}
              value={options.secureJsonData?.bearerToken}
              onChange={(value) => updateSecret('bearerToken', value)}
              onReset={() => resetSecret('bearerToken')}
            />
          </>
        ) : null}
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
