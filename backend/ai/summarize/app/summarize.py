import os
import json
import uuid
import requests
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from contextlib import asynccontextmanager
from llama_cpp import Llama

# ---------- Вспомогательные функции ----------
def prepare_transcript(json_filepath):
    """Склеивает реплики одного спикера из сырого JSON стенограммы."""
    with open(json_filepath, 'r', encoding='utf-8') as f:
        data = json.load(f)

    transcript_lines = []
    current_speaker = None
    current_text = []

    for entry in data:
        speaker = entry.get("speaker", "UNKNOWN")
        text = entry.get("text", "").strip()

        if not text or "[ОШИБКА ПОДКЛЮЧЕНИЯ" in text:
            continue

        if speaker == current_speaker:
            current_text.append(text)
        else:
            if current_speaker is not None:
                transcript_lines.append(f"{current_speaker}: {' '.join(current_text)}")
            current_speaker = speaker
            current_text = [text]

    if current_speaker is not None:
        transcript_lines.append(f"{current_speaker}: {' '.join(current_text)}")

    return "\n".join(transcript_lines)

# ---------- Глобальная модель ----------
llm = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global llm
    model_path = os.environ.get("LOCAL_MODEL_PATH")
    if not model_path or not os.path.exists(model_path):
        raise RuntimeError("Модель не найдена. Укажите LOCAL_MODEL_PATH")

    print(f"[INFO] Загрузка LLM из {model_path} в VRAM...")
    llm = Llama(
        model_path=model_path,
        n_ctx=16384,
        n_gpu_layers=-1,      # все слои на GPU
        flash_attn=True,
        chat_format="chatml",
        verbose=False
    )
    print("[INFO] Модель готова к работе.")
    yield
    # cleanup

app = FastAPI(lifespan=lifespan)

# ---------- Модель запроса ----------
class SummarizeRequest(BaseModel):
    input_url: str   # Presigned GET на JSON стенограммы
    output_url: str  # Presigned PUT для итогового JSON-отчёта
    prompt: str      # Системный промпт (инструкция для LLM)

def run_summarization(input_url: str, output_url: str, prompt: str, task_id: str):
    local_input = f"/tmp/{task_id}_transcript.json"
    local_output = f"/tmp/{task_id}_summary.json"

    try:
        # 1. Скачиваем стенограмму
        print(f"[INFO] Скачивание стенограммы {input_url}")
        r = requests.get(input_url, stream=True)
        r.raise_for_status()
        with open(local_input, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192):
                f.write(chunk)

        # 2. Подготовка текста
        transcript_text = prepare_transcript(local_input)
        print("[INFO] Текст подготовлен, запрос к LLM...")

        # 3. Инференс с переданным промптом
        output = llm.create_chat_completion(
            messages=[
                {"role": "system", "content": prompt},
                {"role": "user", "content": f"Транскрипция:\n{transcript_text}"}
            ],
            max_tokens=4096,
            temperature=0.2
        )
        report_text = output['choices'][0]['message']['content']

        # 4. Сохранение результата в JSON
        result_data = {
            "status": "success",
            "analysis_report": report_text
        }
        with open(local_output, 'w', encoding='utf-8') as f:
            json.dump(result_data, f, ensure_ascii=False, indent=2)

        # 5. Загрузка результата
        print(f"[INFO] Загрузка саммари {output_url}")
        with open(local_output, 'rb') as fout:
            resp = requests.put(output_url, data=fout)
            resp.raise_for_status()
        print("[SUCCESS] Суммаризация завершена.")

    finally:
        for f in [local_input, local_output]:
            if os.path.exists(f):
                os.remove(f)

@app.post("/summarize")
async def summarize(request: SummarizeRequest):
    task_id = str(uuid.uuid4())[:8]
    try:
        run_summarization(request.input_url, request.output_url, request.prompt, task_id)
        return {"status": "ok"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))