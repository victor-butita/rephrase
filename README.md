Rephrase AI – Text Analysis & Enhancement Suite
Rephrase AI is a modern, web-based text improvement platform built in Go, designed to help refine, verify, and perfect written content.
It leverages the Google Gemini API to offer three powerful capabilities:

Humanize Robotic Text – Transform AI-like or stiff prose into natural, human-like writing.

Detect AI-Generated Content – Estimate the likelihood that text was produced by AI.

Check for Plagiarism – Identify potential matches from public internet sources.

This project is a full-stack demo featuring a Go backend and a dynamic vanilla JavaScript frontend.

Tip: Replace the placeholder screenshot with one from your running app – it’s your README’s most eye-catching element.

✨ Features
Multi-Functional Interface
Humanize – Rewrite stiff or AI text with adjustable tone and complexity.

Detect AI – Display a live gauge showing the likelihood of AI generation.

Check Plagiarism – Provide a detailed source report of online matches.

Polished User Experience
Live word counter (200-word cap to optimize API usage)

Responsive dark-mode UI

Clear loading states and robust error handling

Robust Go Backend
Resilient API client – Retries failed API calls with exponential backoff

Clean project structure – Organized into cmd/, internal/handlers, internal/services

Secure configuration – API keys stored in environment variables, never hardcoded

🚀 Why Rephrase AI?
As AI-generated content becomes more common, the need to refine and verify text is greater than ever.
Rephrase AI showcases:

API Integration – Uses Google Gemini for advanced text processing

REST API Design – Single clean endpoint /api/process for all actions

Concurrency & Error Handling – Graceful recovery from network/API failures

Full-Stack Development – Go backend + vanilla JS frontend for a seamless experience

🛠 Getting Started
Prerequisites
Go 1.18+

Google Gemini API Key (obtain from Google AI Studio)

Installation

git clone https://github.com/YOUR-USERNAME/rephrase.git

cd rephrase
(Replace YOUR-USERNAME with your GitHub handle)

Create a .env file in the project root:

ini
Copy
Edit
GEMINI_API_KEY=YOUR_GEMINI_API_KEY_HERE
(This is .gitignored for security.)

Install dependencies:


go mod tidy
Run the server:


go run ./cmd/server/
Open in browser:

arduino

http://localhost:8080
🔬 How to Use
Select an Action – Choose Humanize, Detect AI, or Check Plagiarism

Enter Text – Type or paste into the left panel (word counter updates live)

Adjust Options – For Humanize, choose tone & complexity

Process – Click the action button (loader will appear)

View Results – Right panel updates with refined text, AI likelihood score, or plagiarism results

