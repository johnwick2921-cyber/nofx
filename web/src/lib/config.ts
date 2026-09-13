export interface SystemConfig {
  initialized: boolean
  beta_mode?: boolean
  /** true when this instance is an isolated SANDBOX (own DB, no live trading). */
  sandbox?: boolean
}

let configPromise: Promise<SystemConfig> | null = null
let cachedConfig: SystemConfig | null = null

export function getSystemConfig(): Promise<SystemConfig> {
  if (cachedConfig) {
    return Promise.resolve(cachedConfig)
  }
  if (configPromise) {
    return configPromise
  }
  const request = fetch('/api/config')
    .then((res) => {
      if (!res.ok) throw new Error('Could not read system configuration')
      return res.json()
    })
    .then((data: SystemConfig) => {
      if (configPromise === request) cachedConfig = data
      return data
    })
  configPromise = request
  void request.catch(() => {
    if (configPromise === request) configPromise = null
  })
  return request
}

/** Call after first-time setup completes so next check reflects initialized=true */
export function invalidateSystemConfig() {
  cachedConfig = null
  configPromise = null
  window.dispatchEvent(new Event('system-config-invalidated'))
}
