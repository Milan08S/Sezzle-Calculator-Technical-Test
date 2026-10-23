import type { BinaryOperation, ErrorCode, Operation } from '../api/calculatorClient'
import type { State } from '../state/calculatorReducer'
import { formatNumber } from '../state/format'
import styles from './Display.module.css'

const SYMBOLS: Record<BinaryOperation, string> = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: '^',
  percent: '%',
}

// Friendly text per code; anything else falls back to the backend message.
const ERROR_MESSAGES: Partial<Record<ErrorCode, string>> = {
  DIVISION_BY_ZERO: 'Cannot divide by zero',
  NEGATIVE_SQRT: 'Cannot take the square root of a negative number',
  NON_FINITE_RESULT: 'Result is out of range',
  NETWORK_ERROR: 'Cannot reach the server',
  UNEXPECTED_RESPONSE: 'Unexpected response from the server',
  INTERNAL_ERROR: 'Something went wrong, please try again',
}

function expression(operation: Operation, operands: number[]): string {
  const [a, b] = operands.map(formatNumber)
  return operation === 'sqrt' ? `√(${a}) =` : `${a} ${SYMBOLS[operation]} ${b} =`
}

function describe(state: State): { expression: string; value: string } {
  switch (state.status) {
    case 'idle':
      return { expression: '', value: '0' }
    case 'inputtingFirst':
      return { expression: '', value: state.input }
    case 'operatorPending':
      return { expression: `${formatNumber(state.first)} ${SYMBOLS[state.operator]}`, value: formatNumber(state.first) }
    case 'inputtingSecond':
      return { expression: `${formatNumber(state.first)} ${SYMBOLS[state.operator]}`, value: state.input }
    case 'loading':
      return {
        expression: expression(state.request.operation, state.request.operands),
        value: formatNumber(state.request.operands[state.request.operands.length - 1]),
      }
    case 'result':
      return { expression: '', value: formatNumber(state.value) }
    case 'error':
      return { expression: 'Error', value: ERROR_MESSAGES[state.code] ?? state.message }
  }
}

export function Display({ state }: { state: State }) {
  const { expression, value } = describe(state)
  return (
    <output className={styles.display} aria-live="polite" aria-busy={state.status === 'loading'}>
      <span className={styles.expression}>{expression}</span>
      <span className={state.status === 'error' ? styles.error : styles.value}>{value}</span>
    </output>
  )
}
