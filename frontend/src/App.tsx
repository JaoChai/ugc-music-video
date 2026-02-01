import { QueryProvider } from '@/providers/QueryProvider'
import { AppRouter } from '@/router'

function App() {
  return (
    <QueryProvider>
      <AppRouter />
    </QueryProvider>
  )
}

export default App
