// The only module that talks to the backend. Types mirror the API contract.

export type BinaryOperation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power' | 'percent'
export type UnaryOperation = 'sqrt'
export type Operation = BinaryOperation | UnaryOperation

export interface CalculateRequest {
  operation: Operation
  operands: number[]
}

export interface CalculateResponse {
  operation: Operation
  operands: number[]
  result: number
}

export type ServerErrorCode =
  | 'INVALID_JSON'
  | 'MISSING_FIELD'
  | 'UNKNOWN_OPERATION'
  | 'INVALID_ARITY'
  | 'PAYLOAD_TOO_LARGE'
  | 'DIVISION_BY_ZERO'
  | 'NEGATIVE_SQRT'
  | 'NON_FINITE_RESULT'
  | 'INTERNAL_ERROR'

/** Codes produced by the client itself when no valid server error is available. */
export type ClientErrorCode = 'NETWORK_ERROR' | 'UNEXPECTED_RESPONSE'

export type ErrorCode = ServerErrorCode | ClientErrorCode

export class ApiError extends Error {
  readonly code: ErrorCode
  readonly status?: number

  constructor(code: ErrorCode, message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

// Empty base URL means same origin (Vite proxy in dev, nginx in Docker).
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export async function calculate(request: CalculateRequest): Promise<CalculateResponse> {
  let res: Response
  try {
    res = await fetch(`${BASE_URL}/api/v1/calculate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
  } catch {
    throw new ApiError('NETWORK_ERROR', 'Could not reach the server')
  }

  const body: unknown = await res.json().catch(() => undefined)

  if (!res.ok) {
    if (isErrorBody(body)) {
      throw new ApiError(body.error.code as ServerErrorCode, body.error.message, res.status)
    }
    throw new ApiError('UNEXPECTED_RESPONSE', `Request failed with status ${res.status}`, res.status)
  }
  if (!isCalculateResponse(body)) {
    throw new ApiError('UNEXPECTED_RESPONSE', 'Malformed response from server', res.status)
  }
  return body
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null
}

function isErrorBody(v: unknown): v is { error: { code: string; message: string } } {
  return isObject(v) && isObject(v.error) && typeof v.error.code === 'string' && typeof v.error.message === 'string'
}

function isCalculateResponse(v: unknown): v is CalculateResponse {
  return isObject(v) && typeof v.result === 'number' && typeof v.operation === 'string' && Array.isArray(v.operands)
}
