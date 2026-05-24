FROM python:3.10-slim

ENV DEBIAN_FRONTEND=noninteractive
ENV HF_HOME=/app/models/huggingface
ENV TORCH_HOME=/app/models/torch

RUN apt-get update && \
    apt-get install -y ffmpeg libsndfile1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN pip install --no-cache-dir --pre torch torchaudio --index-url https://download.pytorch.org/whl/nightly/cu128

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY diarize.py .

ENTRYPOINT ["python", "diarize.py"]