I'll architect this system for you. This is a sophisticated multimedia pipeline that doesn't actually use Ollama for video generation (Ollama is a local LLM runner, not a video editor), but rather uses it as an orchestrator/intelligence layer while leveraging proper multimedia tools for rendering.

System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           QURAN MEDIA GENERATOR                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  LAYER 1: ORCHESTRATION (Ollama/LLM)                                        │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │   Ollama    │◄──►│   Python    │◄──►│   Prompt    │                     │
│  │  (LLaMA 3)  │    │   Agent     │    │  Templates  │                     │
│  └─────────────┘    └─────────────┘    └─────────────┘                     │
│         │                                                                   │
│  LAYER 2: DATA SOURCES                                                      │
│  ┌─────────────┐    ┌─────────────────────┐    ┌─────────────────────┐   │
│  │  Quran JSON │    │  Google Drive Audio │    │  Google Drive Images  │   │
│  │  (Local)    │    │  (Stream/Download)    │    │  (Stream/Download)    │   │
│  └─────────────┘    └─────────────────────┘    └─────────────────────┘   │
│         │                                                                   │
│  LAYER 3: MEDIA ENGINE (FFmpeg + MoviePy)                                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │ Audio Splice│    │ Image→Video │    │  Subtitle   │                     │
│  │   Engine    │    │   Engine    │    │   Burner    │                     │
│  └─────────────┘    └─────────────┘    └─────────────┘                     │
│         │                                                                   │
│  LAYER 4: OUTPUT                                                            │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  ~/QuranClips/ch2_v20-23_*.mp4  (Arabic + English + Audio + Images) │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

Complete Implementation

1. Project Structure

```
quran-media-system/
├── config/
│   ├── settings.yaml          # API keys, paths, Ollama config
│   └── prompts/
│       └── video_generator.txt # System prompt for Ollama
├── data/
│   └── quran.json             # Your Quran JSON file
├── src/
│   ├── __init__.py
│   ├── main.py                # CLI entry point
│   ├── ollama_client.py       # LLM orchestration
│   ├── quran_parser.py        # JSON data access
│   ├── gdrive_client.py       # Google Drive integration
│   ├── media_engine.py        # FFmpeg/MoviePy renderer
│   └── subtitle_generator.py  # ASS/SSA subtitle creation
├── output/                    # Generated clips
├── requirements.txt
└── README.md
```

2. Core Implementation Files

`requirements.txt`

```txt
ollama>=0.1.0
google-api-python-client>=2.100.0
google-auth-httplib2>=0.1.1
google-auth-oauthlib>=1.0.0
moviepy>=1.0.3
ffmpeg-python>=0.2.0
pydub>=0.25.1
Pillow>=10.0.0
requests>=2.31.0
pyyaml>=6.0.1
```

`src/quran_parser.py` — Quran JSON Data Access

```python
import json
from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional, Dict

@dataclass
class Verse:
    chapter: int
    verse: int
    arabic: str
    english: str
    transliteration: Optional[str] = None

class QuranRepository:
    def __init__(self, json_path: str = "data/quran.json"):
        self.json_path = Path(json_path)
        self._data = self._load()
        self._index = self._build_index()
    
    def _load(self) -> Dict:
        with open(self.json_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    
    def _build_index(self) -> Dict:
        """Flattened index for O(1) verse lookup"""
        index = {}
        for chapter in self._data.get('chapters', []):
            ch_num = chapter['chapter_number']
            for verse in chapter.get('verses', []):
                v_num = verse['verse_number']
                index[(ch_num, v_num)] = Verse(
                    chapter=ch_num,
                    verse=v_num,
                    arabic=verse['arabic'],
                    english=verse['english_translation'],
                    transliteration=verse.get('transliteration')
                )
        return index
    
    def get_verses(self, chapter: int, start_verse: int, end_verse: int) -> List[Verse]:
        """Get range of verses (inclusive)"""
        verses = []
        for v in range(start_verse, end_verse + 1):
            if (chapter, v) in self._index:
                verses.append(self._index[(chapter, v)])
            else:
                raise ValueError(f"Verse {chapter}:{v} not found in repository")
        return verses
    
    def get_chapter_info(self, chapter: int) -> Dict:
        """Get chapter metadata"""
        for ch in self._data.get('chapters', []):
            if ch['chapter_number'] == chapter:
                return {
                    'name_arabic': ch.get('name_arabic'),
                    'name_english': ch.get('name_english'),
                    'name_transliterated': ch.get('name_transliterated'),
                    'total_verses': ch.get('total_verses')
                }
        raise ValueError(f"Chapter {chapter} not found")
```

