docker run --rm \
  --gpus all \
  -e HF_TOKEN="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" \
  -v $(pwd)/models:/app/models \
  -v $(pwd):/data \
  diarization-app /data/app/clean_speech.wav