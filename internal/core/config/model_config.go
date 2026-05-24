package config

import "accelerator/internal/domains"

type ResourceQuota struct {
	VRAMGB     int
	GPUPercent int
	CPUPercent int
	RAMGB      int // оперативная память
}

// при запуске нужно эмпирическим путем измерить загрузку комплектующих для каждой модели, чтобы конкурентность была эффективной
var StageQuotas = map[string]ResourceQuota{
	string(domains.StatusProcessingDenoise):    {VRAMGB: 5, GPUPercent: 100, CPUPercent: 20, RAMGB: 5},
	string(domains.StatusPendingDiarize):       {VRAMGB: 5, GPUPercent: 100, CPUPercent: 35, RAMGB: 2},
	string(domains.StatusProcessingTranscribe): {VRAMGB: 10, GPUPercent: 100, CPUPercent: 10, RAMGB: 4},
	string(domains.StatusProcessingSummarize):  {VRAMGB: 10, GPUPercent: 100, CPUPercent: 15, RAMGB: 2},
}
