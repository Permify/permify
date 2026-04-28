import { z } from 'zod';

export const ConfigSchema = z.object({
  apiUrl: z.string().url(),
  apiToken: z.string().optional(),
  tenantId: z.string().default('default'),
});

export type Config = z.infer<typeof ConfigSchema>;

export const CheckSchema = z.object({
  subject: z.string(),
  action: z.string(),
  object: z.string(),
});

export const TenantCreateSchema = z.object({
  id: z.string().regex(/[a-zA-Z0-9-,]+/).max(64),
  name: z.string().max(64),
});

export const TenantListSchema = z.object({
  pageSize: z.number().int().gte(1).optional().default(20),
  continuousToken: z.string().optional().default(''),
});