`src/gdrive_client.py` — Google Drive Integration

```python
import os
import io
from pathlib import Path
from typing import Optional, List
from googleapiclient.discovery import build
from googleapiclient.http import MediaIoBaseDownload
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request
import pickle

SCOPES = ['https://www.googleapis.com/auth/drive.readonly']

class GoogleDriveClient:
    def __init__(self, credentials_path: str = "config/credentials.json"):
        self.credentials_path = credentials_path
        self.service = self._authenticate()
        self._cache_dir = Path("cache/gdrive")
        self._cache_dir.mkdir(parents=True, exist_ok=True)
    
    def _authenticate(self):
        creds = None
        token_path = Path("config/token.pickle")
        
        if token_path.exists():
            with open(token_path, 'rb') as token:
                creds = pickle.load(token)
        
        if not creds or not creds.valid:
            if creds and creds.expired and creds.refresh_token:
                creds.refresh(Request())
            else:
                flow = InstalledAppFlow.from_client_secrets_file(
                    self.credentials_path, SCOPES)
                creds = flow.run_local_server(port=0)
            with open(token_path, 'wb') as token:
                pickle.dump(creds, token)
        
        return build('drive', 'v3', credentials=creds)
    
    def find_file(self, name: str, parent_folder: Optional[str] = None) -> Optional[str]:
        """Find file ID by name, optionally in specific folder"""
        query = f"name = '{name}' and trashed = false"
        if parent_folder:
            query += f" and '{parent_folder}' in parents"
        
        results = self.service.files().list(q=query, spaces='drive', fields='files(id, name)').execute()
        files = results.get('files', [])
        return files[0]['id'] if files else None
    
    def download_file(self, file_id: str, filename: str, use_cache: bool = True) -> Path:
        """Download file to cache, return local path"""
        cache_path = self._cache_dir / filename
        
        if use_cache and cache_path.exists():
            return cache_path
        
        request = self.service.files().get_media(fileId=file_id)
        fh = io.BytesIO()
        downloader = MediaIoBaseDownload(fh, request)
        
        done = False
        while not done:
            status, done = downloader.next_chunk()
            print(f"Download {int(status.progress() * 100)}%")
        
        cache_path.write_bytes(fh.getvalue())
        return cache_path
    
    def get_audio_segment(self, chapter: int, start_verse: int, end_verse: int) -> Path:
        """
        Maps verse range to audio file. 
        Strategy: Each chapter has one audio file, or each verse has individual file.
        Adjust based on your Google Drive organization.
        """
        # Example: Look for "002_Al-Baqarah.mp3" or "002_020-023.mp3"
        patterns = [
            f"{chapter:03d}_*.mp3",
            f"{chapter:03d}_{start_verse:03d}-{end_verse:03d}.mp3",
            f"chapter_{chapter}.mp3"
        ]
        
        for pattern in patterns:
            # Simplified - implement actual search logic
            file_id = self.find_file(pattern.replace('*', ''))
            if file_id:
                return self.download_file(file_id, pattern.replace('*', ''))
        
        raise FileNotFoundError(f"Audio for chapter {chapter} verses {start_verse}-{end_verse} not found")
    
    def get_image(self, image_name: str) -> Path:
        """Download specific image by name"""
        file_id = self.find_file(image_name)
        if not file_id:
            raise FileNotFoundError(f"Image '{image_name}' not found in Google Drive")
        return self.download_file(file_id, image_name)
```

