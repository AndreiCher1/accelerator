package worker

import "accelerator/internal/core/config"

type GPUManager struct {
	vram chan struct{}
	gpu  chan struct{}
	cpu  chan struct{}
	ram  chan struct{}
}

// инициализация менеджера ресурсов, заполняем структурами
func NewGPUManager(totalVRAM, totalRAM int) *GPUManager {
	m := &GPUManager{
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
func (m *GPUManager) Acquire(q config.ResourceQuota) {
	for i := 0; i < q.VRAMGB; i++ {
		<-m.vram
	}
	for i := 0; i < q.GPUPercent; i++ {
		<-m.gpu
	}
	for i := 0; i < q.CPUPercent; i++ {
		<-m.cpu
	}
	for i := 0; i < q.RAMGB; i++ {
		<-m.ram
	}
}

// кладем в канал значения, чтобы мы снова могли их читать и занимать память
func (m *GPUManager) Release(q config.ResourceQuota) {
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
