import { BrowserRouter, Route, Routes } from 'react-router-dom'
import ProtectedRoute from './components/ProtectedRoute'
import LoginPage from './pages/LoginPage'
import SnippetEditorPage from './pages/SnippetEditorPage'
import SnippetListPage from './pages/SnippetListPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <SnippetListPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/snippets/new"
          element={
            <ProtectedRoute>
              <SnippetEditorPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/snippets/:id"
          element={
            <ProtectedRoute>
              <SnippetEditorPage />
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
