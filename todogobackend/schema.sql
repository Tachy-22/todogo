-- Database schema for todo app
CREATE DATABASE todoapp;

\c todoapp;

CREATE TABLE todousers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES todousers(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES todousers(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes as specified in README
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_todos_user_id ON todos(user_id);