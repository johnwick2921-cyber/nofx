import { fireEvent, render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import { RiskControlEditor } from './RiskControlEditor'
import type { RiskControlConfig } from '../../types'

it('keeps the owner daily-loss control without requiring a per-trade cap', () => {
  const onChange = vi.fn()
  const config = {
    guardrails_enabled: true,
    daily_loss_enabled: true,
    daily_loss_limit_usd: 450,
  } as RiskControlConfig
  render(
    <RiskControlEditor
      config={config}
      onChange={onChange}
      language="en"
      isFutures
    />
  )
  expect(
    screen.queryByRole('spinbutton', { name: 'MNQ per-trade risk cap' })
  ).not.toBeInTheDocument()
  fireEvent.change(screen.getByDisplayValue('450'), {
    target: { value: '400' },
  })
  expect(onChange).toHaveBeenCalledWith({
    ...config,
    daily_loss_limit_usd: 400,
  })
})

it('shows zero as the inherited breaker threshold and never as Off', () => {
  const onChange = vi.fn()
  const config = {
    guardrails_enabled: true,
    consecutive_loss_halt: 0,
    daily_loss_enabled: false,
  } as RiskControlConfig
  render(
    <RiskControlEditor
      config={config}
      onChange={onChange}
      language="en"
      isFutures
    />
  )
  const control = screen.getByTestId('consecutive-loss-control')
  expect(control).toHaveTextContent('default 8 unless overridden')
  expect(within(control).queryByRole('button')).toBeNull()
  fireEvent.change(within(control).getByRole('spinbutton'), {
    target: { value: '3' },
  })
  expect(onChange).toHaveBeenCalledWith({ ...config, consecutive_loss_halt: 3 })
})
