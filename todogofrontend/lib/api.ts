const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';

export interface User {
  id: number;
  email: string;
  created_at: string;
}

export interface Todo {
  id: number;
  user_id: number;
  title: string;
  completed: boolean;
  created_at: string;
}

export interface LoginResponse {
  session_id: string;
  user_id: number;
  email: string;
}

export interface ApiError extends Error {
  status: number;
}

const createApiError = (status: number, message: string): ApiError => {
  const error = new Error(message) as ApiError;
  error.status = status;
  error.name = 'ApiError';
  return error;
};

const getSessionId = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('sessionId');
};

const setSessionId = (sessionId: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('sessionId', sessionId);
  }
};

const clearSessionId = (): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('sessionId');
  }
};

export const login = async (email: string, password: string): Promise<LoginResponse> => {
  const response = await fetch(`${API_BASE}/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    throw createApiError(response.status, await response.text());
  }

  const data: LoginResponse = await response.json();
  setSessionId(data.session_id);
  return data;
};

export const getTodos = async (): Promise<Todo[]> => {
  const sessionId = getSessionId();
  if (!sessionId) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE}/todos`, {
    headers: {
      'Authorization': sessionId,
    },
  });

  if (!response.ok) {
    throw createApiError(response.status, await response.text());
  }

  return response.json();
};

export const createTodo = async (title: string): Promise<Todo> => {
  const sessionId = getSessionId();
  if (!sessionId) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE}/todos`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': sessionId,
    },
    body: JSON.stringify({ title }),
  });

  if (!response.ok) {
    throw createApiError(response.status, await response.text());
  }

  return response.json();
};

export const logout = (): void => {
  clearSessionId();
};

export const isAuthenticated = (): boolean => {
  return !!getSessionId();
};