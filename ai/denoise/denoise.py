import os
import uuid
import requests
import shutil
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from contextlib import asynccontextmanager
from audio_separator.separator import Separator

# ---------- Единоразовая загрузка модели при старте ----------
separator = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global separator
    models_dir = "models"
    ram_disk = "/dev/shm"
    os.makedirs(models_dir, exist_ok=True)
    os.makedirs(ram_disk, exist_ok=True)

    print("[INFO] Инициализация движка денойзинга...")
    separator = Separator(
        output_dir=ram_disk,
        output_format="WAV",
        output_single_stem="vocals",   # или "Vocals" в зависимости от версии
        use_autocast=True,
        model_file_dir=models_dir,
        mdxc_params={
            "segment_size": 512,
            "overlap": 2,
            "batch_size": 4
        }
    )
    separator.load_model(model_filename="model_bs_roformer_ep_317_sdr_12.9755.ckpt")
    print("[INFO] Модель загружена, готов к работе.")
    yield
    # cleanup (необязательно)

app = FastAPI(lifespan=lifespan)

# ---------- Модель запроса ----------
class ProcessRequest(BaseModel):
    input_url: str   # presigned GET на исходный аудиофайл
    output_url: str  # presigned PUT для результата

# ---------- Функция обработки (монолитная, как в первом варианте) ----------
def process_denoise(input_url: str, output_url: str, task_id: str):
    local_input = f"/tmp/{task_id}_input.wav"
    local_output_tmp = None   # путь к временному результату в /dev/shm (создастся моделью)
    local_output_final = f"/tmp/{task_id}_clean.wav"

    try:
        # 1. Потоковое скачивание входного файла
        print(f"[INFO] Скачивание {input_url}")
        r = requests.get(input_url, stream=True)
        r.raise_for_status()
        with open(local_input, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192):   # по 8 КБ
                f.write(chunk)

        # 2. Обработка (точно как в первом рабочем скрипте)
        print(f"[INFO] Обработка файла целиком...")
        generated_files = separator.separate(local_input)

        # Ищем вокалы среди сгенерированных файлов
        local_output_tmp = None
        for file in generated_files:
            if "Vocals" in file or "vocals" in file:
                local_output_tmp = os.path.join(separator.output_dir, file)
                break

        if not local_output_tmp or not os.path.exists(local_output_tmp):
            raise RuntimeError("Модель не вернула файл с вокалом (возможно, тишина).")

        # Перемещаем результат из RAM-диска в обычный временный файл
        shutil.move(local_output_tmp, local_output_final)

        # 3. Загрузка результата по presigned PUT
        with open(local_output_final, 'rb') as fout:
            resp = requests.put(output_url, data=fout)
            resp.raise_for_status()
        print("[SUCCESS] Денойзинг завершён, результат загружен.")

    finally:
        # Удаляем все временные файлы
        for f in [local_input, local_output_final]:
            if f and os.path.exists(f):
                os.remove(f)
        # На всякий случай пытаемся удалить возможный остаток в ram_disk
        if local_output_tmp and os.path.exists(local_output_tmp):
            os.remove(local_output_tmp)

@app.post("/denoise")
async def denoise(request: ProcessRequest):
    task_id = str(uuid.uuid4())[:8]
    try:
        process_denoise(request.input_url, request.output_url, task_id)
        
        return {"status": "ok"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))