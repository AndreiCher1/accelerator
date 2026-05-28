docker run --rm \
  --gpus all \
  -e HF_TOKEN="${DIARIZE_HF_TOKEN}" \
  -v $(pwd)/models:/app/models \
  -v $(pwd):/data \
  diarization-app /data/app/clean_speech.wav