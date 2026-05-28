#!/bin/bash
set -e

echo "=== Проверка локальной модели GGUF ==="

# Ищем первый попавшийся файл .gguf в примонтированной папке
DETECTED_MODEL=$(find /app/models -maxdepth 1 -name "*.gguf" | head -n 1)

if [ -z "$DETECTED_MODEL" ]; then
    echo "ОШИБКА: Ни один файл .gguf не найден в папке /app/models!"
    echo "Вот что Docker реально видит в этой папке прямо сейчас:"
    ls -la /app/models/
    exit 1
else
    echo "✅ Найдена модель: $DETECTED_MODEL"
fi

echo "=== Запуск основного Python скрипта ==="
# Передаем найденный путь в Python через переменную окружения
export LOCAL_MODEL_PATH="$DETECTED_MODEL"
python3 -u main.py