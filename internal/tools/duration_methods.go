package tools

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// ============================================== MP3 ==================================================
// GetDurationFromMP3 вычисляет длительность MP3-файла в секундах,
// используя потоковый разбор фреймов без декодирования аудиоданных.
// Принимает io.ReadSeeker, что позволяет эффективно работать с *os.File.

func GetDurationFromMP3(r io.Reader) (int, error) {
	br := bufio.NewReader(r)

	// 1. Пропускаем ID3v2 (если есть)
	if err := skipID3v2(br); err != nil {
		return 0, fmt.Errorf("ошибка пропуска ID3v2: %w", err)
	}

	// 2. Ищем первый валидный MPEG-фрейм с ненулевым битрейтом и частотой
	header, err := findFirstValidMPEGFrame(br)
	if err != nil {
		return 0, fmt.Errorf("не найден валидный MP3 фрейм: %w", err)
	}

	version := (header >> 19) & 0x3
	sampleRateIdx := (header >> 10) & 0x3
	sampleRate := sampleRates[version][sampleRateIdx]
	samplesPerFrame := int64(1152)
	if version != 3 {
		samplesPerFrame = 576
	}

	bitrateIdx := (header >> 12) & 0xF
	padding := (header >> 9) & 0x1
	brValue := bitrates[version][1][bitrateIdx] * 1000
	frameLen := int64(144*brValue)/int64(sampleRate) + int64(padding)

	// Читаем оставшуюся часть первого фрейма (уже считали 4 байта заголовка)
	firstFrameData := make([]byte, frameLen-4)
	if _, err := io.ReadFull(br, firstFrameData); err != nil {
		return 0, fmt.Errorf("ошибка чтения первого фрейма: %w", err)
	}

	// 3. Пытаемся извлечь общее число фреймов из Xing/Info (VBR)
	if totalFrames := parseXing(firstFrameData); totalFrames > 0 {
		dur := float64(totalFrames) * float64(samplesPerFrame) / float64(sampleRate)
		return int(math.Ceil(dur)), nil
	}

	// 4. Xing отсутствует → перебираем все последующие фреймы
	totalSamples := samplesPerFrame // первый фрейм уже учтён
	for {
		hdr, err := readFrameHeader(br)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("ошибка чтения заголовка фрейма: %w", err)
		}

		if !isValidMPEGHeader(hdr) {
			// Пропускаем 1 байт и ищем следующий фрейм
			if _, discardErr := br.ReadByte(); discardErr != nil {
				if discardErr == io.EOF {
					break
				}
				return 0, fmt.Errorf("ошибка при восстановлении синхронизации: %w", discardErr)
			}
			continue
		}

		ver := (hdr >> 19) & 0x3
		brIdx := (hdr >> 12) & 0xF
		srIdx := (hdr >> 10) & 0x3
		pad := (hdr >> 9) & 0x1
		bitrate := bitrates[ver][1][brIdx] * 1000
		srate := sampleRates[ver][srIdx]

		if bitrate == 0 || srate == 0 {
			// Некорректный фрейм — пропускаем байт
			if _, discardErr := br.ReadByte(); discardErr != nil {
				if discardErr == io.EOF {
					break
				}
				return 0, discardErr
			}
			continue
		}

		flen := int64(144*bitrate)/int64(srate) + int64(pad)
		if _, err := io.CopyN(io.Discard, br, flen-4); err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("ошибка пропуска данных фрейма: %w", err)
		}

		s := int64(1152)
		if ver != 3 {
			s = 576
		}
		totalSamples += s
	}

	if totalSamples == 0 {
		return 0, fmt.Errorf("не найдено ни одного фрейма")
	}
	dur := float64(totalSamples) / float64(sampleRate)
	return int(math.Ceil(dur)), nil
}

// ----------------- вспомогательные функции -----------------

func skipID3v2(br *bufio.Reader) error {
	peek, err := br.Peek(10)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	if string(peek[:3]) != "ID3" {
		return nil
	}
	if _, err := io.CopyN(io.Discard, br, 10); err != nil {
		return err
	}
	size := int64(peek[6])<<21 | int64(peek[7])<<14 | int64(peek[8])<<7 | int64(peek[9])
	if _, err := io.CopyN(io.Discard, br, size); err != nil {
		return err
	}
	return nil
}

