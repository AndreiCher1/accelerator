CREATE TABLE groups (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    owner_id UUID REFERENCES users(id),  -- администратор группы (admin)
    created_by UUID REFERENCES users(id) NOT NULL, -- при удалении админа, который создал группу она сохраняется, но изменять ее название и описание теперь сможет только креатор
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_groups_id ON groups(id);