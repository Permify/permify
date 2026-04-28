import { Command } from 'commander';
import { ConfigSchema, TenantCreateSchema, TenantListSchema } from './schema.js';
import { CheckStrategy, SchemaWriteStrategy, TenantCreateStrategy, TenantDeleteStrategy, TenantListStrategy } from './router.js';
import * as fs from 'fs';
import * as path from 'path';
const program = new Command();
program
    .name('permify')
    .description('Permify CLI for managing authorization')
    .version('1.0.0');
const getConfig = () => {
    const configPath = path.join(process.env.HOME || process.env.USERPROFILE || '', '.permify', 'config.json');
    if (fs.existsSync(configPath)) {
        return JSON.parse(fs.readFileSync(configPath, 'utf8'));
    }
    return { apiUrl: 'http://localhost:3476', tenantId: 'default' };
};
program
    .command('check')
    .description('Check permissions')
    .requiredOption('-sub, --subject <type:id>', 'Subject (e.g. user:1)')
    .requiredOption('-act, --action <action>', 'Action (e.g. view)')
    .requiredOption('-obj, --object <type:id>', 'Object (e.g. document:1)')
    .action(async (options) => {
    const config = ConfigSchema.parse(getConfig());
    const strategy = new CheckStrategy();
    await strategy.execute(config, options);
});
program
    .command('schema-write')
    .description('Write authorization schema')
    .requiredOption('-f, --file <path>', 'Schema file path (.perm)')
    .action(async (options) => {
    const config = ConfigSchema.parse(getConfig());
    const schema = fs.readFileSync(options.file, 'utf8');
    const strategy = new SchemaWriteStrategy();
    await strategy.execute(config, { schema });
});
const tenant = program.command('tenant').description('Tenant management');
tenant
    .command('list')
    .description('List all tenants')
    .option('-s, --size <number>', 'Page size', '20')
    .option('-t, --token <string>', 'Continuous token', '')
    .action(async (options) => {
    const config = ConfigSchema.parse(getConfig());
    const parsed = TenantListSchema.parse({
        pageSize: parseInt(options.size),
        continuousToken: options.token
    });
    const strategy = new TenantListStrategy();
    await strategy.execute(config, parsed);
});
tenant
    .command('create')
    .description('Create a new tenant')
    .requiredOption('-i, --id <string>', 'Tenant ID')
    .requiredOption('-n, --name <string>', 'Tenant name')
    .action(async (options) => {
    const config = ConfigSchema.parse(getConfig());
    const parsed = TenantCreateSchema.parse(options);
    const strategy = new TenantCreateStrategy();
    await strategy.execute(config, parsed);
});
tenant
    .command('delete')
    .description('Delete a tenant')
    .requiredOption('-i, --id <string>', 'Tenant ID')
    .action(async (options) => {
    const config = ConfigSchema.parse(getConfig());
    const strategy = new TenantDeleteStrategy();
    await strategy.execute(config, options);
});
program.parse();
