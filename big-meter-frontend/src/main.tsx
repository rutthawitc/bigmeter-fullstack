import React, { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import AppRouter from './routes'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from './lib/auth'
import { ErrorBoundary } from './lib/ErrorBoundary'
import { ToastProvider } from './lib/useToast'
import './styles/index.css'

const el = document.getElementById('root')!
const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 60_000, refetchOnWindowFocus: false } },
})

createRoot(el).render(
  <StrictMode>
    <ErrorBoundary>
      <ToastProvider>
        <QueryClientProvider client={queryClient}>
          <AuthProvider>
            <AppRouter />
          </AuthProvider>
        </QueryClientProvider>
      </ToastProvider>
    </ErrorBoundary>
  </StrictMode>
)
