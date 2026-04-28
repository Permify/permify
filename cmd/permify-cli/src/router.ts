import axios from 'axios';
import { Config } from './schema.js';

export interface CommandStrategy {
  execute(config: Config, args: any): Promise<void>;
}

function parseTypeId(value: string): { type: string; id: string } {
  const [type, id] = value.split(':', 2);
  if (!type || !id) {
    throw new Error(`Expected type:id, got ${value}`);
  }
  return { type, id };
}

export class CheckStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken, tenantId } = config;
    try {
      const response = await axios.post(`${apiUrl}/v1/tenants/${tenantId}/permissions/check`, {
        metadata: {
            schema_version: ""
        },
        entity: parseTypeId(args.object),
        permission: args.action,
        subject: parseTypeId(args.subject)
      }, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {},
        timeout: 10000
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
      process.exitCode = 1;
    }
  }
}

export class SchemaWriteStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken, tenantId } = config;
    try {
      const response = await axios.post(`${apiUrl}/v1/tenants/${tenantId}/schemas/write`, {
        schema: args.schema
      }, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {},
        timeout: 10000
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
      process.exitCode = 1;
    }
  }
}

export class TenantCreateStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken } = config;
    try {
      const response = await axios.post(`${apiUrl}/v1/tenants/create`, {
        id: args.id,
        name: args.name
      }, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {},
        timeout: 10000
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
      process.exitCode = 1;
    }
  }
}

export class TenantDeleteStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken } = config;
    try {
      const response = await axios.delete(`${apiUrl}/v1/tenants/${args.id}`, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {},
        timeout: 10000
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
      process.exitCode = 1;
    }
  }
}

export class TenantListStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken } = config;
    try {
      const response = await axios.post(`${apiUrl}/v1/tenants/list`, {
        page_size: args.pageSize,
        continuous_token: args.continuousToken
      }, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {},
        timeout: 10000
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
      process.exitCode = 1;
    }
  }
}
