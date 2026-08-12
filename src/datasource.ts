import { CoreApp, DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  DEFAULT_QUERY,
  IntelligenceGatewayDataSourceOptions,
  IntelligenceGatewayQuery,
} from './types';

export class DataSource extends DataSourceWithBackend<IntelligenceGatewayQuery, IntelligenceGatewayDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<IntelligenceGatewayDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<IntelligenceGatewayQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: IntelligenceGatewayQuery, scopedVars: ScopedVars): IntelligenceGatewayQuery {
    return {
      ...query,
      systemPrompt: getTemplateSrv().replace(query.systemPrompt ?? '', scopedVars),
      userPrompt: getTemplateSrv().replace(query.userPrompt ?? '', scopedVars),
    };
  }

  filterQuery(query: IntelligenceGatewayQuery): boolean {
    return Boolean(query.userPrompt?.trim());
  }
}
