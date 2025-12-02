import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Home } from './pages/Home';
import { Login } from './pages/Login';
import { AdminDashboard } from './pages/AdminDashboard';
import { Dashboard } from './pages/Dashboard';
import { Settings } from './pages/Settings';
import { ScheduledTasks } from './pages/ScheduledTasks';
import { ProtectedRoute } from './components/ProtectedRoute';
import { AuthProvider } from './contexts/AuthContext';
import { Toaster } from './components/ui/sonner';

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* 公开路由 - 首页 */}
          <Route path="/" element={<Home />} />
          
          {/* 管理后台登录页 */}
          <Route path="/admin/login" element={<Login />} />
          
          {/* 受保护的管理后台路由 */}
          <Route
            path="/admin"
            element={
              <ProtectedRoute>
                <AdminDashboard />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="settings" element={<Settings />} />
            <Route path="scheduled-tasks" element={<ScheduledTasks />} />
          </Route>
          
          {/* 404重定向到首页 */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
      <Toaster />
    </AuthProvider>
  );
}

export default App;

