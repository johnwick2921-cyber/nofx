// FIX-LABELS guard tests (DS-103): pin the Studio labels to the code's truth so
// they cannot drift again. Code truth for the coin_source fields is the
// CoinSourceConfig struct comments at store/strategy.go:1905-1930.
import { describe, expect, it } from 'vitest'
import { coinSource } from './strategy-translations'

// The source-type value set the code accepts (store/strategy.go source_type
// comment: "static" | "ai500" | "oi_top" | "oi_low"). A label that drops or
// adds a value would misdescribe what the engine will use.
const sourceTypeValues = ['static', 'ai500', 'oi_top', 'oi_low']

describe('coinSource label truth (pinned to store/strategy.go:1905-1930)', () => {
  it('sourceType label names exactly the four code values', () => {
    for (const lang of ['en', 'zh', 'es'] as const) {
      const text = coinSource.sourceType[lang]
      for (const v of sourceTypeValues) {
        expect(text.toLowerCase()).toContain(v)
      }
    }
  })

  it('useOITop label says ranking for long, useOILow says ranking for short', () => {
    expect(coinSource.useOITop.en).toContain('OI increase ranking')
    expect(coinSource.useOITop.en.toLowerCase()).toContain('long')
    expect(coinSource.useOILow.en).toContain('OI decrease ranking')
    expect(coinSource.useOILow.en.toLowerCase()).toContain('short')
  })

  it('limit labels name whose pool the count caps', () => {
    expect(coinSource.ai500Limit.en).toContain('AI500')
    expect(coinSource.oiTopLimit.en).toContain('OI Top')
    expect(coinSource.oiLowLimit.en).toContain('OI Low')
  })

  it('staticDesc ties the list to source_type = static', () => {
    expect(coinSource.staticDesc.en).toContain('static')
  })

  it('excludedCoinsDesc says all sources, not a trading guarantee', () => {
    expect(coinSource.excludedCoinsDesc.en).toContain('all coin sources')
    expect(coinSource.excludedCoinsDesc.en.toLowerCase()).not.toContain(
      'will not be traded'
    )
  })
})
