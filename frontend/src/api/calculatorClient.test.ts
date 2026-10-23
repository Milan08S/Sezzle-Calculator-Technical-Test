import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, calculate } from './calculatorClient'

function mockFetch(impl: () => Promise<Response>) {
  const fn = vi.fn(impl)
  vi.stubGlobal('fetch', fn)
  return fn
}

function jsonResponse(body: unknown, status: number) {
  return Promise.resolve(new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }))
}

async function caught(promise: Promise<unknown>): Promise<ApiError> {
  try {
    await promise
  } catch (err) {
    expect(err).toBeInstanceOf(ApiError)
    return err as ApiError
  }
  throw new Error('expected promise to reject')
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('calculate', () => {
  it('posts the request as JSON and returns the response', async () => {
    const fetchMock = mockFetch(() => jsonResponse({ operation: 'divide', operands: [10, 2], result: 5 }, 200))

    const res = await calculate({ operation: 'divide', operands: [10, 2] })

    expect(res).toEqual({ operation: 'divide', operands: [10, 2], result: 5 })
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"operation":"divide","operands":[10,2]}',
    })
  })

  it('maps a backend error body to ApiError with its code and status', async () => {
    mockFetch(() => jsonResponse({ error: { code: 'DIVISION_BY_ZERO', message: 'divide: division by zero' } }, 422))

    const err = await caught(calculate({ operation: 'divide', operands: [1, 0] }))

    expect(err.code).toBe('DIVISION_BY_ZERO')
    expect(err.message).toBe('divide: division by zero')
    expect(err.status).toBe(422)
  })

  it('maps a network failure to NETWORK_ERROR', async () => {
    mockFetch(() => Promise.reject(new TypeError('Failed to fetch')))

    const err = await caught(calculate({ operation: 'add', operands: [1, 2] }))

    expect(err.code).toBe('NETWORK_ERROR')
    expect(err.status).toBeUndefined()
  })

  it('maps a non-JSON error response to UNEXPECTED_RESPONSE', async () => {
    mockFetch(() => Promise.resolve(new Response('Bad Gateway', { status: 502 })))

    const err = await caught(calculate({ operation: 'add', operands: [1, 2] }))

    expect(err.code).toBe('UNEXPECTED_RESPONSE')
    expect(err.status).toBe(502)
  })

  it('rejects a 200 response that does not match the contract', async () => {
    mockFetch(() => jsonResponse({ result: 'five' }, 200))

    const err = await caught(calculate({ operation: 'add', operands: [2, 3] }))

    expect(err.code).toBe('UNEXPECTED_RESPONSE')
  })
})
