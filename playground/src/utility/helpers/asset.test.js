import {describe, expect, it, vi} from 'vitest';

describe('toAbsoluteUrl', () => {
  it('prefixes assets with the configured Vite base url', async () => {
    vi.stubEnv('BASE_URL', '/playground/');

    const {toAbsoluteUrl} = await import('./asset');

    expect(toAbsoluteUrl('/media/svg/rocket.svg')).toBe('/playground/media/svg/rocket.svg');
  });
});
