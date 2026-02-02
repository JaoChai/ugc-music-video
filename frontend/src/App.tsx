import { QueryProvider } from '@/providers/QueryProvider'
import { AppRouter } from '@/router'
import { ErrorBoundary } from '@/components'

function App() {
  return (
    <ErrorBoundary>
      <QueryProvider>
        <AppRouter />
      </QueryProvider>
    </ErrorBoundary>
  )
}

export default App
