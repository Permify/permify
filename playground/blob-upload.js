import {handleUpload} from '@vercel/blob/client';
import {SHARE_PATH_PREFIX} from './src/utility/helpers/blob-upload.js';

export const MAX_SHARE_FILE_SIZE_BYTES = 1024 * 1024;
export const SHARE_ALLOWED_CONTENT_TYPES = [
  'text/yaml',
  'text/x-yaml',
  'application/yaml',
  'application/x-yaml',
  'text/plain',
  'application/octet-stream',
];

function isValidSharePathname(pathname) {
  return new RegExp(`^${SHARE_PATH_PREFIX}[A-Za-z0-9-]+\\.ya?ml$`).test(pathname);
}

function getHeader(request, name) {
  if (typeof request.headers?.get === 'function') {
    return request.headers.get(name);
  }

  const value = request.headers?.[name];
  return Array.isArray(value) ? value[0] : value ?? null;
}

function isAllowedOrigin(request) {
  const origin = getHeader(request, 'origin');

  if (!origin) {
    return true;
  }

  const host = getHeader(request, 'x-forwarded-host') || getHeader(request, 'host');

  if (!host) {
    return false;
  }

  return origin === `https://${host}` || origin === `http://${host}`;
}

export async function processBlobUpload({request, body, token}) {
  if (!isAllowedOrigin(request)) {
    throw new Error('Unauthorized upload origin.');
  }

  return handleUpload({
    token,
    request,
    body,
    onBeforeGenerateToken: async (pathname) => {
      if (!isValidSharePathname(pathname)) {
        throw new Error('Invalid upload pathname.');
      }

      return {
        allowedContentTypes: SHARE_ALLOWED_CONTENT_TYPES,
        maximumSizeInBytes: MAX_SHARE_FILE_SIZE_BYTES,
        validUntil: Date.now() + 60 * 1000,
        addRandomSuffix: false,
        allowOverwrite: false,
      };
    },
    onUploadCompleted: async () => {},
  });
}
