CREATE TABLE patterns (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- если удалится пользователь, то его шаблоны тоже удалятся
    pattern_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_patterns_id ON patterns(id); 