// findFirstValidMPEGFrame ищет первый фрейм, у которого битрейт и частота не равны 0
func findFirstValidMPEGFrame(br *bufio.Reader) (uint32, error) {
	for {
		b, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		if b == 0xFF {
			next, err := br.Peek(1)
			if err != nil {
				return 0, err
			}
			if (next[0] & 0xE0) == 0xE0 {
				headerBytes := make([]byte, 4)
				headerBytes[0] = 0xFF
				headerBytes[1] = next[0]
				if _, err := io.ReadFull(br, headerBytes[2:4]); err != nil {
					return 0, err
				}
				header := binary.BigEndian.Uint32(headerBytes)
				if isValidMPEGHeader(header) {
					// Дополнительно проверяем, что битрейт и частота не нулевые
					version := (header >> 19) & 0x3
					bitrateIdx := (header >> 12) & 0xF
					sampleRateIdx := (header >> 10) & 0x3
					if bitrates[version][1][bitrateIdx] != 0 && sampleRates[version][sampleRateIdx] != 0 {
						return header, nil
					}
				}
			}
			// Пропускаем следующий байт, если синхрослово не подтвердилось
			if _, err := br.ReadByte(); err != nil {
				return 0, err
			}
		}
	}
}

func readFrameHeader(br *bufio.Reader) (uint32, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(br, buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func isValidMPEGHeader(h uint32) bool {
	if (h>>21) != 0x7FF {
		return false
	}
	layer := (h >> 17) & 0x3
	if layer != 1 { // только Layer 3
		return false
	}
	bitrateIdx := (h >> 12) & 0xF
	if bitrateIdx == 0 || bitrateIdx == 15 {
		return false
	}
	sampleRateIdx := (h >> 10) & 0x3
	if sampleRateIdx == 3 {
		return false
	}
	return true
}

func parseXing(data []byte) int64 {
	idx := bytes.Index(data, []byte("Xing"))
	if idx == -1 {
		idx = bytes.Index(data, []byte("Info"))
	}
	if idx == -1 || idx+8 > len(data) {
		return 0
	}
	flags := binary.BigEndian.Uint32(data[idx+4:])
	if flags&0x1 == 0 {
		return 0
	}
	return int64(binary.BigEndian.Uint32(data[idx+8:]))
}

var bitrates [4][4][16]int
var sampleRates [4][4]int

func init() {
	// MPEG1 Layer 3
	bitrates[3][1] = [16]int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0}
	// MPEG2/2.5 Layer 3
	bitrates[2][1] = [16]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	bitrates[0][1] = [16]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}

	sampleRates[3] = [4]int{44100, 48000, 32000, 0}
	sampleRates[2] = [4]int{22050, 24000, 16000, 0}
	sampleRates[0] = [4]int{11025, 12000, 8000, 0}
}

// ============================================== WAV ==================================================

