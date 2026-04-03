import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from './App'

describe('App', () => {
  it('renders the task tracker heading', () => {
    render(<App />)
    expect(screen.getByText(/task tracker/i)).toBeInTheDocument()
  })
})
