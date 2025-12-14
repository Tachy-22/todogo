'use client';

import { useState, useEffect } from 'react';
import { getTodos, createTodo, logout, Todo, LoginResponse } from '../lib/api';

interface TodoListProps {
  user: LoginResponse;
  onLogout: () => void;
}

export default function TodoList({ user, onLogout }: TodoListProps) {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [newTodoTitle, setNewTodoTitle] = useState('');
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState('');

  const fetchTodos = async () => {
    setLoading(true);
    setError('');
    try {
      const todoList = await getTodos();
      setTodos(todoList || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch todos');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTodo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTodoTitle.trim()) return;
    
    setCreating(true);
    setError('');
    try {
      const newTodo = await createTodo(newTodoTitle.trim());
      setTodos(prevTodos => [newTodo, ...prevTodos]);
      setNewTodoTitle('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create todo');
    } finally {
      setCreating(false);
    }
  };

  const handleLogout = () => {
    logout();
    onLogout();
  };

  useEffect(() => {
    fetchTodos();
  }, []);

  return (
    <div className="max-w-2xl mx-auto bg-white p-8 rounded-lg shadow-md">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">My Todos</h1>
          <p className="text-gray-600">Welcome, {user.email}</p>
        </div>
        <button
          onClick={handleLogout}
          className="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600"
        >
          Logout
        </button>
      </div>

      <form onSubmit={handleCreateTodo} className="mb-6">
        <div className="flex gap-2">
          <input
            type="text"
            value={newTodoTitle}
            onChange={(e) => setNewTodoTitle(e.target.value)}
            placeholder="What needs to be done?"
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={creating || !newTodoTitle.trim()}
            className="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          >
            {creating ? 'Adding...' : 'Add'}
          </button>
        </div>
      </form>

      {error && (
        <div className="text-red-600 text-sm mb-4 p-3 bg-red-50 rounded-md">
          {error}
        </div>
      )}

      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">Todo Items ({todos.length})</h2>
        <button
          onClick={fetchTodos}
          disabled={loading}
          className="text-blue-600 hover:text-blue-800 disabled:opacity-50"
        >
          {loading ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>

      {loading && todos.length === 0 ? (
        <div className="text-center py-8 text-gray-500">Loading todos...</div>
      ) : todos.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No todos yet. Add one above to get started!
        </div>
      ) : (
        <div className="space-y-3">
          {todos.map((todo) => (
            <div
              key={todo.id}
              className="flex items-center p-3 border border-gray-200 rounded-md hover:bg-gray-50"
            >
              <div className="flex-1">
                <div className="font-medium">{todo.title}</div>
                <div className="text-sm text-gray-500">
                  Created: {new Date(todo.created_at).toLocaleDateString()}
                </div>
              </div>
              <div className={`px-2 py-1 rounded text-sm ${
                todo.completed 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-yellow-100 text-yellow-800'
              }`}>
                {todo.completed ? 'Done' : 'Pending'}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}