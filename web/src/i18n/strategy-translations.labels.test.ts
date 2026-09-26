// FIX-LABELS guard tests (DS-103): pin the Studio labels to the code's truth so
// they cannot drift again. Code truth for the coin_source fields is the
// CoinSourceConfig struct comments at store/strategy.go:1905-1930.
import { describe, expect, it } from 'vitest'
import { coinSource, indicator } from './strategy-translations'

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
// Indicator block descs pin the period defaults the code ships
// (store/strategy.go:1950-1958 + DefaultConfig at store/strategy.go:2205).
describe('indicator label truth (pinned to store/strategy.go:1936-1958)', () => {
  it('ema desc names the code default periods 20, 50 (never 50/200)', () => {
    expect(indicator.emaDesc.en).toContain('20, 50')
    expect(indicator.emaDesc.en).not.toContain('200')
  })

  it('rsi desc names the code default periods 7, 14', () => {
    expect(indicator.rsiDesc.en).toContain('7, 14')
  })

  it('atr desc names the code default period 14', () => {
    expect(indicator.atrDesc.en).toContain('14')
  })

  it('boll desc names period 20 and the fixed std-dev multiplier 2', () => {
    expect(indicator.bollDesc.en).toContain('20')
    expect(indicator.bollDesc.en).toContain('fixed at 2')
  })

  it('svp desc states default OFF (code comment: default OFF)', () => {
    expect(indicator.svpDesc.en.toLowerCase()).toContain('default off')
  })

  it('rawKlines desc states always enabled (force-set on mount)', () => {
    expect(indicator.rawKlinesDesc.en.toLowerCase()).toContain('always enabled')
  })
})
