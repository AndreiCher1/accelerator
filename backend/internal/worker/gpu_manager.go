package worker

import (
	"accelerator/internal/core/config"
	"context"
)

type ResourceManager struct {
	vram chan struct{}
	gpu  chan struct{}
	cpu  chan struct{}
	ram  chan struct{}
}

// инициализация менеджера ресурсов, заполняем структурами
func NewResourceManager(totalVRAM, totalRAM int) *ResourceManager {
	m := &ResourceManager{
		vram: make(chan struct{}, totalVRAM),
		gpu:  make(chan struct{}, 100),
		cpu:  make(chan struct{}, 100),
		ram:  make(chan struct{}, totalRAM),
	}
	// заполняем каналы (для vram, gpu, cpu — как раньше, для ram — аналогично)
	for i := 0; i < totalVRAM; i++ {
		m.vram <- struct{}{}
	}
	for i := 0; i < 100; i++ {
		m.gpu <- struct{}{}
	}
	for i := 0; i < 100; i++ {
		m.cpu <- struct{}{}
	}
	for i := 0; i < totalRAM; i++ {
		m.ram <- struct{}{}
	}
	return m
}

// вычитываем значения из канала, как бы занимая память, если будет занято больше положенного, горутина заблокируется и будет ждать, пока память освободится, потому что читать с канала будет нечего
// пытается занять все ресурсы квоты, но может быть прервана контекстом для нормального завершения приложения, чтобы горутина не оставалась вечно висеть с занятыми ресурсами
func (m *ResourceManager) AcquireWithContext(ctx context.Context, q config.ResourceQuota) error {
	for i := 0; i < q.VRAMGB; i++ {
		select {
		case <-m.vram:
		case <-ctx.Done():
			// Возвращаем то, что уже заняли, чтобы не потерять слоты
			for j := 0; j < i; j++ {
				m.vram <- struct{}{}
			}
			return ctx.Err()
		}
	}
	// Аналогично для gpu, cpu, ram
	for i := 0; i < q.GPUPercent; i++ {
		select {
		case <-m.gpu:
		case <-ctx.Done():
			// откат уже занятых vram и части gpu
			for j := 0; j < q.VRAMGB; j++ {
				m.vram <- struct{}{}
			}
			for j := 0; j < i; j++ {
				m.gpu <- struct{}{}
			}
			return ctx.Err()
		}
	}
	for i := 0; i < q.CPUPercent; i++ {
		select {
		case <-m.cpu:
		case <-ctx.Done():
			// откат vram, gpu и части cpu
			for j := 0; j < q.VRAMGB; j++ {
				m.vram <- struct{}{}
			}
			for j := 0; j < q.GPUPercent; j++ {
				m.gpu <- struct{}{}
			}
			for j := 0; j < i; j++ {
				m.cpu <- struct{}{}
			}
			return ctx.Err()
		}
	}
	for i := 0; i < q.RAMGB; i++ {
		select {
		case <-m.ram:
		case <-ctx.Done():
			for j := 0; j < q.VRAMGB; j++ {
				m.vram <- struct{}{}
			}
			for j := 0; j < q.GPUPercent; j++ {
				m.gpu <- struct{}{}
			}
			for j := 0; j < q.CPUPercent; j++ {
				m.cpu <- struct{}{}
			}
			for j := 0; j < i; j++ {
				m.ram <- struct{}{}
			}
			return ctx.Err()
		}
	}
	return nil
}

// кладем в канал значения, чтобы мы снова могли их читать и занимать память
func (m *ResourceManager) Release(q config.ResourceQuota) {
	for i := 0; i < q.VRAMGB; i++ {
		m.vram <- struct{}{}
	}
	for i := 0; i < q.GPUPercent; i++ {
		m.gpu <- struct{}{}
	}
	for i := 0; i < q.CPUPercent; i++ {
		m.cpu <- struct{}{}
	}
	for i := 0; i < q.RAMGB; i++ {
		m.ram <- struct{}{}
	}
}
