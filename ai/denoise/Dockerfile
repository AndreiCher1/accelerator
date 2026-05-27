FROM nvidia/cuda:12.8.0-cudnn-devel-ubuntu24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHONUNBUFFERED=1

RUN apt-get update && apt-get install -y \
    python3.12 \
    python3.12-venv \
    python3.12-dev \
    python3-pip \
    ffmpeg \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# создаем новое вирттуальное окружение
RUN python3.12 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
# если в композе не передали порт, используем по умолчанию
ENV DENOISE_PORT=8001 

# Ставим ночную сборку PyTorch под свежую архитектуру GPU
RUN pip install --pre torch torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/cu128

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Запуск через uvicorn, открываем порт 8001
# ставим 0.0.0.0 чтобы слушать все сетевые интрерфейсы, включая мост, через который докер прокидывает порт на локальную машину
CMD ["uvicorn", "denoise:app", "--host", "0.0.0.0", "--port", "${DENOISE_PORT}"]