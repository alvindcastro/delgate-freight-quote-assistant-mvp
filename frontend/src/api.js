const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

async function request(path, options = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  })

  const contentType = response.headers.get('content-type') || ''
  const payload = contentType.includes('application/json') ? await response.json() : await response.text()

  if (!response.ok) {
    const message = typeof payload === 'object' && payload.error ? payload.error : 'Request failed'
    throw new Error(message)
  }

  return payload
}

export function createQuote(data) {
  return request('/api/quote', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function getQuotes() {
  return request('/api/quotes')
}

export function parseRequestText(text) {
  return request('/api/parse-request', {
    method: 'POST',
    body: JSON.stringify({ text }),
  })
}
