import { useEffect, useReducer } from 'react'
import { ApiError, calculate } from '../api/calculatorClient'
import { calculatorReducer, initialState, type Action, type Digit } from '../state/calculatorReducer'
import styles from './Calculator.module.css'
import { Display } from './Display'
import { Keypad } from './Keypad'

const KEY_ACTIONS = new Map<string, Action>([
  ['+', { type: 'operator', operator: 'add' }],
  ['-', { type: 'operator', operator: 'subtract' }],
  ['*', { type: 'operator', operator: 'multiply' }],
  ['/', { type: 'operator', operator: 'divide' }],
  ['^', { type: 'operator', operator: 'power' }],
  ['%', { type: 'operator', operator: 'percent' }],
  ['.', { type: 'decimal' }],
  ['Enter', { type: 'equals' }],
  ['=', { type: 'equals' }],
  ['Backspace', { type: 'backspace' }],
  ['Escape', { type: 'clear' }],
])

function keyToAction(key: string): Action | undefined {
  if (/^[0-9]$/.test(key)) return { type: 'digit', digit: key as Digit }
  return KEY_ACTIONS.get(key)
}

export function Calculator() {
  const [state, dispatch] = useReducer(calculatorReducer, initialState)
  const request = state.status === 'loading' ? state.request : null

  // The reducer decides what to compute; this effect only performs the request.
  useEffect(() => {
    if (!request) return
    let cancelled = false
    calculate(request).then(
      (res) => !cancelled && dispatch({ type: 'resolved', result: res.result }),
      (err: unknown) =>
        !cancelled &&
        dispatch(
          err instanceof ApiError
            ? { type: 'rejected', code: err.code, message: err.message }
            : { type: 'rejected', code: 'UNEXPECTED_RESPONSE', message: String(err) },
        ),
    )
    return () => {
      cancelled = true
    }
  }, [request])

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.ctrlKey || e.metaKey || e.altKey) return
      const action = keyToAction(e.key)
      if (!action) return
      e.preventDefault() // stops Enter from also "clicking" a focused button
      dispatch(action)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  return (
    <section className={styles.calculator} aria-label="Calculator">
      <Display state={state} />
      <Keypad onAction={dispatch} disabled={state.status === 'loading'} />
    </section>
  )
}
