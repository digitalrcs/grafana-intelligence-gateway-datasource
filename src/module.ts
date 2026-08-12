import { DataSourcePlugin } from '@grafana/data';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { DataSource } from './datasource';
import { IntelligenceGatewayDataSourceOptions, IntelligenceGatewayQuery } from './types';

export const plugin = new DataSourcePlugin<DataSource, IntelligenceGatewayQuery, IntelligenceGatewayDataSourceOptions>(
  DataSource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
