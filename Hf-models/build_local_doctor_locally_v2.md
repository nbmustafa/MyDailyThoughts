**Here’s your complete, up-to-date (April 2026) step-by-step guide** to do exactly what you asked for with the fine-tuned domain model (Mini-Lawyer or Local Doctor) from the blueprint.

I’m assuming you have already run the fine-tuning Colab/notebook and have the model saved locally in `./fine_tuned_domain_slm` (merged 4-bit) using the code from the previous blueprint. If not, run the final `model.save_pretrained_merged(...)` step first.

---

### 1. Publish the Model on Hugging Face Hub

Unsloth makes this extremely easy with `push_to_hub_merged()`.

#### Step-by-step:

1. **Get a Hugging Face write token**  
   - Go to https://huggingface.co/settings/tokens  
   - Create a new token with **Write** permission.  
   - Copy it.

2. **Add this code right after your fine-tuning** (in the same notebook or a new cell):

```python
from huggingface_hub import login
from unsloth import FastLanguageModel

# Login once (you only need to do this once per session)
login()   # Paste your HF token when prompted

# -------------------------------
# Push merged model to HF Hub
# -------------------------------
HF_REPO_ID = "yourusername/mini-lawyer"   # ← CHANGE THIS (e.g. nashwansoorchi/mini-lawyer)

model.push_to_hub_merged(
    HF_REPO_ID,
    tokenizer,
    save_method="merged_4bit",   # Keeps it small & fast
    token=True,                  # Uses the logged-in token
)

print(f"✅ Model successfully pushed to https://huggingface.co/{HF_REPO_ID}")
```

   - **Alternative one-liner** if you prefer not to use `login()`:
     ```python
     model.push_to_hub_merged(HF_REPO_ID, tokenizer, save_method="merged_4bit", token="hf_xxxxxxxxxxxxxxxx")
     ```

3. **What happens?**  
   - Unsloth automatically creates the repo (if it doesn’t exist) and uploads the merged 4-bit model + tokenizer + config.  
   - The model will appear on your HF profile within seconds.

4. **Optional but recommended**:
   - Add a **model card** (README.md) on the HF repo page describing:
     - Base model used
     - Domain (medical/legal)
     - Fine-tuning dataset size
     - Intended use + safety note (especially important for medical/legal)
   - You can edit this directly on the HF website.

Your model is now publicly available (or private if you set the repo to private) and anyone can pull it with `FastLanguageModel.from_pretrained(...)`.

---

### 2. Build & Run It Locally for Actual Usage

You have **three practical ways** to run it locally. I recommend **Option A (Ollama)** for most people — it’s the simplest and most user-friendly.

#### Option A: Easiest & Recommended — Run with Ollama (chat UI in seconds)

1. **Install Ollama** (if not already installed)  
   - Download from https://ollama.com/download and install (works on Mac, Windows, Linux).

2. **Export your model to GGUF format** (the format Ollama loves)  
   Add this code **after fine-tuning** (or load the saved model first):

```python
# Export to GGUF (Q4_K_M is the sweet spot: great quality + small size)
model.save_pretrained_gguf(
    "gguf_export",                  # folder name
    tokenizer,
    quantization_method="q4_k_m"    # options: q4_k_m, q5_k_m, q8_0, f16
)
print("✅ GGUF exported to ./gguf_export/")
```

   This creates a single `.gguf` file (usually 2–5 GB depending on model size).

3. **Create a Modelfile** (in the same folder as the GGUF file)  
   Create a text file named `Modelfile` with this content:

```dockerfile
FROM ./gguf_export/unsloth_merged_4bit.gguf   # ← exact filename from the export

TEMPLATE """<|user|>
{{ .Prompt }}<|assistant|>
"""

SYSTEM """You are a highly accurate {{domain}} assistant. Always base your answers strictly on the provided context or your trained knowledge. Never hallucinate. If unsure, say "I don't have enough information." Cite sources when possible."""

PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER num_ctx 4096
```

   (Replace `{{domain}}` with “legal” or “medical” as appropriate.)

4. **Import into Ollama** (run in terminal in the folder containing the Modelfile):

```bash
ollama create mini-lawyer -f Modelfile
```

5. **Run it!**

```bash
ollama run mini-lawyer
```

   You now have a full chat interface. Type your questions exactly like you would with ChatGPT.

   Bonus: Use Open WebUI (https://openwebui.com) for a beautiful web UI that connects to your local Ollama.

#### Option B: Run directly in Python (no Ollama, great for custom apps)

After fine-tuning (or loading the saved model):

```python
from unsloth import FastLanguageModel
import torch

model, tokenizer = FastLanguageModel.from_pretrained(
    model_name="./fine_tuned_domain_slm",   # or your HF repo ID
    dtype=None,  # auto
    load_in_4bit=True,
)

model = FastLanguageModel.for_inference(model)   # 2x faster inference

# Simple inference
messages = [
    {"role": "user", "content": "Explain the key elements of a non-disclosure agreement."}
]
inputs = tokenizer.apply_chat_template(messages, tokenize=True, add_generation_prompt=True, return_tensors="pt").to("cuda")

outputs = model.generate(input_ids=inputs, max_new_tokens=512, temperature=0.7)
print(tokenizer.decode(outputs[0], skip_special_tokens=True))
```

#### Option C: Full local RAG app (as shown in the original blueprint)

Use the exact Mermaid architecture from the blueprint + Gradio/Streamlit frontend.  
- Replace the LLM part with the code above (or `ollama` Python library).  
- Keep your Chroma vector DB and embedding model exactly as written.

---

**Security / Sovereignty Tips**  
- Run everything inside Docker with no network access for true air-gapping.  
- For medical/legal use, keep the model and all data on an isolated machine.  
- GGUF Q4_K_M usually gives the best speed/quality trade-off on consumer hardware.

You’re now done!  
Your model is **published on Hugging Face** AND **running locally** as a private, sovereign assistant.

If you hit any error (e.g. exact GGUF filename, HF token issues, or want a ready-to-run Colab with both steps combined), just paste the error here and I’ll give you the exact fix.