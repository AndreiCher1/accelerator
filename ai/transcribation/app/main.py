import os
import sys
import json
import torch
from pydub import AudioSegment

os.environ["HF_HOME"] = "./models"
os.environ["MODELSCOPE_CACHE"] = "./models"

from qwen_asr import Qwen3ASRModel

MODEL_ID = "Qwen/Qwen3-ASR-1.7B" 
AUDIO_FILE = "app/clean_speech.wav"
JSON_FILE = "app/clean_speech.json"
OUTPUT_FILE = "app/clean_speech_result.json"
TEMP_WAV = "app/temp_chunk.wav" 
MAX_CHUNK_MS = 30000 

def process_transcription():
    print(f" Инициализация системы...")
    kwargs = {"device_map": "cuda"} if torch.cuda.is_available() else {"device_map": "cpu"}
    
    wrapper = Qwen3ASRModel.from_pretrained(MODEL_ID, **kwargs)
    
    if hasattr(wrapper, "model") and hasattr(wrapper.model, "generation_config"):
        wrapper.model.generation_config.temperature = 0.0          # Убираем "фантазию"
        wrapper.model.generation_config.do_sample = False          # Строгий выбор токенов
        wrapper.model.generation_config.repetition_penalty = 1.2   # Штраф за зацикливание
        print(" Настройки высокого качества (greedy search + penalty) успешно вшиты в модель!")
    else:
        print(" Не удалось получить доступ к ядру, генерация будет стандартной.")

    with open(JSON_FILE, 'r', encoding='utf-8') as f:
        segments = json.load(f)

    full_audio = AudioSegment.from_file(AUDIO_FILE)
    
    if torch.cuda.is_available():
        torch.cuda.reset_peak_memory_stats()
    
    print(f" Начинаем транскрибацию {len(segments)} сегментов...")
    
    for i, segment in enumerate(segments):
        start_ms = int(segment["start"] * 1000)
        end_ms = int(segment["end"] * 1000)
        
        if end_ms <= start_ms:
            segment["text"] = ""
            continue

        chunk = full_audio[start_ms:end_ms]
        duration_ms = len(chunk)

        if duration_ms < 500:
            chunk = chunk + AudioSegment.silent(duration=500-duration_ms, frame_rate=chunk.frame_rate)

        full_text = ""
        for offset_ms in range(0, max(len(chunk), 1), MAX_CHUNK_MS):
            sub_chunk = chunk[offset_ms : offset_ms + MAX_CHUNK_MS]
            sub_chunk.export(TEMP_WAV, format="wav")
            
            try:
                results = wrapper.transcribe(audio=TEMP_WAV)
                full_text += results[0].text + " "
            except Exception as e:
                print(f"\n Ошибка на сегменте {i}: {e}")
            
        segment["text"] = full_text.strip()
        
        vram_mb = torch.cuda.max_memory_allocated() / (1024 * 1024) if torch.cuda.is_available() else 0
        sys.stdout.write(f"\rОбработано: {i+1}/{len(segments)} | Макс VRAM: {vram_mb:.0f} MB   ")
        sys.stdout.flush()

    if os.path.exists(TEMP_WAV):
        os.remove(TEMP_WAV)

    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
        json.dump(segments, f, ensure_ascii=False, indent=4)
    print("\n Готово!")

if __name__ == "__main__":
    import warnings
    warnings.filterwarnings("ignore")
    process_transcription()