`src/subtitle_generator.py` — Advanced ASS Subtitle Generation

```python
from dataclasses import dataclass
from typing import List
from pathlib import Path
import textwrap

@dataclass
class SubtitleEntry:
    start_time: float  # seconds
    end_time: float
    arabic: str
    english: str
    verse_number: int

class QuranSubtitleGenerator:
    def __init__(self):
        self.style = self._generate_ass_style()
    
    def _generate_ass_style(self) -> str:
        """Generate ASS header with proper Arabic/English styling"""
        return """[Script Info]
Title: Quran Video Subtitles
ScriptType: v4.00+
Collisions: Normal
PlayDepth: 0

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Arabic,Amiri,48,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,2,2,2,10,10,30,178
Style: English,Inter,32,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,2,2,2,10,10,100,1
Style: VerseNum,Inter,24,&H00FFD700,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,1,1,8,10,10,10,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
"""
    
    def _format_time(self, seconds: float) -> str:
        """Convert seconds to ASS time format H:MM:SS.cc"""
        hours = int(seconds // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)
        centis = int((seconds % 1) * 100)
        return f"{hours}:{minutes:02d}:{secs:02d}.{centis:02d}"
    
    def _wrap_text(self, text: str, max_chars: int = 40) -> str:
        """Wrap text for subtitle display"""
        lines = textwrap.wrap(text, max_chars)
        return "\\N".join(lines)
    
    def generate_ass(self, entries: List[SubtitleEntry], output_path: Path) -> Path:
        """Generate .ass subtitle file"""
        content = [self.style]
        
        for entry in entries:
            start = self._format_time(entry.start_time)
            end = self._format_time(entry.end_time)
            
            # Arabic line (top-center)
            arabic_wrapped = self._wrap_text(entry.arabic, 35)
            content.append(
                f"Dialogue: 0,{start},{end},Arabic,,0,0,0,,{arabic_wrapped}"
            )
            
            # English line (below Arabic)
            english_wrapped = self._wrap_text(entry.english, 45)
            content.append(
                f"Dialogue: 0,{start},{end},English,,0,0,0,,{english_wrapped}"
            )
            
            # Verse number (top-right corner)
            content.append(
                f"Dialogue: 0,{start},{end},VerseNum,,0,0,0,,{entry.verse_number}"
            )
        
        output_path.write_text("\n".join(content), encoding='utf-8')
        return output_path
    
    def calculate_timing(self, verses: List, audio_duration: float, 
                        padding_start: float = 1.0,
                        padding_end: float = 2.0) -> List[SubtitleEntry]:
        """
        Distribute subtitle timing evenly across audio duration.
        In production, use forced alignment (aeneas/aeneas or whisper-timestamped)
        """
        usable_duration = audio_duration - padding_start - padding_end
        verse_duration = usable_duration / len(verses)
        
        entries = []
        for i, verse in enumerate(verses):
            start = padding_start + (i * verse_duration)
            end = start + verse_duration
            entries.append(SubtitleEntry(
                start_time=start,
                end_time=end,
                arabic=verse.arabic,
                english=verse.english,
                verse_number=verse.verse
            ))
        
        return entries
```

`src/media_engine.py` — FFmpeg/MoviePy Video Renderer

