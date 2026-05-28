# для одноразового запуска, прокидывает локальную папку и gpu
#!/bin/bash
echo "Сборка Docker-образа..."
docker build -t denoise-worker .

echo "Запуск пайплайна шумоподавления речи через Docker..."
# Пробрасываем текущую папку внутрь контейнера, чтобы он видел app/input.wav и сохранял модели
docker run --rm \
    --gpus all \
    -v "$(pwd)":/workspace \
    denoise-worker
    