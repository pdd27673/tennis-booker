import { describe, it, expect } from 'vitest'

// Simple utility test to verify testing framework is working
describe('Frontend Testing Infrastructure', () => {
  it('can run basic tests', () => {
    expect(1 + 1).toBe(2)
  })

  it('can test string operations', () => {
    const testString = 'Tennis Court Booking'
    expect(testString).toContain('Tennis')
    expect(testString.toLowerCase()).toBe('tennis court booking')
  })

  it('can test array operations', () => {
    const testArray = ['court1', 'court2', 'court3']
    expect(testArray).toHaveLength(3)
    expect(testArray).toContain('court2')
  })
}) 