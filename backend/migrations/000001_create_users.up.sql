CREATE TABLE users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(256) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    position VARCHAR(100) NOT NULL,
    role VARCHAR(10) NOT NULL,
    temporary_password BOOLEAN NOT NULL DEFAULT 'true', -- изначально нуждается в замене после регистрации
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_users_user_id ON users(id); 