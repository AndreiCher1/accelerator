docker run --rm -it \
  --name qwen_worker \
  --gpus all \
  --ipc=host \
  -v "$(pwd)/models:/app/models" \
  -v "$(pwd)/app:/app/src" \
  qwen-local