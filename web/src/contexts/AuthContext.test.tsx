import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { vi, it, expect, beforeEach, afterEach } from 'vitest'
import useSWR from 'swr'
import { AuthProvider, useAuth } from './AuthContext'
vi.mock('./LanguageContext', () => ({
  useLanguage: () => ({ language: 'en' }),
}))
vi.mock('../lib/config', () => ({
  getSystemConfig: async () => ({}),
  invalidateSystemConfig: vi.fn(),
}))
const jwt = `e30.${btoa(JSON.stringify({ exp: 4102444800 }))}.test`
const privateRead = vi.fn()
function Probe() {
  const a = useAuth()
  const { data } = useSWR(a.user ? 'same-private-key' : null, () =>
    privateRead(a.user?.id)
  )
  return (
    <>
      <span data-testid="auth">
        {a.isLoading ? 'loading' : a.user?.id || 'guest'}
      </span>
      <span data-testid="data">{data || 'empty'}</span>
      <button onClick={() => a.login('bob', 'pw')}>login</button>
      <button onClick={a.logout}>logout</button>
    </>
  )
}
beforeEach(() => {
  localStorage.clear()
  privateRead.mockReset()
})
afterEach(() => vi.unstubAllGlobals())
it('completes hydration when stored user JSON is malformed', async () => {
  localStorage.setItem('auth_token', jwt)
  localStorage.setItem('auth_user', '{bad')
  render(
    <MemoryRouter>
      <AuthProvider>
        <Probe />
      </AuthProvider>
    </MemoryRouter>
  )
  await waitFor(() =>
    expect(screen.getByTestId('auth')).toHaveTextContent('guest')
  )
  expect(localStorage.getItem('auth_token')).toBeNull()
})
it('gives the next authenticated session a new SWR cache', async () => {
  localStorage.setItem('auth_token', jwt)
  localStorage.setItem('auth_user', JSON.stringify({ id: 'alice', email: 'a' }))
  privateRead.mockImplementation((id: string) =>
    id === 'alice' ? Promise.resolve('alice secret') : new Promise(() => {})
  )
  vi.stubGlobal(
    'fetch',
    vi
      .fn()
      .mockResolvedValue({
        ok: true,
        json: async () => ({ token: jwt, user_id: 'bob', email: 'b' }),
      })
  )
  render(
    <MemoryRouter>
      <AuthProvider>
        <Probe />
      </AuthProvider>
    </MemoryRouter>
  )
  await waitFor(() =>
    expect(screen.getByTestId('data')).toHaveTextContent('alice secret')
  )
  fireEvent.click(screen.getByText('logout'))
  fireEvent.click(screen.getByText('login'))
  await waitFor(() =>
    expect(screen.getByTestId('auth')).toHaveTextContent('bob')
  )
  expect(screen.getByTestId('data')).toHaveTextContent('empty')
})
