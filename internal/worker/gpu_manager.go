package worker

type GPUManager struct {
    available chan struct{}
}

// инициализация менеджера ресурсов, заполняем структурами
func NewGPUManager(totalVRAMGB int) *GPUManager {
    ch := make(chan struct{}, totalVRAMGB)
    for i := 0; i < totalVRAMGB; i++ {
        ch <- struct{}{}
    }
    return &GPUManager{available: ch}
}

// вычитываем значения из канала, как бы занимая память, если будет занято больше положенного, горутина заблокируется и будет ждать, пока память освободится, потому что читать с канала будет нечего
func (m *GPUManager) Acquire(needGB int) {
    for i := 0; i < needGB; i++ {
        <-m.available
    }
}

// кладем в канал значения, чтобы мы снова могли их читать и занимать память
func (m *GPUManager) Release(needGB int) {
    for i := 0; i < needGB; i++ {
        m.available <- struct{}{}
    }
}