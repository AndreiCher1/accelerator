import json
import os
from tensorrt_llm import LLM, SamplingParams

def prepare_transcript(json_filepath):
    """
    Преобразует сырой JSON в текстовый диалог, отфильтровывая ошибки.
    """
    print(f"Чтение файла: {json_filepath}")
    with open(json_filepath, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    transcript_lines = []
    for entry in data:
        speaker = entry.get("speaker", "UNKNOWN")
        text = entry.get("text", "")
        if "[ОШИБКА ПОДКЛЮЧЕНИЯ" not in text:
            transcript_lines.append(f"{speaker}: {text}")
            
    return "\n".join(transcript_lines)

def main():
    json_path = "/app/src/data/clean_speech_result.json" # Положи свой json в папку data рядом с docker-compose
    engine_dir = "/app/models/engine_int8"
    
    if not os.path.exists(json_path):
        print(f"Файл {json_path} не найден. Создай папку 'data' и положи туда JSON!")
        return

    transcript_text = prepare_transcript(json_path)
    print("\n[Подготовленный текст транскрипции (фрагмент)]:")
    print(transcript_text[:300] + "...\n")

    system_prompt = """Ты — бизнес-аналитик. Твоя задача — глубоко проанализировать транскрипцию совещания.
Сформируй отчет:
1. Executive Summary: Детальное резюме.
2. Decisions Made: Принятые решения.
3. Action Items: Поставленные задачи.
Отвечай строго на русском языке."""

    prompt = f"<|im_start|>system\n{system_prompt}<|im_end|>\n<|im_start|>user\nВот транскрипция:\n{transcript_text}<|im_end|>\n<|im_start|>assistant\n"

    print(f"Загрузка движка TensorRT-LLM из {engine_dir} в видеопамять (VRAM)...")
    llm = LLM(model=engine_dir)
    
    sampling_params = SamplingParams(
        temperature=0.7,
        max_tokens=2048
    )

    print("Генерация ответа. Пожалуйста, подождите...")
    output = llm.generate([prompt], sampling_params=sampling_params)
    
    result_text = output[0].outputs[0].text
    
    print("\n================ РЕЗУЛЬТАТ ================\n")
    print(result_text)
    print("\n===========================================\n")

if __name__ == "__main__":
    main()