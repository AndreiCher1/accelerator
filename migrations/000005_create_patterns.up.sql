CREATE TABLE patterns (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- шаблоны привязываем к группе, чтобы юзеры могли их видеть
    name VARCHAR(255) NOT NULL,
    description TEXT,
    summary_prompt TEXT NOT NULL,
    additional_prompt TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL, -- при удалении шаблонов
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_patterns_id ON patterns(id); 