// пиздец как будто на ассемблере пишу, что за говно
// вычисляет длительность WAV-файла в секундах
// Возвращает длительность и возможную ошибку
func GetDurationFromWAV(r io.Reader) (int, error) {
	// 	Чтобы вычислить общую длительность в секундах, нужны всего четыре параметра из заголовка fmt:
	// SampleRate (uint32) — частота дискретизации (количество отсчётов в секунду).
	// NumChannels (uint16) — количество аудиоканалов (1 для моно, 2 для стерео).
	// BitsPerSample (uint16) — битность одного отсчёта (например, 16 бит).
	// DataSize (uint32) — размер блока с данными в байтах.

	// 1. Пропускаем RIFF-заголовок (12 байт):
	//    ChunkID (4 байта: "RIFF") -> пропускаем
	//    ChunkSize (4 байта) -> пропускаем
	//    Format (4 байта: "WAVE") -> пропускаем
	_, err := io.CopyN(io.Discard, r, 12)
	if err != nil {
		return 0, fmt.Errorf("failed to skip RIFF header: %w", err)
	}

	// 2. Ищем подблок "fmt ".
	for {
		// Читаем 4 байта — идентификатор подблока.
		var chunkID [4]byte
		if _, err := io.ReadFull(r, chunkID[:]); err != nil {
			return 0, fmt.Errorf("failed to read chunk ID: %w", err)
		}

		// Читаем 4 байта — размер подблока.
		var chunkSize uint32
		if err := binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
			return 0, fmt.Errorf("failed to read chunk size: %w", err)
		}

		// Если нашли "fmt ", останавливаемся.
		if string(chunkID[:]) == "fmt " {
			break
		}
		// Пропускаем содержимое чужих блоков.
		if _, err := io.CopyN(io.Discard, r, int64(chunkSize)); err != nil {
			return 0, fmt.Errorf("failed to skip unknown chunk: %w", err)
		}
	}

	// 3. Парсим заголовок "fmt " (минимум 16 байт).
	var audioFormat, numChannels, bitsPerSample uint16
	var sampleRate, byteRate uint32
	var blockAlign uint16

	if err := binary.Read(r, binary.LittleEndian, &audioFormat); err != nil {
		return 0, fmt.Errorf("failed to read audio format: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &numChannels); err != nil {
		return 0, fmt.Errorf("failed to read number of channels: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &sampleRate); err != nil {
		return 0, fmt.Errorf("failed to read sample rate: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &byteRate); err != nil {
		return 0, fmt.Errorf("failed to read byte rate: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &blockAlign); err != nil {
		return 0, fmt.Errorf("failed to read block align: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &bitsPerSample); err != nil {
		return 0, fmt.Errorf("failed to read bits per sample: %w", err)
	}

	_ = audioFormat // Можно использовать для валидации (1 = PCM).
	_ = byteRate    // Может понадобиться для других целей.

	// 4. Ищем подблок "data".
	for {
		// Читаем 4 байта — идентификатор подблока.
		var chunkID [4]byte
		if _, err := io.ReadFull(r, chunkID[:]); err != nil {
			return 0, fmt.Errorf("failed to read chunk ID: %w", err)
		}

		if string(chunkID[:]) == "data" {
			var dataSize uint32
			if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
				return 0, fmt.Errorf("failed to read data chunk size: %w", err)
			}

			// 5. Вычисляем длительность.
			bytesPerSample := int64(bitsPerSample) / 8
			totalSamples := int64(dataSize) / (int64(numChannels) * bytesPerSample)
			duration := float64(totalSamples) / float64(sampleRate)
			return int(math.Ceil(duration)), nil
		}

		// Пропускаем содержимое чужих блоков.
		var chunkSize uint32
		if err := binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
			return 0, fmt.Errorf("failed to read chunk size: %w", err)
		}
		if _, err := io.CopyN(io.Discard, r, int64(chunkSize)); err != nil {
			return 0, fmt.Errorf("failed to skip chunk: %w", err)
		}
	}
}

// readLittleEndianInt64 читает 8-байтовое целое число в формате Little Endian.
// используется для чтения granule position из заголовка OggS.
func readLittleEndianInt64(data []byte) (int64, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("недостаточно данных для чтения 8-байтового целого")
	}
	var value int64
	err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &value)
	return value, err
}

// readLittleEndianInt32 читает 4-байтовое целое число в формате Little Endian.
// используется для чтения sample rate из заголовка Vorbis.
func readLittleEndianInt32(data []byte) (int32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("недостаточно данных для чтения 4-байтового целого")
	}
	var value int32
	err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &value)
	return value, err
}

// ============================================== OGG/Opus ==================================================

