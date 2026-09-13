import { expect, it } from 'vitest'
import { readSSE } from './sse'
it('parses every byte split, CRLF, Unicode, multiline data and final unterminated event', async () => {
  const bytes = new TextEncoder().encode(
    'event: delta\r\ndata: "你"\r\n\r\nevent: plan\ndata: one\ndata: two'
  )
  for (let split = 0; split <= bytes.length; split++) {
    const reader = new ReadableStream<Uint8Array>({
      start(c) {
        c.enqueue(bytes.slice(0, split))
        c.enqueue(bytes.slice(split))
        c.close()
      },
    }).getReader()
    const got = []
    for await (const event of readSSE(reader)) got.push(event)
    expect(got).toEqual([
      { type: 'delta', data: '"你"' },
      { type: 'plan', data: 'one\ntwo' },
    ])
  }
})
