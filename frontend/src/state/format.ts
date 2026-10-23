/**
 * Formats a backend result for display: 12 significant digits hides float64 noise
 * (0.1 + 0.2 → "0.3"), and Number() strips the trailing zeros toPrecision adds.
 */
export function formatNumber(n: number): string {
  return String(Number(n.toPrecision(12)))
}
