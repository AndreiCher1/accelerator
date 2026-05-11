CREATE TABLE group_members (
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- если удалится группа, то все строки с участниками тоже удалятся
    added_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE -- если удалится пользователь, то и из участников группы он исчезнет
);

CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE UNIQUE INDEX unique_admin_in_group ON group_members (group_id) -- частичный уникальный индекс, не может быть больше 1 admin в группе, чтобы избежать гонки данных в AddUserGroupService между проверкой, есть ли админ и вставкой
WHERE (user_id IN (SELECT id FROM users WHERE role = 'admin'));