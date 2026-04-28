import axios from 'axios';
import { Config } from './schema.js';

export interface CommandStrategy {
  execute(config: Config, args: any): Promise<void>;
}

export class CheckStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken, tenantId } = config;
    try {
      const response = await axios.post(`${apiUrl}/v1/tenants/${tenantId}/permissions/check`, {
        metadata: {
            schema_version: ""
        },
        entity: {
            type: args.object.split(':')[0],
            id: args.object.split(':')[1],
        },
        permission: args.action,
        subject: {
            type: args.subject.split(':')[0],
            id: args.subject.split(':')[1],
        }
      }, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
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
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
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
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
    }
  }
}

export class TenantDeleteStrategy implements CommandStrategy {
  async execute(config: Config, args: any) {
    const { apiUrl, apiToken } = config;
    try {
      const response = await axios.delete(`${apiUrl}/v1/tenants/${args.id}`, {
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
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
        headers: apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}
      });
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error(`Error: ${error.response?.data?.message || error.message}`);
    }
  }
}