```python
import subprocess
import tempfile
from pathlib import Path
from typing import List, Optional
from dataclasses import dataclass
import shutil

@dataclass
class VideoConfig:
    resolution: tuple = (1920, 1080)
    fps: int = 30
    bitrate: str = "8000k"
    audio_codec: str = "aac"
    video_codec: str = "libx264"
    crf: int = 18  # High quality
    transition_duration: float = 1.5  # Crossfade between images

class MediaEngine:
    def __init__(self, config: VideoConfig = None):
        self.config = config or VideoConfig()
        self.ffmpeg_path = shutil.which("ffmpeg")
        if not self.ffmpeg_path:
            raise RuntimeError("FFmpeg not found. Install: brew install ffmpeg (macOS) or apt install ffmpeg (Linux)")
    
    def _build_ffmpeg_cmd(self, 
                         audio_path: Path,
                         image_paths: List[Path],
                         subtitle_path: Path,
                         output_path: Path,
                         duration: float) -> List[str]:
        """
        Build complex FFmpeg filtergraph for:
        - Image slideshow with Ken Burns effect
        - Audio overlay
        - ASS subtitle burn-in
        """
        inputs = []
        filters = []
        stream_counter = 0
        
        # Add audio input
        inputs.extend(["-i", str(audio_path)])
        audio_input = 0
        stream_counter += 1
        
        # Add image inputs and create slideshow
        for img in image_paths:
            inputs.extend(["-loop", "1", "-i", str(img)])
        
        # Calculate image display duration
        img_duration = duration / len(image_paths)
        
        # Build filter complex
        filter_parts = []
        
        # Scale and pad images to target resolution with Ken Burns
        for i, img in enumerate(image_paths):
            img_idx = i + 1  # +1 because audio is 0
            filter_parts.append(
                f"[{img_idx}:v]scale={self.config.resolution[0]}:{self.config.resolution[1]}:"
                f"force_original_aspect_ratio=decrease,pad={self.config.resolution[0]}:"
                f"{self.config.resolution[1]}:(ow-iw)/2:(oh-ih)/2,"
                f"zoompan=z='min(zoom+0.0015,1.5)':d={int(img_duration * self.config.fps)}:"
                f"x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':s="
                f"{self.config.resolution[0]}x{self.config.resolution[1]}[v{i}];"
            )
        
        # Concatenate images with crossfade
        if len(image_paths) == 1:
            filter_parts.append(f"[v0]trim=duration={duration}[outv];")
        else:
            # Build concat filter for multiple images with xfades
            concat_parts = []
            for i in range(len(image_paths) - 1):
                xfade_dur = min(self.config.transition_duration, img_duration * 0.3)
                if i == 0:
                    concat_parts.append(f"[v0][v1]xfade=transition=fade:duration={xfade_dur}:offset={img_duration - xfade_dur}[vt0];")
                else:
                    concat_parts.append(f"[vt{i-1}][v{i+1}]xfade=transition=fade:duration={xfade_dur}:offset={(i+1)*img_duration - xfade_dur}[vt{i}];")
            
            filter_parts.extend(concat_parts)
            filter_parts.append(f"[vt{len(image_paths)-2}]trim=duration={duration}[outv];")
        
        # Add subtitle burn-in
        filter_parts.append(
            f"[outv]ass={subtitle_path}[finalv]"
        )
        
        # Audio filter (normalize)
        filter_parts.append(
            f"[{audio_input}:a]loudnorm=I=-16:TP=-1.5:LRA=11[aout]"
        )
        
        filter_complex = "".join(filter_parts)
        
        cmd = [
            self.ffmpeg_path,
            "-y",  # Overwrite output
            *inputs,
            "-filter_complex", filter_complex,
            "-map", "[finalv]",
            "-map", "[aout]",
            "-c:v", self.config.video_codec,
            "-preset", "slow",
            "-crf", str(self.config.crf),
            "-c:a", self.config.audio_codec,
            "-b:a", "320k",
            "-r", str(self.config.fps),
            "-pix_fmt", "yuv420p",
            "-movflags", "+faststart",
            str(output_path)
        ]
        
        return cmd
    
    def render(self,
               audio_path: Path,
               image_paths: List[Path],
               subtitle_path: Path,
               output_path: Path,
               duration: Optional[float] = None) -> Path:
        """
        Render final video with audio, images, and subtitles
        """
        # Get audio duration if not provided
        if duration is None:
            probe = subprocess.run(
                [self.ffmpeg_path, "-i", str(audio_path)],
                capture_output=True, text=True
            )
            # Parse duration from stderr (FFmpeg outputs info to stderr)
            # In production, use ffprobe
            duration = 30.0  # Default fallback
        
        cmd = self._build_ffmpeg_cmd(
            audio_path, image_paths, subtitle_path, output_path, duration
        )
        
        result = subprocess.run(cmd, capture_output=True, text=True)
        if result.returncode != 0:
            raise RuntimeError(f"FFmpeg failed: {result.stderr}")
        
        return output_path
```

