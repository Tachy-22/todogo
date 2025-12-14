'use client';

import { useState, useEffect } from 'react';
import LoginForm from '../components/LoginForm';
import TodoList from '../components/TodoList';
import { LoginResponse, isAuthenticated } from '../lib/api';

export default function Home() {
  const [user, setUser] = useState<LoginResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (isAuthenticated()) {
      const userData = {
        session_id: localStorage.getItem('sessionId') || '',
        user_id: 0,
        email: localStorage.getItem('userEmail') || 'user@example.com'
      };
      setUser(userData);
    }
    setLoading(false);
  }, []);

  const handleLoginSuccess = (response: LoginResponse) => {
    localStorage.setItem('userEmail', response.email);
    setUser(response);
  };

  const handleLogout = () => {
    localStorage.removeItem('userEmail');
    setUser(null);
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        {user ? (
          <TodoList user={user} onLogout={handleLogout} />
        ) : (
          <LoginForm onLoginSuccess={handleLoginSuccess} />
        )}
      </div>
    </div>
  );
}
