docker run --rm --gpus all --ipc=host \
  -v "$(pwd):/app" \
  -v "$(pwd)/models:/app/models" \
  asr-worker