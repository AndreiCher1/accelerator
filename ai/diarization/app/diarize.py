import argparse
import os
import json
import torch
import sys
import types
import soundfile as sf
import numpy as np
from datetime import datetime


import torchaudio

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
        filepath, 
        dtype='float32', 
        start=frame_offset, 
        frames=num_frames
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

from pyannote.audio import Pipeline

def run_diarization(audio_path, auth_token):
    start_time = datetime.now()
    print(f"\n[{start_time.strftime('%H:%M:%S')}] === СТАРТ ОБРАБОТКИ ===")
    
    base_name = os.path.splitext(audio_path)[0]
    output_json_path = f"{base_name}.json"

    print("Загружаем модель диаризации...")
    pipeline = Pipeline.from_pretrained(
        "pyannote/speaker-diarization-3.1",
        use_auth_token=auth_token
    )

    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    pipeline.to(device)
    print(f"Модель загружена на устройство: {device}")

    print(f"Анализируем файл: {audio_path}")
    diarization = pipeline(audio_path)

    raw_results = []
    print("Обработка завершена. Склеиваем фрагменты...")
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

    with open(output_json_path, 'w', encoding='utf-8') as f:
        json.dump(merged_results, f, ensure_ascii=False, indent=4)
    
    end_time = datetime.now()
    execution_time = end_time - start_time
    
    print(f"\n[{end_time.strftime('%H:%M:%S')}] === ФИНИШ ===")
    print(f"Результат сохранен в: {output_json_path}")
    print(f"Было сегментов: {len(raw_results)}. Стало после склейки: {len(merged_results)}.")
    print(f"Затрачено времени на всю работу: {execution_time}\n")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Анализ аудио и создание JSON")
    parser.add_argument("input_audio", help="Путь к входному аудиофайлу")
    args = parser.parse_args()

    hf_token = os.environ.get("HF_TOKEN")
    if not hf_token:
        print("Ошибка: Не задана переменная окружения HF_TOKEN!")
        exit(1)
        
    run_diarization(args.input_audio, hf_token)