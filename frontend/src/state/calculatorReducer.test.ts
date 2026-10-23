import { describe, expect, it } from 'vitest'
import { calculatorReducer, initialState, MAX_DIGITS, type Action, type Digit, type State } from './calculatorReducer'

const run = (actions: Action[], from: State = initialState) => actions.reduce(calculatorReducer, from)
const digits = (s: string): Action[] => [...s].map((d) => (d === '.' ? { type: 'decimal' } : { type: 'digit', digit: d as Digit }))

const first = (input: string): State => ({ status: 'inputtingFirst', input })
const second = (input: string): State => ({ status: 'inputtingSecond', first: 5, operator: 'add', input })
const pending: State = { status: 'operatorPending', first: 5, operator: 'add' }
const loading: State = { status: 'loading', request: { operation: 'add', operands: [5, 3] } }
const result: State = { status: 'result', value: 8 }
const error: State = { status: 'error', code: 'DIVISION_BY_ZERO', message: 'division by zero' }

describe('digit input', () => {
  it.each<[string, State, Action[], State]>([
    ['idle starts the first operand', initialState, digits('5'), first('5')],
    ['appends digits', first('1'), digits('23'), first('123')],
    ['replaces a leading zero', first('0'), digits('7'), first('7')],
    ['replaces a negative leading zero', first('-0'), digits('7'), first('-7')],
    ['keeps zeros after the decimal point', first('0.'), digits('05'), first('0.05')],
    ['operatorPending starts the second operand', pending, digits('3'), second('3')],
    ['appends to the second operand', second('3'), digits('4'), second('34')],
    ['result starts a new calculation', result, digits('9'), first('9')],
    ['error starts a new calculation', error, digits('9'), first('9')],
  ])('%s', (_, from, actions, want) => {
    expect(run(actions, from)).toEqual(want)
  })

  it(`ignores digits beyond ${MAX_DIGITS}`, () => {
    const full = '1'.repeat(MAX_DIGITS)
    expect(run(digits('1'), first(full))).toEqual(first(full))
    expect(run(digits('1'), first('-' + full))).toEqual(first('-' + full))
  })

  it('does not count the decimal point toward the limit', () => {
    const almost = '1'.repeat(MAX_DIGITS - 1) + '.'
    expect(run(digits('2'), first(almost))).toEqual(first(almost + '2'))
  })
})

describe('decimal point', () => {
  it.each<[string, State, State]>([
    ['idle starts with 0.', initialState, first('0.')],
    ['appends to an integer', first('12'), first('12.')],
    ['rejects a second decimal point', first('1.5'), first('1.5')],
    ['keeps the sign on -0', first('-0'), first('-0.')],
    ['starts the second operand with 0.', pending, second('0.')],
    ['result starts a new number', result, first('0.')],
  ])('%s', (_, from, want) => {
    expect(calculatorReducer(from, { type: 'decimal' })).toEqual(want)
  })
})

describe('sign toggle', () => {
  it.each<[string, State, State]>([
    ['negates the input', first('5'), first('-5')],
    ['removes the minus', first('-5'), first('5')],
    ['idle starts with -0', initialState, first('-0')],
    ['starts a negative second operand', pending, second('-0')],
    ['negates the second operand', second('3'), second('-3')],
    ['negates a result', result, { status: 'result', value: -8 }],
  ])('%s', (_, from, want) => {
    expect(calculatorReducer(from, { type: 'toggleSign' })).toEqual(want)
  })
})

describe('backspace', () => {
  it.each<[string, State, State]>([
    ['removes the last character', first('123'), first('12')],
    ['removes the decimal point', first('1.'), first('1')],
    ['leaves 0 when the first operand empties', first('5'), first('0')],
    ['leaves 0 when only a minus would remain', first('-5'), first('0')],
    ['leaves 0 in the second operand instead of showing the first', second('3'), second('0')],
    ['does nothing on a result', result, result],
    ['does nothing while an operator is pending', pending, pending],
  ])('%s', (_, from, want) => {
    expect(calculatorReducer(from, { type: 'backspace' })).toEqual(want)
  })
})

