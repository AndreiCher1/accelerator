import os
import sys
import json
import uuid
import warnings
import requests
import torch
from pydub import AudioSegment
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

warnings.filterwarnings("ignore")

# ---------- Переменные окружения для кеша моделей ----------
os.environ.setdefault("HF_HOME", "./models")
os.environ.setdefault("MODELSCOPE_CACHE", "./models")

# ---------- Импорт модели ----------
from qwen_asr import Qwen3ASRModel

MODEL_ID = "Qwen/Qwen3-ASR-1.7B"

# ---------- Глобальная модель ----------
asr_model = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global asr_model
    kwargs = {"device_map": "cuda"} if torch.cuda.is_available() else {"device_map": "cpu"}
    print(f"[INFO] Загрузка модели ASR {MODEL_ID}...")
    asr_model = Qwen3ASRModel.from_pretrained(MODEL_ID, **kwargs)

    # Настройка greedy decoding
    if hasattr(asr_model, "model") and hasattr(asr_model.model, "generation_config"):
        asr_model.model.generation_config.temperature = 0.2
        asr_model.model.generation_config.do_sample = False
        asr_model.model.generation_config.repetition_penalty = 1.2
        print("[INFO] Greedy search + penalty настроены")
    else:
        print("[WARN] Не удалось настроить generation_config")

    print("[INFO] Модель готова к транскрипции")
    yield

app = FastAPI(lifespan=lifespan)

# ---------- Тело запроса ----------
class TranscribeRequest(BaseModel):
    input_url: str         # Presigned GET на JSON диаризации
    denoised_url: str      # Presigned GET на аудио (denoised.wav)
    output_url: str        # Presigned PUT для результата (JSON с текстом)

# ---------- Обработчик ----------
@app.post("/transcribe")
async def transcribe(req: TranscribeRequest):
    task_id = str(uuid.uuid4())[:8]

    # Временные файлы
    audio_path = f"/tmp/{task_id}_audio.wav"
    diar_json_path = f"/tmp/{task_id}_diar.json"
    result_path = f"/tmp/{task_id}_result.json"
    temp_chunk = f"/tmp/{task_id}_chunk.wav"

    try:
        # 1. Скачать аудио
        print(f"[INFO] Скачивание аудио {req.denoised_url}")
        r = requests.get(req.denoised_url, stream=True)
        r.raise_for_status()
        with open(audio_path, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192):
                f.write(chunk)

        # 2. Скачать JSON диаризации
        print(f"[INFO] Скачивание диаризации {req.input_url}")
        r = requests.get(req.input_url)
        r.raise_for_status()
        with open(diar_json_path, 'wb') as f:
            f.write(r.content)

        # 3. Загрузка сегментов
        with open(diar_json_path, 'r', encoding='utf-8') as f:
            segments = json.load(f)

        # 4. Обработка
        full_audio = AudioSegment.from_file(audio_path)

        if torch.cuda.is_available():
            torch.cuda.reset_peak_memory_stats()

        print(f"[INFO] Транскрибация {len(segments)} сегментов...")

        for i, segment in enumerate(segments):
            start_ms = int(segment["start"] * 1000)
            end_ms = int(segment["end"] * 1000)

            if end_ms <= start_ms:
                segment["text"] = ""
                continue

            chunk = full_audio[start_ms:end_ms]
            duration_ms = len(chunk)

            # Увеличиваем короткие чанки до 500 мс
            if duration_ms < 500:
                chunk = chunk + AudioSegment.silent(
                    duration=500 - duration_ms,
                    frame_rate=chunk.frame_rate
                )

            full_text = ""
            MAX_CHUNK_MS = 30000  # 30 секунд
            for offset_ms in range(0, max(len(chunk), 1), MAX_CHUNK_MS):
                sub_chunk = chunk[offset_ms : offset_ms + MAX_CHUNK_MS]
                sub_chunk.export(temp_chunk, format="wav")

                try:
                    results = asr_model.transcribe(audio=temp_chunk)
                    full_text += results[0].text + " "
                except Exception as e:
                    print(f"\n[ERROR] Ошибка на сегменте {i}: {e}")

            segment["text"] = full_text.strip()

            vram_mb = torch.cuda.max_memory_allocated() / (1024 * 1024) if torch.cuda.is_available() else 0
            sys.stdout.write(f"\rОбработано: {i+1}/{len(segments)} | Макс VRAM: {vram_mb:.0f} MB   ")
            sys.stdout.flush()

        # 5. Сохранение и загрузка результата
        with open(result_path, 'w', encoding='utf-8') as f:
            json.dump(segments, f, ensure_ascii=False, indent=2)

        print(f"\n[INFO] Загрузка результата {req.output_url}")
        with open(result_path, 'rb') as fout:
            resp = requests.put(req.output_url, data=fout)
            resp.raise_for_status()

        print("[SUCCESS] Транскрипция завершена успешно")
        return {"status": "ok"}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

    finally:
        # Удаляем временные файлы
        for f in [audio_path, diar_json_path, result_path, temp_chunk]:
            if os.path.exists(f):
                os.remove(f)