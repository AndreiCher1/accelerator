package service

import (
	"accelerator/internal/core/config"
	"accelerator/internal/core/storage"
	"accelerator/internal/features/tasks/repository"
)

type TasksService struct {
	repo  *repository.TasksRepo
	minio *storage.MinIOClient
	cfg   *config.Config
}

func NewTasksService(repo *repository.TasksRepo, minio *storage.MinIOClient, cfg *config.Config) *TasksService {
	return &TasksService{
		repo:  repo,
		minio: minio,
		cfg:   cfg,
	}
}
