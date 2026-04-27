I'll architect this system for you. This is a local media generation pipeline that uses Ollama for orchestration, not for actual video/audio processing (since LLMs can't generate video). Let me break down the architecture:

System Architecture: Tilawat4Heart Video Generator

Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    USER INTERFACE LAYER                      │
│         (CLI / Web UI / Natural Language Parser)             │
│   "Create clip for Chapter 2, Verses 20-23 with image xyz"  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              OLLAMA INTELLIGENCE LAYER (Local LLM)           │
│  • Parses natural language requests                          │
│  • Validates chapter/verse ranges against JSON Quran DB      │
│  • Selects appropriate audio segments from local files       │
│  • Selects images from Documents/Tilawat4Heart               │
│  • Generates FFmpeg command sequences                        │
│  • Returns structured execution plans (JSON)                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              MEDIA PROCESSING ENGINE (FFmpeg)                │
│  • Audio slicing (from full Surah audio files)               │
│  • Image sequencing / Ken Burns effect                       │
│  • Arabic + English subtitle burning                         │
│  • Video encoding (MP4 with H.264/AAC)                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              DATA LAYER                                      │
│  • Quran JSON (structured: chapter, verse, arabic, english)  │
│  • Audio files: Documents/Tilawat4Heart/audio/               │
│  • Images: Documents/Tilawat4Heart/images/                   │
│  • Output: Documents/Tilawat4Heart/output/                   │
└─────────────────────────────────────────────────────────────┘
```

---

1. Project Structure

```
Tilawat4Heart/
├── data/
│   └── quran.json                 # Your structured Quran data
├── audio/
│   ├── 001_al_fatiha.mp3          # Full Surah audio files
│   ├── 002_al_baqarah.mp3
│   └── ...
├── images/
│   ├── nature_001.jpg
│   ├── abstract_002.png
│   └── xyz.jpg                    # Your referenced image
├── output/                        # Generated clips
├── src/
│   ├── parser.py                  # NL → structured request
│   ├── media_builder.py           # FFmpeg command generator
│   ├── subtitle_renderer.py       # ASS/SSA subtitle generation
│   └── main.py                    # Orchestrator
└── prompts/
    └── video_generator.txt        # Ollama system prompt
```

---

2. Quran JSON Schema (Expected Format)

Your JSON should follow this structure for seamless integration:

```json
{
  "chapters": [
    {
      "id": 2,
      "name": "Al-Baqarah",
      "verses": [
        {
          "verse_number": 20,
          "arabic": "يَكَادُ الْبَرْقُ يَخْطَفُ أَبْصَارَهُمْ...",
          "english": "The lightning almost snatches away their sight...",
          "audio_start": 245.5,
          "audio_end": 258.2
        }
      ],
      "audio_file": "002_al_baqarah.mp3"
    }
  ]
}
```

> Critical: If your JSON doesn't have `audio_start`/`audio_end` timestamps, you'll need to either:
- Add them manually (tedious for 6,000+ verses)
- Use AI audio segmentation (whisper.cpp locally)
- Use per-verse audio files if available

---

3. Ollama System Prompt (`prompts/video_generator.txt`)

```text
You are Tilawat4Heart Engine, a Quran video generation orchestrator.

## Your Role
Convert natural language requests into structured FFmpeg execution plans. You do NOT generate video directly—you generate precise JSON instructions for the media pipeline.

## Input Format
User requests like: "Create clip for chapter 2 verses 20-23 with image xyz, slow transitions"

## Output Format (STRICT JSON)
{
  "request_id": "uuid",
  "chapter_id": 2,
  "verse_range": {"start": 20, "end": 23},
  "assets": {
    "audio_source": "Documents/Tilawat4Heart/audio/002_al_baqarah.mp3",
    "image_source": "Documents/Tilawat4Heart/images/xyz.jpg",
    "output_path": "Documents/Tilawat4Heart/output/ch2_v20-23_xyz.mp4"
  },
  "timing": {
    "audio_start_seconds": 245.5,
    "audio_end_seconds": 278.0,
    "total_duration": 32.5
  },
  "visual_style": {
    "image_effect": "ken_burns_slow_zoom",
    "subtitle_layout": "bilingual_stacked",
    "transition": "fade"
  },
  "subtitles": [
    {
      "verse": 20,
      "start": 0.0,
      "end": 8.5,
      "arabic": "يَكَادُ الْبَرْقُ يَخْطَفُ أَبْصَارَهُمْ...",
      "english": "The lightning almost snatches away their sight..."
    }
  ],
  "ffmpeg_command": "ffmpeg -ss 245.5 -t 32.5 -i audio/002_al_baqarah.mp3 ..."
}

