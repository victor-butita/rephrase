Rephrase AI – Text Analysis & Enhancement Suite
Rephrase AI is a modern, web-based tool built in Go that provides a suite of services to analyze and perfect text.
It leverages the Google Gemini API to offer three core functions:

Humanize robotic text – Rewrite AI-like or stiff prose into natural, human-like writing.

Detect AI-generated content – Estimate the likelihood that text was produced by AI.

Check for plagiarism – Identify potential matches from public internet sources.

This project is a full-stack demonstration of an API-driven application with a Go backend and a dynamic vanilla JavaScript frontend.

Action Item: Replace the placeholder screenshot with one from your working application – it’s the most important visual in your README.

✨ Features
Multi-Functional Interface
Humanize – Rewrite AI or stiff text into natural prose with adjustable tone & complexity.

Detect AI – Get a percentage score indicating the likelihood of AI generation, displayed on a live gauge.

Check Plagiarism – Scan for similar content online and get a detailed source report.

Polished User Experience
Live word counter (200-word cap to manage API usage)

Responsive, professional dark-mode UI

Clear loading states and error handling

Robust Go Backend
Resilient API Client – Retries failed API calls with exponential backoff

Clean Project Structure – Organized into cmd/, internal/handlers, and internal/services

Secure Config – API keys stored in environment variables, never hardcoded

🚀 Why Rephrase AI?
This project solves a growing challenge: refining AI-generated output and ensuring text integrity.
It showcases:

API Integration – Using Google Gemini for advanced text processing

REST API Design – Clean, stateless endpoint /api/process for multiple actions

Concurrency & Error Handling – Graceful recovery from network/API failures

Full-Stack Development – Go backend + vanilla JS frontend for a seamless interactive experience

🛠️ Getting Started
Prerequisites
Go 1.18+

Google Gemini API Key (get from Google AI Studio)

Installation
bash
Copy
Edit
git clone https://github.com/your-username/rephrase.git
cd rephrase
(Replace your-username with your actual GitHub username)

Create a .env file in the root directory:

ini
Copy
Edit
GEMINI_API_KEY=YOUR_GEMINI_API_KEY_HERE
(This is .gitignored to keep your key safe)

Install dependencies:

bash
Copy
Edit
go mod tidy
Run the server:

bash
Copy
Edit
go run ./cmd/server/
Open in browser:

arduino
Copy
Edit
http://localhost:8080
🔬 How to Use
Select an Action – Click “Humanize”, “Detect AI”, or “Check Plagiarism”

Enter Text – Type or paste into the left panel (word count updates live)

Adjust Options – For “Humanize”, choose tone & complexity

Process – Click the action button (loader appears while processing)

View Results – Right panel updates with refined text, AI likelihood score, or plagiarism report

# rephrase
