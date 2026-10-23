import { useEffect, useReducer, useState } from 'react'
import { ApiError, calculate } from '../api/calculatorClient'
import { calculatorReducer, initialState } from '../state/calculatorReducer'
import styles from './Calculator.module.css'
import { Display } from './Display'
import { findKey, Keypad } from './Keypad'

export function Calculator() {
  const [state, dispatch] = useReducer(calculatorReducer, initialState)
  const [pressed, setPressed] = useState<string | null>(null)
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
      const key = findKey(e.key)
      if (!key) return
      e.preventDefault() // stops Enter from also "clicking" a focused button
      dispatch(key.action)
      setPressed(key.name)
    }
    // Releasing any key (or leaving the window) ends the pressed animation.
    const release = () => setPressed(null)
    window.addEventListener('keydown', onKeyDown)
    window.addEventListener('keyup', release)
    window.addEventListener('blur', release)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('keyup', release)
      window.removeEventListener('blur', release)
    }
  }, [])

  return (
    <section className={styles.calculator} aria-label="Calculator">
      <Display state={state} />
      <Keypad onAction={dispatch} disabled={state.status === 'loading'} pressed={pressed} />
    </section>
  )
}
