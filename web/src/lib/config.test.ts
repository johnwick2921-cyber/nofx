import { it, vi, expect, afterEach } from 'vitest'
import { getSystemConfig, invalidateSystemConfig } from './config'
afterEach(() => {
  invalidateSystemConfig()
  vi.unstubAllGlobals()
})
it('retries after an HTTP error instead of caching the failure', async () => {
  invalidateSystemConfig()
  vi.stubGlobal(
    'fetch',
    vi
      .fn()
      .mockResolvedValueOnce({ ok: false })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ initialized: true }),
      })
  )
  await expect(getSystemConfig()).rejects.toThrow()
  await expect(getSystemConfig()).resolves.toEqual({ initialized: true })
})
