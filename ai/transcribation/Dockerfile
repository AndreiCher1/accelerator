FROM nvidia/cuda:12.1.0-base-ubuntu22.04

WORKDIR /app

RUN apt-get update && apt-get install -y \
    python3 python3-pip ffmpeg git sox libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip3 install --no-cache-dir "numpy<2.0.0" typing_extensions wheel setuptools

RUN pip3 install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python3", "main.py"]