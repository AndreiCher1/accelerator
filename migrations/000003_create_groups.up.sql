CREATE TABLE groups (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_by REFERENCES users(id) ON DELETE SET NULL, -- при удалении админа, который создал группу она сохраняется, но изменять ее название и описание теперь сможет только креатор
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_groups_id ON groups(id);