## Rules
- ALWAYS validate chapter_id exists (1-114)
- ALWAYS validate verse_range exists in that chapter
- Resolve image filenames fuzzily (xyz matches xyz.jpg, xyz.png)
- Default to 1080p output, H.264, AAC 192kbps
- Include full FFmpeg command with subtitle burn-in
```

---

4. Core Python Implementation

`src/parser.py` — Ollama Integration

```python
import ollama
import json
import re
from pathlib import Path

class QuranRequestParser:
    def __init__(self, model="llama3.2", quran_json_path="data/quran.json"):
        self.client = ollama.Client(host='http://localhost:11434')
        self.model = model
        self.quran_data = self._load_quran(quran_json_path)
        
    def _load_quran(self, path):
        with open(path, 'r', encoding='utf-8') as f:
            return json.load(f)
    
    def parse_request(self, user_input: str) -> dict:
        """
        Send request to Ollama with Quran context for intelligent parsing
        """
        # Build context from Quran JSON
        chapters_context = self._build_chapters_context()
        
        prompt = f"""You are the Tilawat4Heart video generation engine.

Available Quran Chapters (sample):
{chapters_context}

Available Images Directory: Documents/Tilawat4Heart/images/
Available Audio Directory: Documents/Tilawat4Heart/audio/

User Request: "{user_input}"

Parse this into a structured video generation plan. Return ONLY valid JSON.
"""
        
        response = self.client.generate(
            model=self.model,
            prompt=prompt,
            system="You generate precise JSON execution plans for Quran video clips. Never hallucinate verse numbers. Only use verified chapter/verse combinations.",
            format="json",
            options={"temperature": 0.1}  # Low temp for accuracy
        )
        
        plan = json.loads(response['response'])
        return self._validate_and_enrich(plan)
    
    def _build_chapters_context(self):
        """Build lightweight context of available chapters"""
        ctx = []
        for ch in self.quran_data['chapters'][:5]:  # Send first 5 as examples
            ctx.append(f"Ch {ch['id']}: {ch['name']} ({len(ch['verses'])} verses)")
        return "\n".join(ctx)
    
    def _validate_and_enrich(self, plan: dict) -> dict:
        """Cross-validate Ollama output against actual Quran data"""
        ch_id = plan['chapter_id']
        v_start = plan['verse_range']['start']
        v_end = plan['verse_range']['end']
        
        chapter = next((c for c in self.quran_data['chapters'] if c['id'] == ch_id), None)
        if not chapter:
            raise ValueError(f"Chapter {ch_id} not found")
        
        verses = [v for v in chapter['verses'] if v_start <= v['verse_number'] <= v_end]
        if len(verses) != (v_end - v_start + 1):
            raise ValueError(f"Invalid verse range {v_start}-{v_end} for chapter {ch_id}")
        
        # Enrich with actual verse data
        plan['subtitles'] = verses
        plan['assets']['audio_source'] = f"Documents/Tilawat4Heart/audio/{chapter['audio_file']}"
        
        # Calculate timing if timestamps exist
        if 'audio_start' in verses[0]:
            plan['timing']['audio_start_seconds'] = verses[0]['audio_start']
            plan['timing']['audio_end_seconds'] = verses[-1]['audio_end']
            plan['timing']['total_duration'] = verses[-1]['audio_end'] - verses[0]['audio_start']
        
        return plan
```

`src/media_builder.py` — FFmpeg Pipeline

```python
import subprocess
import tempfile
import os
from pathlib import Path
from dataclasses import dataclass

