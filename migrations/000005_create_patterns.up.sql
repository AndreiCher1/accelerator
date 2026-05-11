CREATE TABLE patterns (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- шаблоны привязаны к админу, который их создал
    pattern_name VARCHAR(255) NOT NULL,
    description TEXT,
    summary_prompt TEXT NOT NULL, -- сам получившийся промпт шаблона
    additional_prompt TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_patterns_id ON patterns(id); 