`src/ollama_client.py` — LLM Orchestration Layer

```python
import ollama
import json
import re
from typing import Dict, List, Optional
from dataclasses import asdict

class QuranVideoOrchestrator:
    def __init__(self, model: str = "llama3"):
        self.model = model
        self.system_prompt = """You are a Quran Video Generation Orchestrator. 
Your job is to parse user requests and output structured JSON for video generation.

Rules:
1. Extract chapter number, verse range, and image references
2. Validate verse ranges exist (Quran has 114 chapters)
3. Suggest appropriate image transitions and effects
4. Output ONLY valid JSON

JSON Schema:
{
  "intent": "generate_video",
  "chapter": int,
  "start_verse": int,
  "end_verse": int,
  "image_references": [str],
  "style": {
    "transition": "fade|slide|zoom|none",
    "subtitle_position": "bottom|top|center",
    "highlight_words": bool
  },
  "audio_source": "gdrive",
  "output_format": "mp4"
}"""
    
    def parse_request(self, user_input: str) -> Dict:
        """Use Ollama to parse natural language into structured command"""
        response = ollama.chat(
            model=self.model,
            messages=[
                {"role": "system", "content": self.system_prompt},
                {"role": "user", "content": user_input}
            ],
            format="json",
            options={"temperature": 0.1}  # Low temp for deterministic parsing
        )
        
        try:
            result = json.loads(response['message']['content'])
            return self._validate_and_enhance(result)
        except json.JSONDecodeError:
            # Fallback regex parser
            return self._fallback_parse(user_input)
    
    def _validate_and_enhance(self, parsed: Dict) -> Dict:
        """Validate chapter/verse ranges and add metadata"""
        chapter = parsed.get('chapter', 1)
        start = parsed.get('start_verse', 1)
        end = parsed.get('end_verse', start)
        
        # Basic validation
        if not 1 <= chapter <= 114:
            raise ValueError(f"Invalid chapter: {chapter}. Quran has 114 chapters.")
        if start < 1 or end < start:
            raise ValueError(f"Invalid verse range: {start}-{end}")
        
        parsed['validated'] = True
        return parsed
    
    def _fallback_parse(self, text: str) -> Dict:
        """Regex fallback when LLM fails"""
        patterns = [
            r"chapter\s+(\d+).*?verse[s]?\s+(\d+)(?:\s*-\s*(\d+))?",
            r"surah\s+(\d+).*?ayah[s]?\s+(\d+)(?:\s*to\s+(\d+))?",
            r"(\d+):(\d+)(?:-(\d+))?"
        ]
        
        for pattern in patterns:
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                chapter = int(match.group(1))
                start = int(match.group(2))
                end = int(match.group(3)) if match.group(3) else start
                
                # Extract image references
                images = re.findall(r"image\s+(\w+)", text, re.IGNORECASE)
                
                return {
                    "intent": "generate_video",
                    "chapter": chapter,
                    "start_verse": start,
                    "end_verse": end,
                    "image_references": images or ["default"],
                    "style": {"transition": "fade", "subtitle_position": "bottom"},
                    "audio_source": "gdrive",
                    "output_format": "mp4",
                    "validated": True
                }
        
        raise ValueError("Could not parse request. Format: 'chapter X verse Y-Z image NAME'")
    
    def generate_preview_description(self, verses: List, chapter_info: Dict) -> str:
        """Generate human-readable preview of what will be created"""
        prompt = f"""Describe this Quran video preview in 2 sentences:
Chapter: {chapter_info['name_english']} ({chapter_info['name_arabic']})
Verses: {verses[0].verse} to {verses[-1].verse}
Content preview: {verses[0].english[:100]}..."""
        
        response = ollama.generate(
            model=self.model,
            prompt=prompt,
            options={"temperature": 0.7}
        )
        return response['response']
```

