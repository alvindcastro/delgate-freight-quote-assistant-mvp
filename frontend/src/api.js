const API_URL = import.meta.env.VITE_API_URL || ''
const API_TOKEN = import.meta.env.VITE_API_TOKEN || ''

async function request(path, options = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(API_TOKEN ? { Authorization: `Bearer ${API_TOKEN}` } : {}),
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
  }).then(normalizeParseResponse)
}

function normalizeParseResponse(payload) {
  const response = payload && typeof payload === 'object' ? payload : {}
  const draft = response.draft && typeof response.draft === 'object' ? response.draft : {}

  return {
    ...response,
    draft: {
      ...draft,
      origin: normalizeLocation(draft.origin),
      destination: normalizeLocation(draft.destination),
      accessorials: normalizeArray(draft.accessorials),
    },
    extractedFields: normalizeArray(response.extractedFields),
    missingHints: normalizeArray(response.missingHints),
    warnings: normalizeArray(response.warnings),
  }
}

function normalizeLocation(location) {
  return location && typeof location === 'object' ? location : {}
}

function normalizeArray(value) {
  return Array.isArray(value) ? value : []
}
