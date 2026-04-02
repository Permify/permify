import {defineConfig} from 'vite';
import react from '@vitejs/plugin-react';
import tsconfigPaths from 'vite-tsconfig-paths';
import {BLOB_UPLOAD_ROUTE} from './src/utility/helpers/blob-upload.js';
import {processBlobUpload} from './blob-upload.js';

function readJsonBody(req) {
  return new Promise((resolve, reject) => {
    let rawBody = '';

    req.on('data', (chunk) => {
      rawBody += chunk;
    });

    req.on('end', () => {
      try {
        resolve(rawBody ? JSON.parse(rawBody) : {});
      } catch (error) {
        reject(error);
      }
    });

    req.on('error', reject);
  });
}

function blobUploadDevPlugin() {
  return {
    name: 'playground-blob-upload-dev-route',
    configureServer(server) {
      server.middlewares.use(BLOB_UPLOAD_ROUTE, async (req, res, next) => {
        if (req.method !== 'POST') {
          next();
          return;
        }

        try {
          const body = await readJsonBody(req);
          const jsonResponse = await processBlobUpload({
            request: req,
            body,
            token: process.env.BLOB_READ_WRITE_TOKEN,
          });

          res.statusCode = 200;
          res.setHeader('Content-Type', 'application/json');
          res.end(JSON.stringify(jsonResponse));
        } catch (error) {
          res.statusCode = 400;
          res.setHeader('Content-Type', 'application/json');
          res.end(
            JSON.stringify({
              error: error instanceof Error ? error.message : 'Blob upload failed.',
            })
          );
        }
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), tsconfigPaths(), blobUploadDevPlugin()],
  envPrefix: ['VITE_', 'REACT_APP_'],
  esbuild: {
    loader: 'tsx',
    include: /src\/.*\.[jt]sx?$/,
    exclude: [],
  },
  optimizeDeps: {
    esbuildOptions: {
      loader: {
        '.js': 'jsx',
        '.jsx': 'jsx',
        '.ts': 'ts',
        '.tsx': 'tsx',
      },
    },
  },
  server: {
    host: '127.0.0.1',
    port: 3000,
  },
  preview: {
    host: '127.0.0.1',
    port: 4173,
  },
  build: {
    outDir: 'build',
    emptyOutDir: true,
  },
});