`src/main.py` — CLI Entry Point

```python
#!/usr/bin/env python3
import argparse
import sys
from pathlib import Path
from datetime import datetime

# Add src to path
sys.path.insert(0, str(Path(__file__).parent))

from quran_parser import QuranRepository
from gdrive_client import GoogleDriveClient
from subtitle_generator import QuranSubtitleGenerator, SubtitleEntry
from media_engine import MediaEngine, VideoConfig
from ollama_client import QuranVideoOrchestrator

class QuranMediaCLI:
    def __init__(self):
        self.quran = QuranRepository()
        self.gdrive = GoogleDriveClient()
        self.subtitles = QuranSubtitleGenerator()
        self.media = MediaEngine()
        self.orchestrator = QuranVideoOrchestrator()
        self.output_dir = Path.home() / "QuranClips"
        self.output_dir.mkdir(exist_ok=True)
    
    def process(self, user_command: str):
        """Main processing pipeline"""
        print(f"🧠 Parsing command: '{user_command}'")
        
        # Step 1: Parse with Ollama
        plan = self.orchestrator.parse_request(user_command)
        print(f"📋 Plan: Chapter {plan['chapter']}, Verses {plan['start_verse']}-{plan['end_verse']}")
        
        # Step 2: Fetch Quran data
        verses = self.quran.get_verses(
            plan['chapter'], 
            plan['start_verse'], 
            plan['end_verse']
        )
        chapter_info = self.quran.get_chapter_info(plan['chapter'])
        print(f"📖 Loaded {len(verses)} verses from {chapter_info['name_english']}")
        
        # Step 3: Fetch audio from Google Drive
        print("🎵 Fetching audio...")
        audio_path = self.gdrive.get_audio_segment(
            plan['chapter'],
            plan['start_verse'],
            plan['end_verse']
        )
        
        # Step 4: Fetch images from Google Drive
        print("🖼️  Fetching images...")
        image_paths = []
        for img_ref in plan['image_references']:
            try:
                img_path = self.gdrive.get_image(f"{img_ref}.jpg")  # or .png
                image_paths.append(img_path)
            except FileNotFoundError:
                print(f"⚠️  Image '{img_ref}' not found, using placeholder")
                # Create placeholder or use default
        
        if not image_paths:
            raise ValueError("No valid images found")
        
        # Step 5: Generate subtitles
        print("📝 Generating subtitles...")
        # Get audio duration (simplified - use ffprobe in production)
        import subprocess
        ffprobe = subprocess.run(
            ["ffprobe", "-v", "error", "-show_entries", "format=duration",
             "-of", "default=noprint_wrappers=1:nokey=1", str(audio_path)],
            capture_output=True, text=True
        )
        duration = float(ffprobe.stdout.strip())
        
        subtitle_entries = self.subtitles.calculate_timing(verses, duration)
        subtitle_path = self.output_dir / f"temp_subtitles_{plan['chapter']}_{plan['start_verse']}.ass"
        self.subtitles.generate_ass(subtitle_entries, subtitle_path)
        
        # Step 6: Render video
        print("🎬 Rendering video with FFmpeg...")
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        output_path = self.output_dir / (
            f"Quran_{plan['chapter']:03d}_"
            f"v{plan['start_verse']:03d}-{plan['end_verse']:03d}_"
            f"{timestamp}.mp4"
        )
        
        self.media.render(
            audio_path=audio_path,
            image_paths=image_paths,
            subtitle_path=subtitle_path,
            output_path=output_path,
            duration=duration
        )
        
        # Step 7: Preview description
        preview = self.orchestrator.generate_preview_description(verses, chapter_info)
        print(f"\n✅ Video created: {output_path}")
        print(f"📝 Preview: {preview}")
        print(f"📊 Size: {output_path.stat().st_size / 1024 / 1024:.1f} MB")
        
        return output_path

def main():
    parser = argparse.ArgumentParser(description="Quran Media Generator")
    parser.add_argument("command", help="Natural language command (e.g., 'create audio clip for chapter 2 verse 20-23 image nature')")
    parser.add_argument("--model", default="llama3", help="Ollama model to use")
    args = parser.parse_args()
    
    cli = QuranMediaCLI()
    try:
        output = cli.process(args.command)
        print(f"\n🎉 Success! Video saved to: {output}")
    except Exception as e:
        print(f"\n❌ Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
```

