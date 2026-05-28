import os
import uuid
import json
import requests
import torch
import sys
import types
import soundfile as sf
import numpy as np
from datetime import datetime
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from contextlib import asynccontextmanager

# ---------- Патчи для torchaudio (как в исходном скрипте) ----------
if not hasattr(torchaudio, 'set_audio_backend'):
    torchaudio.set_audio_backend = lambda x: None
if not hasattr(torchaudio, 'get_audio_backend'):
    torchaudio.get_audio_backend = lambda: "soundfile"

if 'torchaudio.backend' not in sys.modules:
    mock_backend = types.ModuleType('torchaudio.backend')
    mock_common = types.ModuleType('torchaudio.backend.common')
    mock_common.AudioMetaData = type('AudioMetaData', (), {})
    mock_backend.common = mock_common
    sys.modules['torchaudio.backend'] = mock_backend
    sys.modules['torchaudio.backend.common'] = mock_common
    torchaudio.backend = mock_backend

original_torch_load = torch.load
def patched_torch_load(*args, **kwargs):
    kwargs['weights_only'] = False
    return original_torch_load(*args, **kwargs)
torch.load = patched_torch_load

def direct_soundfile_load(filepath, frame_offset=0, num_frames=-1, *args, **kwargs):
    if isinstance(filepath, dict):
        filepath = filepath.get("audio", filepath)
    audio_data, sample_rate = sf.read(
        filepath, dtype='float32', start=frame_offset, frames=num_frames
    )
    if audio_data.ndim == 1:
        audio_data = audio_data.reshape(1, -1)
    else:
        audio_data = audio_data.T
    return torch.from_numpy(audio_data), sample_rate

torchaudio.load = direct_soundfile_load

class RealAudioMetaData:
    def __init__(self, num_frames, sample_rate, num_channels):
        self.num_frames = num_frames
        self.sample_rate = sample_rate
        self.num_channels = num_channels
        self.bits_per_sample = 16
        self.encoding = "PCM_S"

def direct_soundfile_info(filepath, *args, **kwargs):
    if isinstance(filepath, dict):
        filepath = filepath.get("audio", filepath)
    info = sf.info(filepath)
    return RealAudioMetaData(
        num_frames=info.frames,
        sample_rate=info.samplerate,
        num_channels=info.channels
    )

torchaudio.info = direct_soundfile_info

# ---------- Импорт pyannote ----------
from pyannote.audio import Pipeline

# ---------- Глобальная переменная для пайплайна ----------
diarization_pipeline = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global diarization_pipeline
    hf_token = os.environ.get("HF_TOKEN")
    if not hf_token:
        raise RuntimeError("Переменная окружения HF_TOKEN не задана!")
    print("[INFO] Загрузка модели диаризации...")
    diarization_pipeline = Pipeline.from_pretrained(
        "pyannote/speaker-diarization-3.1",
        use_auth_token=hf_token
    )
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    diarization_pipeline.to(device)
    print(f"[INFO] Модель диаризации загружена на устройство: {device}")
    yield
    # cleanup (опционально)

app = FastAPI(lifespan=lifespan)

class DiarizeRequest(BaseModel):
    input_url: str   # Presigned GET на аудиофайл
    output_url: str  # Presigned PUT для JSON-результата

def run_diarization(input_url: str, output_url: str, task_id: str):
    # Временные пути
    local_audio = f"/tmp/{task_id}_audio.wav"
    local_json = f"/tmp/{task_id}_diarization.json"

    try:
        # 1. Скачиваем аудио
        print(f"[INFO] Скачивание аудио {input_url}")
        r = requests.get(input_url, stream=True)
        r.raise_for_status()
        with open(local_audio, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192):
                f.write(chunk)

        # 2. Диаризация
        print("[INFO] Запуск диаризации...")
        diarization = diarization_pipeline(local_audio)

        # 3. Склейка соседних сегментов одного спикера (как в исходном скрипте)
        raw_results = []
        for turn, _, speaker in diarization.itertracks(yield_label=True):
            raw_results.append({
                "start": round(turn.start, 2),
                "end": round(turn.end, 2),
                "speaker": speaker
            })

        MAX_GAP = 1.5
        merged_results = []
        for segment in raw_results:
            if not merged_results:
                merged_results.append(segment)
            else:
                last_segment = merged_results[-1]
                gap = segment["start"] - last_segment["end"]
                if segment["speaker"] == last_segment["speaker"] and gap <= MAX_GAP:
                    last_segment["end"] = segment["end"]
                else:
                    merged_results.append(segment)

        # 4. Сохраняем результат в JSON и загружаем обратно
        with open(local_json, 'w', encoding='utf-8') as f:
            json.dump(merged_results, f, ensure_ascii=False, indent=2)

        print(f"[INFO] Загрузка результата {output_url}")
        with open(local_json, 'rb') as fout:
            resp = requests.put(output_url, data=fout)
            resp.raise_for_status()
        print("[SUCCESS] Диаризация завершена успешно")

    finally:
        for f in [local_audio, local_json]:
            if os.path.exists(f):
                os.remove(f)

@app.post("/diarize")
async def diarize(request: DiarizeRequest):
    task_id = str(uuid.uuid4())[:8]
    try:
        run_diarization(request.input_url, request.output_url, task_id)
        return {"status": "ok"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))