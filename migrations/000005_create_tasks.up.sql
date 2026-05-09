CREATE TABLE tasks (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- при удалении пользователя задача остается
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE, -- при удалении группы все задачи в группе удаляются
    task_name VARCHAR(255) NOT NULL,
    description TEXT,
    meeting_date DATE,
    asr_model VARCHAR(50) NOT NULL,
    llm_model VARCHAR(50) NOT NULL,
    tokens INTEGER NOT NULL,
    summary_prompt TEXT NOT NULL,
    additional_prompt TEXT,
    file_path VARCHAR(512) NOT NULL,
    file_name VARCHAR(100) NOT NULL,
    duration INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    result_json JSONB,
    stage_entered_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    started_at TIMESTAMP,
    competed_at TIMESTAMP
);

CREATE INDEX idx_tasks_group_id ON tasks(group_id);
CREATE INDEX idx_tasks_status ON tasks(status);