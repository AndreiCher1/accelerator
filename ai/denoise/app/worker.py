import os
import numpy as np
import soundfile as sf
from audio_separator.separator import Separator

def process_audio():
    input_file = "app/input.wav"
    output_file = "app/clean_speech.wav"
    models_dir = "models"
    ram_disk = "/dev/shm"
    
    if not os.path.exists(input_file):
        print(f"[ERROR] Входной файл {input_file} не найден.")
        return

    os.makedirs(models_dir, exist_ok=True)

    print("[INFO] Инициализация движка...")
    
    separator = Separator(
        output_dir=ram_disk, 
        output_format="WAV",
        output_single_stem="vocals",
        use_autocast=True,
        model_file_dir=models_dir,
        mdxc_params={
            "segment_size": 512,
            "overlap": 2,
            "batch_size": 4
        }
    )

    print("[INFO] Загрузка весов модели в VRAM...")
    separator.load_model(model_filename="model_bs_roformer_ep_317_sdr_12.9755.ckpt")

    print("[INFO] Чтение потока с диска и обработка в ОЗУ...")
    
    info = sf.info(input_file)
    sr = info.samplerate
    
    chunk_duration = 30
    chunks_in_ram = 3
    block_frames = chunk_duration * chunks_in_ram * sr
    
    processed_audio_list = []
    
    for i, block in enumerate(sf.blocks(input_file, blocksize=block_frames, always_2d=True)):
        print(f"[INFO] Обработка блока {i+1} ({(i)*90} - {(i+1)*90} сек)...")
        
        temp_in = os.path.join(ram_disk, f"batch_in_{i}.wav")
        sf.write(temp_in, block, sr)
        
        generated_files = separator.separate(temp_in)
        
        chunk_processed = False
        
        for file in generated_files:
            if "Vocals" in file:
                clean_path = os.path.join(ram_disk, file)
                if os.path.exists(clean_path) and os.path.getsize(clean_path) > 0:
                    try:
                        clean_audio, _ = sf.read(clean_path, always_2d=True)
                        processed_audio_list.append(clean_audio)
                        os.remove(clean_path)
                        chunk_processed = True
                    except Exception as e:
                        print(f"[WARNING] Ошибка чтения {clean_path}: {e}")
                break
                
        if not chunk_processed:
            print(f"[INFO] Блок {i+1} содержит тишину. Заполняем пустотами для сохранения хронометража.")
            processed_audio_list.append(np.zeros_like(block))
            
        os.remove(temp_in)

    print("[INFO] Выгрузка склеенного файла из ОЗУ на диск...")
    final_audio = np.concatenate(processed_audio_list, axis=0)
    
    sf.write(output_file, final_audio, sr)
    
    del processed_audio_list
    del final_audio

    print(f"[SUCCESS] Файл сохранен как: {output_file}")

if __name__ == "__main__":
    process_audio()