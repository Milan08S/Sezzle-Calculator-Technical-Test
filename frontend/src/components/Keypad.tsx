import type { Action } from '../state/calculatorReducer'
import styles from './Keypad.module.css'

interface Key {
  label: string
  name: string // accessible name
  action: Action
  variant?: 'operator' | 'function' | 'equals' | 'wide'
}

const KEYS: Key[] = [
  { label: 'C', name: 'Clear', action: { type: 'clear' }, variant: 'function' },
  { label: '⌫', name: 'Backspace', action: { type: 'backspace' }, variant: 'function' },
  { label: '±', name: 'Toggle sign', action: { type: 'toggleSign' }, variant: 'function' },
  { label: '÷', name: 'Divide', action: { type: 'operator', operator: 'divide' }, variant: 'operator' },
  { label: '√', name: 'Square root', action: { type: 'sqrt' }, variant: 'function' },
  { label: 'xʸ', name: 'Power', action: { type: 'operator', operator: 'power' }, variant: 'function' },
  { label: '%', name: 'Percent of', action: { type: 'operator', operator: 'percent' }, variant: 'function' },
  { label: '×', name: 'Multiply', action: { type: 'operator', operator: 'multiply' }, variant: 'operator' },
  { label: '7', name: '7', action: { type: 'digit', digit: '7' } },
  { label: '8', name: '8', action: { type: 'digit', digit: '8' } },
  { label: '9', name: '9', action: { type: 'digit', digit: '9' } },
  { label: '−', name: 'Subtract', action: { type: 'operator', operator: 'subtract' }, variant: 'operator' },
  { label: '4', name: '4', action: { type: 'digit', digit: '4' } },
  { label: '5', name: '5', action: { type: 'digit', digit: '5' } },
  { label: '6', name: '6', action: { type: 'digit', digit: '6' } },
  { label: '+', name: 'Add', action: { type: 'operator', operator: 'add' }, variant: 'operator' },
  { label: '1', name: '1', action: { type: 'digit', digit: '1' } },
  { label: '2', name: '2', action: { type: 'digit', digit: '2' } },
  { label: '3', name: '3', action: { type: 'digit', digit: '3' } },
  { label: '=', name: 'Equals', action: { type: 'equals' }, variant: 'equals' },
  { label: '0', name: '0', action: { type: 'digit', digit: '0' }, variant: 'wide' },
  { label: '.', name: 'Decimal point', action: { type: 'decimal' } },
]

export function Keypad({ onAction, disabled }: { onAction: (action: Action) => void; disabled: boolean }) {
  return (
    <div className={styles.keypad}>
      {KEYS.map((key) => (
        <button
          key={key.name}
          type="button"
          aria-label={key.name}
          className={`${styles.key} ${key.variant ? styles[key.variant] : ''}`}
          disabled={disabled}
          onClick={() => onAction(key.action)}
        >
          {key.label}
        </button>
      ))}
    </div>
  )
}