3. Setup & Usage

Initial Setup

```bash
# 1. Clone/setup project
mkdir quran-media-system && cd quran-media-system
python -m venv venv
source venv/bin/activate

# 2. Install dependencies
pip install -r requirements.txt

# 3. Install FFmpeg
# macOS:
brew install ffmpeg
# Ubuntu/Debian:
sudo apt install ffmpeg

# 4. Setup Ollama
ollama pull llama3

# 5. Setup Google Drive API
# - Go to Google Cloud Console → APIs & Services → Credentials
# - Create OAuth 2.0 credentials, download as `config/credentials.json`
# - Run once to authenticate: python src/main.py "test"

# 6. Place your Quran JSON in data/quran.json
```

Running the System

```bash
# Basic usage
python src/main.py "create audio clip for chapter 2 verse 20-23 image sunset"

# More examples
python src/main.py "generate video for surah 1 ayah 1-7 image mosque"
python src/main.py "make clip chapter 112 verse 1-4 image tawhid"
```

4. Expected JSON Format for Your Quran Data

```json
{
  "chapters": [
    {
      "chapter_number": 2,
      "name_arabic": "البقرة",
      "name_english": "Al-Baqarah",
      "name_transliterated": "Al-Baqarah",
      "total_verses": 286,
      "verses": [
        {
          "verse_number": 20,
          "arabic": "يَكَادُ الْبَرْقُ يَخْطَفُ أَبْصَارَهُمْ...",
          "english_translation": "The lightning almost snatches away their sight...",
          "transliteration": "Yakādu l-barqu yakhtaṭifu abṣārahum..."
        }
      ]
    }
  ]
}
```

5. Advanced Enhancements

Feature	Implementation	
Word-level highlighting	Use `aeneas` for forced alignment, then animate ASS `\t` tags	
Multiple image transitions	Extend FFmpeg filtergraph with `xfade` transitions	
Background nasheed	Mix secondary audio track at -20dB	
Batch processing	Add `--batch` flag with CSV input	
Web UI	Wrap with Gradio/Streamlit	

6. Output Example

Your generated file at `~/QuranClips/Quran_002_v020-023_20240427_143022.mp4` will contain:
- 🎵 Full audio recitation from your Google Drive
- 🖼️ Ken Burns zoom/pan effects on images
- 📝 Arabic text (top, Amiri font) + English translation (bottom, Inter font)
- ✨ Golden verse numbers in corner
- 🎬 1080p H.264 with AAC audio

Would you like me to elaborate on any specific component, such as the Google Drive folder structure conventions, the forced audio alignment for precise word highlighting, or a web-based UI wrapper?