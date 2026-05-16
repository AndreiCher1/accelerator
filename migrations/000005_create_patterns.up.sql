CREATE TABLE patterns (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- шаблоны привязываем к группе, чтобы юзеры могли их видеть, если группа NULL, значит это глабальный шаблон и его видят все
    name VARCHAR(255) NOT NULL,
    description TEXT,
    summary_prompt TEXT NOT NULL,
    additional_prompt JSONB,
    created_by UUID REFERENCES users(id) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_patterns_id ON patterns(id); 