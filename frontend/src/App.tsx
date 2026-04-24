import { BrowserRouter, Routes, Route, Navigate, Link, useNavigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './AuthContext';
import LoginPage from './pages/LoginPage';
import CaseLibraryPage from './pages/CaseLibraryPage';
import ViewerPage from './pages/ViewerPage';
import ClassroomPage from './pages/ClassroomPage';
import DashboardPage from './pages/DashboardPage';
import './index.css';

function Header() {
  const { user, logout } = useAuth();
  const nav = useNavigate();
  if (!user) return null;
  return (
    <header className="header">
      <div className="header-inner">
        <Link to="/" className="logo">
          <span className="logo-icon">🏥</span>
          <span>RadiMaster</span>
        </Link>
        <nav className="header-nav">
          <Link to="/" className="nav-link">Case Library</Link>
          <Link to="/classrooms" className="nav-link">Classrooms</Link>
          <Link to="/dashboard" className="nav-link">Dashboard</Link>
          <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>
            {user.name} ({user.role})
          </span>
          <button className="btn btn-secondary btn-sm" onClick={() => { logout(); nav('/login'); }}>
            Logout
          </button>
        </nav>
      </div>
    </header>
  );
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  if (loading) return <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-muted)' }}>Loading...</div>;
  return user ? <>{children}</> : <Navigate to="/login" />;
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Header />
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/" element={<ProtectedRoute><CaseLibraryPage /></ProtectedRoute>} />
          <Route path="/cases/:id" element={<ProtectedRoute><ViewerPage /></ProtectedRoute>} />
          <Route path="/classrooms" element={<ProtectedRoute><ClassroomPage /></ProtectedRoute>} />
          <Route path="/classrooms/:code" element={<ProtectedRoute><ClassroomPage /></ProtectedRoute>} />
          <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
export default App;
