package config

import (
	"accelerator/internal/domains"
	"math"
)

type ResourceQuota struct {
	VRAMGB     int
	GPUPercent int
	CPUPercent int
	RAMGB      int // оперативная память
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ВРЕМЕННОЕ РЕШЕНИЕ, В ДАЛЬНЕЙШЕМ НУЖНО БУДЕТ ПОЛУЧАТЬ РАЗМЕР АУДИО В ОПЕРАТИВКЕ
var cfg = LoadConfig()
var additionalAudioRAM = int(math.Ceil(float64(cfg.SizeLimitAudioMB) / 1024))

// при запуске нужно эмпирическим путем измерить загрузку комплектующих для каждой модели, чтобы конкурентность была эффективной
var StageQuotas = map[string]ResourceQuota{
	string(domains.StatusProcessingDenoise):    {VRAMGB: 5, GPUPercent: 100, CPUPercent: 20, RAMGB: 5 + additionalAudioRAM},
	string(domains.StatusPendingDiarize):       {VRAMGB: 5, GPUPercent: 100, CPUPercent: 35, RAMGB: 2 + additionalAudioRAM},
	string(domains.StatusProcessingTranscribe): {VRAMGB: 10, GPUPercent: 100, CPUPercent: 10, RAMGB: 4 + additionalAudioRAM},
	string(domains.StatusProcessingSummarize):  {VRAMGB: 10, GPUPercent: 100, CPUPercent: 15, RAMGB: 2},
}

// для примерное оценки длительности, также заполняется исходя из тестов своего железа
// сколько минут аудио орабатывается за минуту обработки
var MinutesOfAudioPerMinuteOfProcessing = map[string]int{
	string(domains.StatusProcessingDenoise):    17,
	string(domains.StatusPendingDiarize):       40,
	string(domains.StatusProcessingTranscribe): 20,
	string(domains.StatusProcessingSummarize):  30,
}
