// Keep event state across arbitrary UTF-8 transport chunks.
export async function* readSSE(
  reader: ReadableStreamDefaultReader<Uint8Array>
) {
  const decoder = new TextDecoder()
  let buffer = '',
    type = '',
    data: string[] = []
  while (true) {
    const part = await reader.read()
    buffer += part.done
      ? decoder.decode()
      : decoder.decode(part.value, { stream: true })
    if (part.done && buffer && !buffer.endsWith('\n')) buffer += '\n'
    let end: number
    while ((end = buffer.indexOf('\n')) >= 0) {
      const line = buffer.slice(0, end).replace(/\r$/, '')
      buffer = buffer.slice(end + 1)
      if (!line) {
        if (data.length)
          yield { type: type || 'message', data: data.join('\n') }
        type = ''
        data = []
      } else if (!line.startsWith(':')) {
        const colon = line.indexOf(':')
        const field = colon < 0 ? line : line.slice(0, colon)
        const value = colon < 0 ? '' : line.slice(colon + 1).replace(/^ /, '')
        if (field === 'event') type = value
        if (field === 'data') data.push(value)
      }
    }
    if (part.done) {
      if (data.length) yield { type: type || 'message', data: data.join('\n') }
      break
    }
  }
}
