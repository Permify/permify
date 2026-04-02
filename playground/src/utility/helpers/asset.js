import {getBaseUrl} from "./env";

export const toAbsoluteUrl = pathname => {
  const normalizedBase = getBaseUrl().replace(/\/$/, '');
  const normalizedPath = pathname.startsWith('/') ? pathname : `/${pathname}`;

  return `${normalizedBase}${normalizedPath}`;
};
