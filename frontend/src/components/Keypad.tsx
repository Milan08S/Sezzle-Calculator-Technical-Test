import type { Action, Digit } from '../state/calculatorReducer'
import styles from './Keypad.module.css'

interface Key {
  label: string
  name: string // accessible name
  action: Action
  shortcuts?: string[] // KeyboardEvent.key values that press this key
  variant?: 'operator' | 'function' | 'equals' | 'wide'
}

const digit = (d: Digit, variant?: Key['variant']): Key => ({
  label: d,
  name: d,
  action: { type: 'digit', digit: d },
  shortcuts: [d],
  variant,
})

const KEYS: Key[] = [
  { label: 'C', name: 'Clear', action: { type: 'clear' }, shortcuts: ['Escape'], variant: 'function' },
  { label: '⌫', name: 'Backspace', action: { type: 'backspace' }, shortcuts: ['Backspace'], variant: 'function' },
  { label: '±', name: 'Toggle sign', action: { type: 'toggleSign' }, variant: 'function' },
  { label: '÷', name: 'Divide', action: { type: 'operator', operator: 'divide' }, shortcuts: ['/'], variant: 'operator' },
  { label: '√', name: 'Square root', action: { type: 'sqrt' }, variant: 'function' },
  { label: 'xʸ', name: 'Power', action: { type: 'operator', operator: 'power' }, shortcuts: ['^'], variant: 'function' },
  { label: '%', name: 'Percent of', action: { type: 'operator', operator: 'percent' }, shortcuts: ['%'], variant: 'function' },
  { label: '×', name: 'Multiply', action: { type: 'operator', operator: 'multiply' }, shortcuts: ['*'], variant: 'operator' },
  digit('7'),
  digit('8'),
  digit('9'),
  { label: '−', name: 'Subtract', action: { type: 'operator', operator: 'subtract' }, shortcuts: ['-'], variant: 'operator' },
  digit('4'),
  digit('5'),
  digit('6'),
  { label: '+', name: 'Add', action: { type: 'operator', operator: 'add' }, shortcuts: ['+'], variant: 'operator' },
  digit('1'),
  digit('2'),
  digit('3'),
  { label: '=', name: 'Equals', action: { type: 'equals' }, shortcuts: ['Enter', '='], variant: 'equals' },
  digit('0', 'wide'),
  { label: '.', name: 'Decimal point', action: { type: 'decimal' }, shortcuts: ['.'] },
]

/** The keypad key a physical keyboard key maps to, if any. */
export function findKey(shortcut: string): Key | undefined {
  return KEYS.find((k) => k.shortcuts?.includes(shortcut))
}

interface KeypadProps {
  onAction: (action: Action) => void
  disabled: boolean
  pressed: string | null // name of the key held on the physical keyboard
}

export function Keypad({ onAction, disabled, pressed }: KeypadProps) {
  return (
    <div className={styles.keypad}>
      {KEYS.map((key) => (
        <button
          key={key.name}
          type="button"
          aria-label={key.name}
          className={`${styles.key} ${key.variant ? styles[key.variant] : ''}`}
          data-pressed={pressed === key.name || undefined}
          disabled={disabled}
          onClick={() => onAction(key.action)}
        >
          {key.label}
        </button>
      ))}
    </div>
  )
}