@dataclass
class VideoConfig:
    resolution: str = "1920x1080"
    fps: int = 30
    video_bitrate: str = "4M"
    audio_bitrate: str = "192k"
    subtitle_font_size_arabic: int = 48
    subtitle_font_size_english: int = 32

class FFmpegVideoBuilder:
    def __init__(self, config: VideoConfig = None):
        self.config = config or VideoConfig()
    
    def generate_subtitle_ass(self, plan: dict, output_path: str) -> str:
        """
        Generate Advanced SubStation Alpha (.ass) subtitles
        Supports Arabic + English stacked layout with custom styling
        """
        ass_content = """[Script Info]
Title: Tilawat4Heart Subtitles
ScriptType: v4.00+
PlayResX: 1920
PlayResY: 1080

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Arabic,Amiri,48,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,1,0,0,0,100,100,0,0,1,3,0,2,100,100,200,1
Style: English,Roboto,32,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,0,2,100,100,100,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
"""
        
        for sub in plan['subtitles']:
            start = self._seconds_to_ass_time(sub['start'])
            end = self._seconds_to_ass_time(sub['end'])
            
            # Arabic line (top)
            arabic_line = f"Dialogue: 0,{start},{end},Arabic,,0,0,0,,{sub['arabic']}"
            # English line (bottom)
            english_line = f"Dialogue: 0,{start},{end},English,,0,0,0,,{sub['english']}"
            
            ass_content += arabic_line + "\n" + english_line + "\n"
        
        ass_path = output_path.replace('.mp4', '.ass')
        with open(ass_path, 'w', encoding='utf-8') as f:
            f.write(ass_content)
        return ass_path
    
    def build_ffmpeg_command(self, plan: dict, subtitle_path: str) -> list:
        """
        Construct FFmpeg command for video generation
        """
        assets = plan['assets']
        timing = plan['timing']
        
        # Input 0: Image (looped)
        # Input 1: Audio (sliced)
        cmd = [
            'ffmpeg',
            '-y',  # Overwrite output
            # Image input with slow zoom (Ken Burns)
            '-loop', '1',
            '-i', assets['image_source'],
            # Audio input with trim
            '-ss', str(timing['audio_start_seconds']),
            '-t', str(timing['total_duration']),
            '-i', assets['audio_source'],
            # Video filters
            '-vf', (
                f"zoompan=z='min(zoom+0.0005,1.5)':"
                f"d={int(timing['total_duration'] * self.config.fps)}:"
                f"s={self.config.resolution}:fps={self.config.fps},"
                f"ass={subtitle_path}"
            ),
            # Audio codec
            '-c:a', 'aac',
            '-b:a', self.config.audio_bitrate,
            # Video codec
            '-c:v', 'libx264',
            '-b:v', self.config.video_bitrate,
            '-pix_fmt', 'yuv420p',
            '-shortest',  # Match shortest input
            assets['output_path']
        ]
        return cmd
    
    def render(self, plan: dict) -> str:
        """
        Full render pipeline: subtitles → FFmpeg → output video
        """
        # Create temp subtitle file
        subtitle_path = self.generate_subtitle_ass(plan, plan['assets']['output_path'])
        
        # Build and execute FFmpeg
        cmd = self.build_ffmpeg_command(plan, subtitle_path)
        
        print(f"Executing: {' '.join(cmd)}")
        result = subprocess.run(cmd, capture_output=True, text=True)
        
        if result.returncode != 0:
            raise RuntimeError(f"FFmpeg failed: {result.stderr}")
        
        # Cleanup subtitle file (optional)
        # os.remove(subtitle_path)
        
        return plan['assets']['output_path']
    
    @staticmethod
    def _seconds_to_ass_time(seconds: float) -> str:
        """Convert seconds to ASS time format H:MM:SS.cc"""
        hours = int(seconds // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)
        centis = int((seconds % 1) * 100)
        return f"{hours}:{minutes:02d}:{secs:02d}.{centis:02d}"
```

`src/main.py` — Orchestrator

```python
#!/usr/bin/env python3
import argparse
import sys
from pathlib import Path
from parser import QuranRequestParser
from media_builder import FFmpegVideoBuilder, VideoConfig

def main():
    parser = argparse.ArgumentParser(description='Tilawat4Heart Quran Video Generator')
    parser.add_argument('request', nargs='?', help='Natural language request')
    parser.add_argument('--chapter', '-c', type=int, help='Chapter number')
    parser.add_argument('--verses', '-v', help='Verse range (e.g., 20-23)')
    parser.add_argument('--image', '-i', help='Image filename')
    parser.add_argument('--output', '-o', help='Output filename')
    parser.add_argument('--model', default='llama3.2', help='Ollama model')
    
    args = parser.parse_args()
    
    # Ensure Ollama is running
    try:
        parser_engine = QuranRequestParser(model=args.model)
    except Exception as e:
        print(f"Error connecting to Ollama: {e}")
        print("Ensure Ollama is running: ollama serve")
        sys.exit(1)
    
    # Build request from CLI args or natural language
    if args.request:
        user_input = args.request
    else:
        user_input = f"Create clip for chapter {args.chapter} verses {args.verses} with image {args.image}"
        if args.output:
            user_input += f" output as {args.output}"
    
    print(f"🎯 Parsing request: {user_input}")
    
    # Step 1: Ollama parses and validates
    try:
        plan = parser_engine.parse_request(user_input)
        print(f"✅ Plan generated: Ch {plan['chapter_id']}, Verses {plan['verse_range']['start']}-{plan['verse_range']['end']}")
    except Exception as e:
        print(f"❌ Parsing failed: {e}")
        sys.exit(1)
    
    # Step 2: Build video
    print("🎬 Rendering video...")
    builder = FFmpegVideoBuilder(VideoConfig())
    
    try:
        output_path = builder.render(plan)
        print(f"✅ Video saved: {output_path}")
    except Exception as e:
        print(f"❌ Render failed: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
```

---

5. Usage Examples

```bash
# Natural language (Ollama parses everything)
python src/main.py "Create a beautiful clip for Surah Baqarah verses 20-23 using the sunset image, slow zoom effect"

# Explicit parameters (bypass NL parsing)
python src/main.py -c 2 -v 20-23 -i xyz.jpg -o my_clip.mp4

# With custom style
python src/main.py "Chapter 2 verses 255-257 with image nature_001, dramatic transitions, 4K quality"
```

---

6. Prerequisites & Setup

Component	Installation	Purpose	
Ollama	`curl -fsSL https://ollama.com/install.sh \| sh`	Local LLM inference	
FFmpeg	`brew install ffmpeg` (Mac) / `apt install ffmpeg` (Linux)	Video/audio processing	
Python	3.9+	Orchestration	
Fonts	Amiri (Arabic), Roboto (English)	Subtitle rendering	

```bash
# Setup commands
ollama pull llama3.2  # or mistral, qwen2.5 for better Arabic

# Install Python deps
pip install ollama pathlib

# Verify FFmpeg has ASS support
ffmpeg -filters | grep ass
```

---

7. Key Design Decisions

Decision	Rationale	
Ollama for parsing, not generation	LLMs can't generate video/audio. Ollama converts your natural language into structured machine instructions.	
FFmpeg for actual rendering	Industry-standard, handles Arabic text via libass, supports complex filters (zoompan for Ken Burns).	
ASS subtitles over SRT	ASS supports Arabic shaping, custom positioning (stacked bilingual), and styling per line.	
JSON as single source of truth	Your Quran data drives both validation and subtitle content.	
Local-only processing	No cloud dependencies, works offline, preserves religious data privacy.	

---

8. Enhancement Roadmap

1. Audio Timestamp Detection: If your JSON lacks `audio_start`/`audio_end`, integrate `whisper.cpp` to auto-segment full Surah MP3s by verse.
2. Image Sequences: Support multiple images with crossfade transitions between verse changes.
3. Template System: Predefined visual themes (Ramadan, Nature, Minimalist) with Ollama selecting appropriate filters.
4. Batch Generation: `"Create clips for all of Chapter 2, one verse per clip, using random nature images"`
5. GPU Acceleration: Add `-hwaccel cuda` for NVIDIA GPUs to speed up encoding.

Would you like me to elaborate on any component—particularly the audio timestamp extraction strategy, or help you adapt this to your specific JSON structure?
