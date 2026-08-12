import { DataQuery, DataSourceJsonData } from '@grafana/schema';

export type Provider = 'openai' | 'lmstudio' | 'custom';

export interface IntelligenceGatewayQuery extends DataQuery {
  systemPrompt?: string;
  userPrompt?: string;
  model?: string;
  temperature?: number;
  maxOutputTokens?: number;
}

export const DEFAULT_QUERY: Partial<IntelligenceGatewayQuery> = {
  temperature: 0.2,
};

export interface IntelligenceGatewayDataSourceOptions extends DataSourceJsonData {
  provider?: Provider;
  baseUrl?: string;
  defaultModel?: string;
  timeoutSeconds?: number;
  allowedModels?: string[];
  maxOutputTokens?: number;
  allowStreaming?: boolean;
  allowInsecureHttp?: boolean;
}

/** Values are written on save and never returned to browser code. */
export interface IntelligenceGatewaySecureJsonData {
  apiKey?: string;
  bearerToken?: string;
  clientSecret?: string;
}

export const DEFAULT_CONFIG: Required<Omit<IntelligenceGatewayDataSourceOptions, keyof DataSourceJsonData>> = {
  provider: 'openai',
  baseUrl: 'https://api.openai.com/v1',
  defaultModel: 'gpt-4.1-mini',
  timeoutSeconds: 300,
  allowedModels: ['gpt-4.1-mini'],
  maxOutputTokens: 256000,
  allowStreaming: false,
  allowInsecureHttp: false,
};