// GetDurationFromOGG вычисляет длительность OGG Vorbis/Opus файла, читая поток последовательно.
// Не загружает весь файл в память – подходит для файлов любого размера.
func GetDurationFromOGG(r io.Reader) (int, error) {
	var sampleRate int32
	foundVorbis := false
	var lastGranule int64
	haveGranule := false

	// буфер для минимального заголовка OGG-страницы (27 байт)
	headerBuf := make([]byte, 27)

	for {
		// читаем заголовок страницы
		_, err := io.ReadFull(r, headerBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("ошибка чтения OGG заголовка: %w", err)
		}

		// проверяем сигнатуру "OggS"
		if string(headerBuf[:4]) != "OggS" {
			return 0, fmt.Errorf("ожидалась OGG-страница")
		}
		version := headerBuf[4]
		if version != 0 {
			return 0, fmt.Errorf("неподдерживаемая версия OGG: %d", version)
		}

		headerType := headerBuf[5]
		granulePos := int64(binary.LittleEndian.Uint64(headerBuf[6:14]))
		pageSegments := int(headerBuf[26])

		// читаем таблицу сегментов
		segTable := make([]byte, pageSegments)
		if _, err := io.ReadFull(r, segTable); err != nil {
			return 0, fmt.Errorf("ошибка чтения таблицы сегментов: %w", err)
		}

		// вычисляем общий размер данных страницы
		var dataSize int64
		for _, segLen := range segTable {
			dataSize += int64(segLen)
		}

		// если это первая страница (BOS) и мы ещё не определили sample rate – пробуем извлечь его из Vorbis
		if (headerType&0x02) != 0 && !foundVorbis {
			if dataSize > 0 {
				// читаем начало данных, чтобы проверить кодек
				probeSize := dataSize
				if probeSize > 64 {
					probeSize = 64
				}
				probe := make([]byte, probeSize)
				if _, err := io.ReadFull(r, probe); err != nil {
					return 0, fmt.Errorf("ошибка чтения первой страницы: %w", err)
				}
				dataSize -= int64(len(probe))

				// идентификационный пакет Vorbis: packet_type (1 байт = 1), затем "vorbis"
				if len(probe) >= 1 && probe[0] == 0x01 {
					if len(probe) >= 7 && string(probe[1:7]) == "vorbis" {
						// после packet_type + "vorbis" (7 байт) + 4 байта версии + 1 байт каналов -> 12 байт от начала,
						// далее 4 байта sample rate (LE)
						if len(probe) >= 12+4 {
							sampleRate = int32(binary.LittleEndian.Uint32(probe[12:16]))
							foundVorbis = true
						}
					}
				}
				// дочитываем оставшиеся данные этой страницы
				if dataSize > 0 {
					if _, err := io.CopyN(io.Discard, r, dataSize); err != nil {
						return 0, fmt.Errorf("ошибка пропуска данных страницы: %w", err)
					}
				}
			}
		} else {
			// не первая страница или sample rate уже известен – просто пропускаем данные
			if dataSize > 0 {
				if _, err := io.CopyN(io.Discard, r, dataSize); err != nil {
					return 0, fmt.Errorf("ошибка пропуска данных страницы: %w", err)
				}
			}
		}

		// обновляем последнюю валидную granule позицию
		if granulePos > 0 {
			lastGranule = granulePos
			haveGranule = true
		}
	}

	if !haveGranule {
		return 0, fmt.Errorf("не найдена granule position в OGG файле")
	}
	if !foundVorbis {
		sampleRate = 48000 // Opus (или неизвестный кодек) по умолчанию 48000 Гц
	}
	if sampleRate <= 0 {
		return 0, fmt.Errorf("некорректная частота дискретизации: %d", sampleRate)
	}

	duration := float64(lastGranule) / float64(sampleRate)
	return int(math.Ceil(duration)), nil
}

// ============================================== AAC ==================================================
func GetDurationFromAAC(reader io.Reader) (int, error) {
	br := bufio.NewReader(reader)

	var totalSamples int64
	var sampleRate int
	firstFrame := true

	for {
		sr, _, err := readADTSFrame(br, firstFrame)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("ошибка чтения AAC фрейма: %w", err)
		}
		if firstFrame {
			sampleRate = sr
			firstFrame = false
		}
		totalSamples += 1024
	}

	if sampleRate == 0 || totalSamples == 0 {
		return 0, fmt.Errorf("не удалось определить длительность AAC")
	}

	duration := float64(totalSamples) / float64(sampleRate)
	return int(math.Ceil(duration)), nil
}

