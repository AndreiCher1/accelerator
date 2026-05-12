package service

import "accelerator/internal/features/patterns/repository"

type PatternsService struct {
	repo *repository.PatternsRepository
}

func NewPatternsService(repo *repository.PatternsRepository) *PatternsService {
	return &PatternsService{
		repo: repo,
	}
} 

