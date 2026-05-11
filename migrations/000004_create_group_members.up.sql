CREATE TABLE group_members (
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- если удалится группа, то все строки с участниками тоже удалятся
    added_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- если удалится пользователь, то и из участников группы он исчезнет
    PRIMARY KEY (group_id, user_id)  --
);

CREATE INDEX idx_group_members_user_id ON group_members(user_id);
-- надо пофиксить
-- частичный уникальный индекс, не может быть больше 1 admin в группе, чтобы избежать гонки данных в AddUserGroupService между проверкой, есть ли админ и вставкой