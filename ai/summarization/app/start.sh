set -e

MODEL_REPO="Qwen/Qwen3.5-9B"
MODELS_DIR="/app/models"
HF_DIR="$MODELS_DIR/hf_model"
CKPT_DIR="$MODELS_DIR/ckpt_int8"
ENGINE_DIR="$MODELS_DIR/engine_int8"

echo "=== Проверка движка TensorRT-LLM ==="

if [ ! -d "$ENGINE_DIR" ]; then
    echo "Движок не найден. Начинаем процесс подготовки..."
    echo "Веса уже скачаны вручную, пропускаем загрузку HuggingFace..."
    
    echo "Откат к стабильной версии transformers..."
    pip install "transformers==4.57.3"
    
    echo "Патчим config.json для совместимости..."
    if [ -f "$HF_DIR/config.json" ]; then
        sed -i 's/"model_type": "qwen3_5"/"model_type": "qwen2"/g' "$HF_DIR/config.json"
    fi

    echo "Квантование модели в INT8..."
    python3 -m tensorrt_llm.commands.convert_checkpoint \
        --model_dir "$HF_DIR" \
        --output_dir "$CKPT_DIR" \
        --dtype bfloat16 \
        --use_weight_only \
        --weight_only_precision int8 \
        --architecture QWenForCausalLM

    echo "Сборка .engine файла..."
    trtllm-build \
        --checkpoint_dir "$CKPT_DIR" \
        --output_dir "$ENGINE_DIR" \
        --gemm_plugin bfloat16 \
        --max_batch_size 1 \
        --max_input_len 8192 \
        --max_seq_len 10240

    if [ -f "$HF_DIR/config.json" ]; then
        sed -i 's/"model_type": "qwen2"/"model_type": "qwen3_5"/g' "$HF_DIR/config.json"
    fi

    echo "🎉 Сборка завершена успешно! Движок сохранен."
else
    echo "Скомпилированный движок найден. Пропускаем сборку."
fi

echo "=== Запуск основного Python скрипта ==="
python3 main.py