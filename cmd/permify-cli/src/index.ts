import { Command } from 'commander';
import { ConfigSchema } from './schema.js';
import { CheckStrategy, SchemaWriteStrategy } from './router.js';
import * as fs from 'fs';
import * as path from 'path';
import { os } from 'os';

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

program.parse();
