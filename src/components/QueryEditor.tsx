import React, { ChangeEvent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { InlineField, Input, Stack, TextArea } from '@grafana/ui';
import { DataSource } from '../datasource';
import { IntelligenceGatewayDataSourceOptions, IntelligenceGatewayQuery } from '../types';

type Props = QueryEditorProps<DataSource, IntelligenceGatewayQuery, IntelligenceGatewayDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const update = <K extends keyof IntelligenceGatewayQuery>(key: K, value: IntelligenceGatewayQuery[K]) =>
    onChange({ ...query, [key]: value });

  return (
    <Stack direction="column" gap={1}>
      <InlineField label="System prompt" labelWidth={18}>
        <TextArea
          id="query-system-prompt"
          rows={3}
          cols={70}
          value={query.systemPrompt ?? ''}
          onChange={(event: ChangeEvent<HTMLTextAreaElement>) => update('systemPrompt', event.target.value)}
          onBlur={onRunQuery}
        />
      </InlineField>
      <InlineField label="User prompt" labelWidth={18} required>
        <TextArea
          id="query-user-prompt"
          rows={6}
          cols={70}
          value={query.userPrompt ?? ''}
          onChange={(event: ChangeEvent<HTMLTextAreaElement>) => update('userPrompt', event.target.value)}
          onBlur={onRunQuery}
        />
      </InlineField>
      <InlineField label="Model override" labelWidth={18}>
        <Input
          id="query-model"
          width={36}
          value={query.model ?? ''}
          placeholder="Use administrator default"
          onChange={(event: ChangeEvent<HTMLInputElement>) => update('model', event.target.value)}
          onBlur={onRunQuery}
        />
      </InlineField>
      <InlineField label="Temperature" labelWidth={18}>
        <Input
          id="query-temperature"
          type="number"
          min={0}
          max={2}
          step={0.1}
          width={12}
          value={query.temperature ?? 0.2}
          onChange={(event: ChangeEvent<HTMLInputElement>) => update('temperature', Number(event.target.value))}
          onBlur={onRunQuery}
        />
      </InlineField>
      <InlineField label="Output token cap" labelWidth={18}>
        <Input
          id="query-max-output-tokens"
          type="number"
          min={1}
          width={18}
          value={query.maxOutputTokens ?? ''}
          placeholder="Admin ceiling"
          onChange={(event: ChangeEvent<HTMLInputElement>) =>
            update('maxOutputTokens', event.target.value ? Number(event.target.value) : undefined)
          }
          onBlur={onRunQuery}
        />
      </InlineField>
    </Stack>
  );
}
