FROM nvidia/cuda:13.1.0-devel-ubuntu22.04

RUN apt-get update && apt-get install -y python3-pip git wget curl
RUN pip3 install --no-cache-dir --upgrade pip
RUN pip3 install tensorrt_llm -U --pre --extra-index-url https://pypi.nvidia.com
RUN pip3 install huggingface_hub

RUN apt-get update && apt-get install -y openmpi-bin libopenmpi-dev

WORKDIR /app

COPY ./app /app/src
WORKDIR /app/src

RUN chmod +x start.sh
CMD ["./start.sh"]