To give Claude "fresh eyes" on today's news, we will integrate **NewsAPI** (which is free for developers). The script will now fetch the top headlines about "Artificial Intelligence," feed those into Claude's prompt, and then use EngineMailer to send the result.
### 1. Updated GitHub Action Secrets
You now need to add one more secret to your repository:
| Secret Name | Source |
|---|---|
| NEWS_API_KEY | Get it for free at newsapi.org |
### 2. The Enhanced Python Script (send_news.py)
This version fetches real data before talking to Claude.
```python
import os
import smtplib
import requests
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from anthropic import Anthropic

def get_latest_news():
    """Fetch real-time AI headlines from NewsAPI."""
    api_key = os.environ.get("NEWS_API_KEY")
    url = f"https://newsapi.org/v2/everything?q=AI+OR+Artificial+Intelligence&sortBy=publishedAt&language=en&pageSize=10&apiKey={api_key}"
    
    try:
        response = requests.get(url)
        data = response.json()
        articles = data.get('articles', [])
        
        # Format headlines into a string for Claude
        news_string = ""
        for i, art in enumerate(articles, 1):
            news_string += f"{i}. {art['title']} (Source: {art['source']['name']})\n"
        return news_string if news_string else "No news found today."
    except Exception as e:
        print(f"Error fetching news: {e}")
        return "Could not fetch latest news."

def get_claude_article(news_context):
    """Pass real headlines to Claude to generate a unique article."""
    client = Anthropic(api_key=os.environ.get("ANTHROPIC_API_KEY"))
    
    prompt = f"""
    Below are the latest AI news headlines from today:
    {news_context}

    Using these headlines, write a unique, professional daily briefing article. 
    Analyze the trends, provide a cohesive narrative, and format it in clean HTML 
    suitable for an email. Use <h2> tags for sections and <p> for body text.
    """
    
    response = client.messages.create(
        model="claude-3-5-sonnet-20240620",
        max_tokens=2000,
        messages=[{"role": "user", "content": prompt}]
    )
    return response.content[0].text

def send_email(content):
    """Send the HTML article via EngineMailer SMTP."""
    smtp_server = "smtp.enginemailer.com"
    smtp_port = 587
    username = os.environ.get("ENGINEMAILER_USER")
    password = os.environ.get("ENGINEMAILER_API_KEY")
    from_email = os.environ.get("EMAIL_FROM")
    to_email = "distro-list@yourcompany.com"

    msg = MIMEMultipart()
    msg['From'] = from_email
    msg['To'] = to_email
    msg['Subject'] = f"Daily AI Intelligence: {os.popen('date +%Y-%m-%d').read().strip()}"
    msg.attach(MIMEText(content, 'html'))

    try:
        server = smtplib.SMTP(smtp_server, smtp_port)
        server.starttls()
        server.login(username, password)
        server.send_message(msg)
        server.quit()
        print("Success: Email sent via EngineMailer!")
    except Exception as e:
        print(f"Failed to send email: {e}")

if __name__ == "__main__":
    print("Fetching news...")
    current_news = get_latest_news()
    
    print("Generating article with Claude...")
    article_html = get_claude_article(current_news)
    
    print("Dispatching email...")
    send_email(article_html)

```
### 3. Updated YAML Workflow
We need to add requests to the dependencies list and include the new secret.
```yaml
name: Daily AI News Briefing

on:
  schedule:
    - cron: '0 8 * * *' # 8 AM UTC
  workflow_dispatch:

jobs:
  build-and-send:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'

      - name: Install dependencies
        run: pip install anthropic requests

      - name: Generate and Send News
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
          NEWS_API_KEY: ${{ secrets.NEWS_API_KEY }}
          ENGINEMAILER_USER: ${{ secrets.ENGINEMAILER_USER }}
          ENGINEMAILER_API_KEY: ${{ secrets.ENGINEMAILER_API_KEY }}
          EMAIL_FROM: ${{ secrets.EMAIL_FROM }}
        run: python send_news.py

```
### Why this is better:
 * **Accuracy:** Claude isn't guessing; it's summarizing real headlines published in the last 24 hours.
 * **Uniqueness:** Every morning the headlines change, so Claude's narrative will naturally evolve.
 * **Reliability:** EngineMailer's SMTP relay is very stable for high-deliverability distribution lists.
