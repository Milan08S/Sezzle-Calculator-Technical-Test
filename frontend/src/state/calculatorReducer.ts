import type { BinaryOperation, CalculateRequest, ErrorCode } from '../api/calculatorClient'

/** Max digits per operand; float64 is only exact to ~15 significant digits. */
export const MAX_DIGITS = 15

export type State =
  | { status: 'idle' }
  | { status: 'inputtingFirst'; input: string }
  | { status: 'operatorPending'; first: number; operator: BinaryOperation }
  | { status: 'inputtingSecond'; first: number; operator: BinaryOperation; input: string }
  // nextOperator is set when an operator key triggered the request ("5 + 3 ×"): the result becomes the next first operand.
  | { status: 'loading'; request: CalculateRequest; nextOperator?: BinaryOperation }
  | { status: 'result'; value: number }
  | { status: 'error'; code: ErrorCode; message: string }

export type Digit = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'

export type Action =
  | { type: 'digit'; digit: Digit }
  | { type: 'decimal' }
  | { type: 'toggleSign' }
  | { type: 'backspace' }
  | { type: 'clear' }
  | { type: 'operator'; operator: BinaryOperation }
  | { type: 'sqrt' }
  | { type: 'equals' }
  | { type: 'resolved'; result: number }
  | { type: 'rejected'; code: ErrorCode; message: string }

export const initialState: State = { status: 'idle' }

export function calculatorReducer(state: State, action: Action): State {
  if (action.type === 'clear') return initialState
  if (state.status === 'loading') return settle(state, action)

  switch (action.type) {
    case 'toggleSign':
      if (state.status === 'result') return { status: 'result', value: -state.value }
      return edit(state, (input) => editInput(input, action))
    case 'digit':
    case 'decimal':
      return edit(state, (input) => editInput(input, action))
    case 'backspace':
      return backspace(state)
    case 'operator':
      return pressOperator(state, action.operator)
    case 'sqrt':
      return pressSqrt(state)
    case 'equals':
      return state.status === 'inputtingSecond'
        ? { status: 'loading', request: { operation: state.operator, operands: [state.first, Number(state.input)] } }
        : state
    default:
      return state // resolved/rejected outside loading are stale responses
  }
}

function settle(state: Extract<State, { status: 'loading' }>, action: Action): State {
  switch (action.type) {
    case 'resolved':
      return state.nextOperator
        ? { status: 'operatorPending', first: action.result, operator: state.nextOperator }
        : { status: 'result', value: action.result }
    case 'rejected':
      return { status: 'error', code: action.code, message: action.message }
    default:
      return state // input is disabled while loading
  }
}

/** Applies an input edit to whichever operand is being typed, starting a new one if needed. */
function edit(state: State, change: (input: string) => string): State {
  switch (state.status) {
    case 'inputtingFirst':
    case 'inputtingSecond':
      return { ...state, input: change(state.input) }
    case 'operatorPending':
      return { status: 'inputtingSecond', first: state.first, operator: state.operator, input: change('') }
    default: // idle, result, error: start a fresh calculation
      return { status: 'inputtingFirst', input: change('') }
  }
}

function editInput(input: string, action: Action): string {
  const negative = input.startsWith('-')
  const body = negative ? input.slice(1) : input
  const sign = negative ? '-' : ''

  switch (action.type) {
    case 'digit': {
      if (body === '' || body === '0') return sign + action.digit
      if (body.replace('.', '').length >= MAX_DIGITS) return input
      return input + action.digit
    }
    case 'decimal':
      if (body.includes('.')) return input
      return sign + (body === '' ? '0' : body) + '.'
    case 'toggleSign':
      return negative ? body : '-' + (body === '' ? '0' : body)
    default:
      return input
  }
}

// Deleting the last digit leaves 0 (like a pocket calculator), so the operand never shows a stale value.
function backspace(state: State): State {
  if (state.status !== 'inputtingFirst' && state.status !== 'inputtingSecond') return state
  const input = state.input.slice(0, -1)
  return { ...state, input: input === '' || input === '-' ? '0' : input }
}

function pressOperator(state: State, operator: BinaryOperation): State {
  switch (state.status) {
    case 'idle':
      return { status: 'operatorPending', first: 0, operator }
    case 'inputtingFirst':
      return { status: 'operatorPending', first: Number(state.input), operator }
    case 'operatorPending':
      return { ...state, operator }
    case 'inputtingSecond':
      return {
        status: 'loading',
        request: { operation: state.operator, operands: [state.first, Number(state.input)] },
        nextOperator: operator,
      }
    case 'result':
      return { status: 'operatorPending', first: state.value, operator }
    default: // error: user must type a number or clear
      return state
  }
}

function pressSqrt(state: State): State {
  switch (state.status) {
    case 'idle':
      return { status: 'loading', request: { operation: 'sqrt', operands: [0] } }
    case 'inputtingFirst':
      return { status: 'loading', request: { operation: 'sqrt', operands: [Number(state.input)] } }
    case 'result':
      return { status: 'loading', request: { operation: 'sqrt', operands: [state.value] } }
    default: // sqrt applies to the current value only, not to a pending second operand
      return state
  }
}
