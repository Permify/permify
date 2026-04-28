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