describe('clear', () => {
  it.each<[string, State]>([
    ['first operand', first('12')],
    ['second operand', second('3')],
    ['result', result],
    ['error', error],
    ['loading', loading],
  ])('resets from %s', (_, from) => {
    expect(calculatorReducer(from, { type: 'clear' })).toEqual(initialState)
  })
})

describe('binary operators', () => {
  it.each<[string, State, State]>([
    ['idle uses 0 as the first operand', initialState, { status: 'operatorPending', first: 0, operator: 'multiply' }],
    ['commits the first operand', first('-2.5'), { status: 'operatorPending', first: -2.5, operator: 'multiply' }],
    ['replaces a pending operator', pending, { status: 'operatorPending', first: 5, operator: 'multiply' }],
    ['chains from a result', result, { status: 'operatorPending', first: 8, operator: 'multiply' }],
    [
      'computes the pending operation before chaining',
      second('3'),
      { status: 'loading', request: { operation: 'add', operands: [5, 3] }, nextOperator: 'multiply' },
    ],
    ['is ignored in error', error, error],
  ])('%s', (_, from, want) => {
    expect(calculatorReducer(from, { type: 'operator', operator: 'multiply' })).toEqual(want)
  })
})

describe('equals', () => {
  it('requests the pending binary operation', () => {
    expect(calculatorReducer(second('3'), { type: 'equals' })).toEqual(loading)
  })

  it.each<[string, State]>([
    ['idle', initialState],
    ['first operand only', first('5')],
    ['empty second operand', pending],
    ['result', result],
    ['error', error],
  ])('does nothing with %s', (_, from) => {
    expect(calculatorReducer(from, { type: 'equals' })).toEqual(from)
  })
})

describe('sqrt', () => {
  it.each<[string, State, number]>([
    ['of the first operand', first('16'), 16],
    ['of a result', result, 8],
    ['of idle (0)', initialState, 0],
  ])('%s', (_, from, operand) => {
    expect(calculatorReducer(from, { type: 'sqrt' })).toEqual({
      status: 'loading',
      request: { operation: 'sqrt', operands: [operand] },
    })
  })

  it.each<[string, State]>([
    ['pending operator', pending],
    ['second operand', second('3')],
    ['error', error],
  ])('is ignored with %s', (_, from) => {
    expect(calculatorReducer(from, { type: 'sqrt' })).toEqual(from)
  })
})

describe('loading', () => {
  it('shows the result when resolved', () => {
    expect(calculatorReducer(loading, { type: 'resolved', result: 8 })).toEqual(result)
  })

  it('chains into the next operator when one was pressed', () => {
    const chaining: State = { ...loading, nextOperator: 'divide' }
    expect(calculatorReducer(chaining, { type: 'resolved', result: 8 })).toEqual({
      status: 'operatorPending',
      first: 8,
      operator: 'divide',
    })
  })

  it('stores the error when rejected', () => {
    expect(
      calculatorReducer(loading, { type: 'rejected', code: 'DIVISION_BY_ZERO', message: 'division by zero' }),
    ).toEqual(error)
  })

  it.each<Action>([
    { type: 'digit', digit: '1' },
    { type: 'decimal' },
    { type: 'toggleSign' },
    { type: 'backspace' },
    { type: 'operator', operator: 'add' },
    { type: 'sqrt' },
    { type: 'equals' },
  ])('ignores $type', (action) => {
    expect(calculatorReducer(loading, action)).toBe(loading)
  })

  it.each<[string, State]>([
    ['idle', initialState],
    ['result', result],
  ])('ignores stale responses in %s', (_, from) => {
    expect(calculatorReducer(from, { type: 'resolved', result: 1 })).toBe(from)
    expect(calculatorReducer(from, { type: 'rejected', code: 'NETWORK_ERROR', message: 'x' })).toBe(from)
  })
})

describe('full flows', () => {
  it('5 + 3 = then chains × 2', () => {
    let s = run([...digits('5'), { type: 'operator', operator: 'add' }, ...digits('3'), { type: 'equals' }])
    expect(s).toEqual(loading)
    s = run([{ type: 'resolved', result: 8 }, { type: 'operator', operator: 'multiply' }, ...digits('2'), { type: 'equals' }], s)
    expect(s).toEqual({ status: 'loading', request: { operation: 'multiply', operands: [8, 2] } })
  })
})