// readADTSFrame находит следующий ADTS-фрейм, читает его заголовок,
// пропускает аудиоданные и возвращает:
// - sampleRate (только для firstFrame)
// - длину фрейма
// - ошибку (io.EOF при конце файла)
func readADTSFrame(br *bufio.Reader, firstFrame bool) (int, int, error) {
	// Поиск синхрослова 0xFFF без использования UnreadByte/Peek
	var header [7]byte
	for {
		// Читаем первый байт синхрослова
		b, err := br.ReadByte()
		if err != nil {
			return 0, 0, err
		}
		if b != 0xFF {
			continue
		}

		// Читаем второй байт
		next, err := br.ReadByte()
		if err != nil {
			return 0, 0, err
		}
		// Проверяем: старшие 4 бита = 0xF, layer (биты 1,2) == 0
		if (next&0xF0) == 0xF0 && (next&0x06) == 0 {
			// Нашли синхрослово, сохраняем два первых байта заголовка
			header[0] = 0xFF
			header[1] = next
			// Дочитываем оставшиеся 5 байт заголовка
			if _, err := io.ReadFull(br, header[2:7]); err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return 0, 0, io.EOF
				}
				return 0, 0, err
			}
			break
		}
		// Иначе продолжаем поиск с текущей позиции (после next)
	}

	// Проверяем заголовок
	if header[0] != 0xFF || (header[1]&0xF0) != 0xF0 || (header[1]&0x06) != 0 {
		return 0, 0, fmt.Errorf("не ADTS фрейм")
	}

	protectionAbsent := (header[1] & 0x01) == 1
	samplingFreqIndex := (header[2] >> 2) & 0x0F
	frameLength := (uint16(header[3])&0x03)<<11 | uint16(header[4])<<3 | uint16(header[5])>>5

	if samplingFreqIndex == 15 {
		return 0, 0, fmt.Errorf("невалидный индекс частоты дискретизации")
	}
	if frameLength < 7 {
		return 0, 0, fmt.Errorf("длина фрейма < 7")
	}

	// Определяем размер заголовка (7 или 9 байт)
	headerSize := 7
	if !protectionAbsent {
		crc := make([]byte, 2)
		if _, err := io.ReadFull(br, crc); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return 0, 0, io.EOF
			}
			return 0, 0, err
		}
		headerSize = 9
	}

	// Пропускаем аудиоданные
	skipSize := int64(frameLength) - int64(headerSize)
	if skipSize > 0 {
		if _, err := io.CopyN(io.Discard, br, skipSize); err != nil {
			if err == io.EOF {
				return 0, 0, io.EOF
			}
			return 0, 0, err
		}
	}

	// Частота дискретизации для первого фрейма
	var sr int
	if firstFrame {
		sr = aacSampleRate(samplingFreqIndex)
	}
	return sr, int(frameLength), nil
}

func aacSampleRate(idx byte) int {
	switch idx {
	case 0: return 96000
	case 1: return 88200
	case 2: return 64000
	case 3: return 48000
	case 4: return 44100
	case 5: return 32000
	case 6: return 24000
	case 7: return 22050
	case 8: return 16000
	case 9: return 12000
	case 10: return 11025
	case 11: return 8000
	case 12: return 7350
	default: return 44100
	}
}

// ============================================== FLAC ==================================================

func GetDurationFromFLAC(reader io.Reader) (int, error) {
	// Читаем заголовок "fLaC"
	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return 0, fmt.Errorf("ошибка чтения FLAC заголовка: %w", err)
	}
	if string(header) != "fLaC" {
		return 0, fmt.Errorf("некорректный FLAC файл: заголовок не fLaC")
	}

	// Читаем первый метаданных-блок (всегда StreamInfo)
	metaHeader := make([]byte, 4)
	if _, err := io.ReadFull(reader, metaHeader); err != nil {
		return 0, fmt.Errorf("ошибка чтения заголовка метаданных: %w", err)
	}

	blockType := metaHeader[0] & 0x7F
	if blockType != 0 {
		return 0, fmt.Errorf("первый блок не StreamInfo")
	}

	blockLength := int(metaHeader[1])<<16 | int(metaHeader[2])<<8 | int(metaHeader[3])
	if blockLength < 34 { // StreamInfo всегда 34 байта
		return 0, fmt.Errorf("блок StreamInfo поврежден")
	}

	// Читаем 18 байт StreamInfo — этого хватит для sample rate и total samples
	streamInfo := make([]byte, 18)
	if _, err := io.ReadFull(reader, streamInfo); err != nil {
		return 0, fmt.Errorf("ошибка чтения StreamInfo: %w", err)
	}

	// Частота дискретизации (20 бит, big‑endian)
	// Старшие 8 бит в streamInfo[10], средние 8 бит в streamInfo[11],
	// младшие 4 бита в streamInfo[12] (старшие биты байта)
	sampleRate := int(streamInfo[10])<<12 | int(streamInfo[11])<<4 | int(streamInfo[12])>>4

	// Общее количество сэмплов (36 бит)
	// Старшие 4 бита — в младших битах streamInfo[13],
	// остальные 32 бита — в streamInfo[14..17] (big‑endian)
	totalSamples := int64(streamInfo[13]&0x0F) << 32
	totalSamples |= int64(streamInfo[14]) << 24
	totalSamples |= int64(streamInfo[15]) << 16
	totalSamples |= int64(streamInfo[16]) << 8
	totalSamples |= int64(streamInfo[17])

	if sampleRate == 0 {
		return 0, fmt.Errorf("некорректная частота дискретизации")
	}

	duration := float64(totalSamples) / float64(sampleRate)
	return int(math.Ceil(duration)), nil
}
