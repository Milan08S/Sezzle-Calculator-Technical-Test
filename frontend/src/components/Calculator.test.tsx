import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError, calculate, type CalculateRequest } from '../api/calculatorClient'
import { Calculator } from './Calculator'

// Mock only the network call; keep the real ApiError class.
vi.mock('../api/calculatorClient', async (importActual) => ({
  ...(await importActual<typeof import('../api/calculatorClient')>()),
  calculate: vi.fn(),
}))
const calculateMock = vi.mocked(calculate)

/** A tiny fake backend for the happy path. */
function respondWith(compute: (req: CalculateRequest) => number) {
  calculateMock.mockImplementation(async (req) => ({ ...req, result: compute(req) }))
}

function setup() {
  const user = userEvent.setup()
  render(<Calculator />)
  const display = screen.getByRole('status')
  const press = async (...names: string[]) => {
    for (const name of names) await user.click(screen.getByRole('button', { name }))
  }
  return { user, display, press }
}

beforeEach(() => {
  calculateMock.mockReset()
})

describe('Calculator', () => {
  it('computes 5 + 3 = 8', async () => {
    respondWith(({ operands: [a, b] }) => a + b)
    const { display, press } = setup()

    await press('5', 'Add', '3', 'Equals')

    expect(await within(display).findByText('8')).toBeInTheDocument()
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'add', operands: [5, 3] })
  })

  it('shows the division by zero error', async () => {
    calculateMock.mockRejectedValue(new ApiError('DIVISION_BY_ZERO', 'divide: division by zero', 422))
    const { display, press } = setup()

    await press('8', 'Divide', '0', 'Equals')

    expect(await within(display).findByText('Cannot divide by zero')).toBeInTheDocument()
  })

  it('falls back to the backend message for codes without friendly text', async () => {
    calculateMock.mockRejectedValue(new ApiError('INVALID_ARITY', 'sqrt expects 1 operand(s), got 2', 400))
    const { display, press } = setup()

    await press('9', 'Square root')

    expect(await within(display).findByText('sqrt expects 1 operand(s), got 2')).toBeInTheDocument()
  })

  it('shows a network error', async () => {
    calculateMock.mockRejectedValue(new ApiError('NETWORK_ERROR', 'Could not reach the server'))
    const { display, press } = setup()

    await press('1', 'Add', '1', 'Equals')

    expect(await within(display).findByText('Cannot reach the server')).toBeInTheDocument()
  })

  it('supports keyboard input', async () => {
    respondWith(({ operands: [a, b] }) => a * b)
    const { user, display } = setup()

    await user.keyboard('12*3{Enter}')

    expect(await within(display).findByText('36')).toBeInTheDocument()
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'multiply', operands: [12, 3] })
  })

  it('supports Backspace, Escape and = from the keyboard', async () => {
    respondWith(({ operands: [a, b] }) => a - b)
    const { user, display } = setup()

    await user.keyboard('123{Backspace}')
    expect(within(display).getByText('12')).toBeInTheDocument()

    await user.keyboard('{Escape}')
    expect(within(display).getByText('0')).toBeInTheDocument()

    await user.keyboard('9-4=')
    expect(await within(display).findByText('5')).toBeInTheDocument()
  })

  it('disables input while loading and re-enables it with the result', async () => {
    let resolve!: (n: number) => void
    calculateMock.mockImplementation(
      (req) => new Promise((r) => (resolve = (result) => r({ ...req, result }))),
    )
    const { user, display, press } = setup()

    await press('2', 'Power', '3', 'Equals')

    expect(screen.getByRole('button', { name: '7' })).toBeDisabled()
    expect(display).toHaveAttribute('aria-busy', 'true')
    expect(display).toHaveTextContent('2 ^ 3 =')
    await user.keyboard('9') // ignored while loading
    resolve(8)

    expect(await within(display).findByText('8')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '7' })).toBeEnabled()
    expect(calculateMock).toHaveBeenCalledTimes(1)
  })

  it('ignores a late response after Clear', async () => {
    let resolve!: (n: number) => void
    calculateMock.mockImplementation(
      (req) => new Promise((r) => (resolve = (result) => r({ ...req, result }))),
    )
    const { user, display, press } = setup()

    await press('2', 'Add', '2', 'Equals')
    await user.keyboard('{Escape}')
    resolve(4)

    await vi.waitFor(() => expect(within(display).getByText('0')).toBeInTheDocument())
    expect(within(display).queryByText('4')).not.toBeInTheDocument()
  })

  it('chains from a result and hides float noise', async () => {
    respondWith(({ operation, operands: [a, b] }) => (operation === 'add' ? a + b : a * b))
    const { display, press } = setup()

    await press('0', 'Decimal point', '1', 'Add', '0', 'Decimal point', '2', 'Equals')
    expect(await within(display).findByText('0.3')).toBeInTheDocument()

    await press('Multiply', '2', 'Equals')
    expect(await within(display).findByText('0.6')).toBeInTheDocument()
    expect(calculateMock).toHaveBeenLastCalledWith({ operation: 'multiply', operands: [0.30000000000000004, 2] })
  })

  it('computes a square root and shows the expression', async () => {
    respondWith(({ operands: [a] }) => Math.sqrt(a))
    const { display, press } = setup()

    await press('1', '6', 'Toggle sign', 'Toggle sign', 'Square root')

    expect(await within(display).findByText('4')).toBeInTheDocument()
    expect(calculateMock).toHaveBeenCalledWith({ operation: 'sqrt', operands: [16] })
  })

  it('shows 0, not the first operand, after deleting the second operand', async () => {
    const { display, press } = setup()

    await press('8', 'Add', '5', 'Backspace')

    expect(display).toHaveTextContent('8 +')
    expect(within(display).getByText('0')).toBeInTheDocument()
  })

  it('shows the pending operator expression', async () => {
    const { display, press } = setup()

    await press('7', 'Percent of')

    expect(display).toHaveTextContent('7 %')
  })

  it('shows a keypad key as pressed while its keyboard key is held', async () => {
    const { user } = setup()
    const five = screen.getByRole('button', { name: '5' })
    const equals = screen.getByRole('button', { name: 'Equals' })

    await user.keyboard('{5>}') // hold
    expect(five).toHaveAttribute('data-pressed')
    expect(equals).not.toHaveAttribute('data-pressed')

    await user.keyboard('{/5}') // release
    expect(five).not.toHaveAttribute('data-pressed')

    await user.keyboard('{Enter>}')
    expect(equals).toHaveAttribute('data-pressed')
  })

  it('gives every key an accessible name', () => {
    setup()
    for (const button of screen.getAllByRole('button')) {
      expect(button).toHaveAccessibleName()
    }
  })
})
