document.addEventListener('DOMContentLoaded', () => {
    // --- State ---
    let currentAction = 'humanize';
    const WORD_LIMIT = 200;

    // --- Element Selectors ---
    const navLinks = document.querySelectorAll('.nav-link');
    const pageTitle = document.getElementById('page-title');
    const processButton = document.getElementById('processButton');
    const inputText = document.getElementById('inputText');
    const wordCountEl = document.getElementById('wordCount');
    const resultsContainer = document.getElementById('results-container');
    const outputPlaceholder = document.getElementById('output-placeholder');
    const loader = document.getElementById('loader');
    const errorMessage = document.getElementById('error-message');
    const optionsWrapper = document.getElementById('options-wrapper');
    const toneSelect = document.getElementById('tone');
    const complexitySelect = document.getElementById('complexity');
    const dialectSelect = document.getElementById('dialect');
    const freezeKeywordsInput = document.getElementById('freezeKeywords');
    
    // --- WebSocket for Live Stats ---
    function connectWebSocket() {
        const ws = new WebSocket(`ws://${window.location.host}/ws`);
        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'stats') {
                    document.getElementById('stat-humanize').textContent = data.humanize_count || 0;
                    document.getElementById('stat-detect').textContent = data.detect_count || 0;
                    document.getElementById('stat-plagiarize').textContent = data.plagiarize_count || 0;
                    document.getElementById('stat-research').textContent = data.research_count || 0;
                }
            } catch (e) {
                console.error("Failed to parse websocket message:", e);
            }
        };
        ws.onclose = () => { setTimeout(connectWebSocket, 3000); };
        ws.onerror = (err) => {
            console.error("WebSocket error:", err);
            ws.close();
        };
    }
    
    // --- Event Listeners ---
    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            currentAction = link.dataset.action;
            navLinks.forEach(l => l.classList.remove('active'));
            link.classList.add('active');
            updateUIForAction();
        });
    });

    inputText.addEventListener('input', validateInputs);
    processButton.addEventListener('click', handleProcessRequest);

    // --- Core Functions ---
    function validateInputs() {
        const text = inputText.value;
        const count = text.trim() === '' ? 0 : text.trim().split(/\s+/).length;
        wordCountEl.textContent = `${count} / ${WORD_LIMIT} words`;
        const isOverLimit = currentAction !== 'research' && count > WORD_LIMIT;
        wordCountEl.classList.toggle('limit-exceeded', isOverLimit);
        processButton.disabled = count === 0 || isOverLimit;
    }

    function updateUIForAction() {
        const actionText = {
            humanize: 'Humanizer', detect: 'AI Detector', 
            plagiarize: 'Plagiarism Check', research: 'AI Research'
        }[currentAction];

        pageTitle.textContent = actionText;
        processButton.textContent = currentAction === 'research' ? 'Research Topic' : actionText;
        
        optionsWrapper.style.display = currentAction === 'humanize' ? 'block' : 'none';
        wordCountEl.style.display = currentAction === 'research' ? 'none' : 'block';
        
        resultsContainer.innerHTML = '';
        resultsContainer.appendChild(outputPlaceholder);
        outputPlaceholder.classList.remove('hidden');
        errorMessage.textContent = '';
        
        inputText.placeholder = currentAction === 'research' ? 'Enter a topic to research...' : 'Enter text to begin...';
        validateInputs();
    }

    async function handleProcessRequest() {
        processButton.classList.add('hidden');
        loader.classList.remove('hidden');
        errorMessage.textContent = '';
        resultsContainer.innerHTML = '';
        outputPlaceholder.classList.add('hidden');

        const requestBody = {
            text: inputText.value,
            action: currentAction,
            tone: toneSelect.value,
            complexity: complexitySelect.value,
            dialect: dialectSelect.value,
            freeze_keywords: freezeKeywordsInput.value
        };

        try {
            const response = await fetch('/api/process', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(requestBody) });
            const data = await response.json();
            if (!response.ok) throw new Error(data.error || 'An unknown error occurred.');
            renderResults(data);
        } catch (error) {
            errorMessage.textContent = error.message;
        } finally {
            loader.classList.add('hidden');
            processButton.classList.remove('hidden');
        }
    }
    
    function renderResults(data) {
        resultsContainer.innerHTML = '';
        switch(data.result_type) {
            case 'humanize':
                // **UI FIX:** Use a div, escape HTML, then replace newlines with <br> to preserve paragraphs without breaking layout.
                const humanizedText = escapeHtml(data.text).replace(/\n/g, '<br>');
                resultsContainer.innerHTML = `<div class="humanize-result">${humanizedText}</div>`;
                break;
            case 'detect':
                const detection = data.detection_result;
                if (!detection) {
                    errorMessage.textContent = "Received an invalid detection result.";
                    return;
                }
                // **UI FIX:** This function now returns raw HTML without <pre> tags.
                const highlightedHTML = createHighlightedTextHTML(inputText.value, detection.red_flags || []);
                const gaugeHTML = createGaugeHTML(detection.overall_score);
                resultsContainer.innerHTML = `<div class="detect-header">${gaugeHTML}<p class="ai-analysis">${escapeHtml(detection.analysis)}</p></div><div class="highlighted-text-container">${highlightedHTML}</div>`;
                break;
            case 'plagiarize':
                // **LOGIC FIX:** This function correctly processes the new structured object.
                resultsContainer.innerHTML = createPlagiarismReportHTML(data.plagiarism_result);
                break;
            case 'research':
                // **LOGIC FIX:** This function correctly builds markdown from the structured object.
                const researchData = data.research_result;
                let markdownString = `## Research on: ${researchData.topic}\n\n`;
                markdownString += `### Executive Summary\n${researchData.executive_summary}\n\n`;
                markdownString += `### Historical Context\n${researchData.historical_context}\n\n`;
                markdownString += `### Core Concepts\n` + researchData.core_concepts.map(c => `- ${c}`).join('\n') + `\n\n`;
                markdownString += `### Controversies & Critiques\n` + researchData.controversies_and_critiques.map(c => `- ${c}`).join('\n') + `\n\n`;
                markdownString += `### Practical Applications\n` + researchData.practical_applications.map(c => `- ${c}`).join('\n');
                const researchHTML = marked.parse(markdownString);
                resultsContainer.innerHTML = `<div class="research-result">${researchHTML}</div>`;
                break;
        }
    }

    function escapeHtml(unsafe) {
        return unsafe.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
    }

    // **UI FIX:** This helper no longer wraps the output in <pre> tags.
    function createHighlightedTextHTML(originalText, sentencesToHighlight) {
        let highlightedText = escapeHtml(originalText);
        const highlightSet = new Set(sentencesToHighlight);
        highlightSet.forEach(sentence => {
            const escapedSentence = escapeHtml(sentence);
            const regex = new RegExp(escapedSentence.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&'), 'g');
            highlightedText = highlightedText.replace(regex, `<mark>${escapedSentence}</mark>`);
        });
        // Convert newlines to <br> for proper rendering in a div.
        return highlightedText.replace(/\n/g, '<br>');
    }

    function createGaugeHTML(score) {
        let scoreMessage = `Low AI Likelihood`;
        let color = 'var(--green)';
        if (score > 65) { scoreMessage = `High AI Likelihood`; color = 'var(--red)'; }
        else if (score > 35) { scoreMessage = `Moderate AI Likelihood`; color = 'var(--yellow)';}
        return `<div class="ai-score-container"><div class="ai-score-circle" style="--score-color: ${color}; --score-percent: ${score}"><span>${score}%</span></div><p class="ai-score-text">${scoreMessage}</p></div>`;
    }
    
    // **LOGIC FIX:** This function correctly processes the structured plagiarism object from the backend.
    function createPlagiarismReportHTML(report) {
        if (!report || !report.is_similarity_found) {
            return `<div class="plagiarism-unique"><h3>No Significant Plagiarism Found</h3><p>The provided text appears to be unique based on our analysis.</p><p class="confidence-score">Overall Confidence: ${((report.overall_confidence || 0) * 100).toFixed(0)}%</p></div>`;
        }

        let html = `<h3>Potential Matches Found</h3><p class="confidence-score">Overall Confidence: ${(report.overall_confidence * 100).toFixed(0)}%</p><ul class="plagiarism-list">`;
        report.matches.forEach(match => {
            html += `<li>
                <p class="plagiarism-text">"${escapeHtml(match.matching_text)}"</p>
                <div class="plagiarism-source">
                    <strong>Possible Source:</strong> ${escapeHtml(match.potential_source)}<br>
                    <strong>Confidence:</strong> ${(match.confidence * 100).toFixed(0)}%
                </div>
            </li>`;
        });
        html += '</ul>';
        return html;
    }

    // --- Initial Setup ---
    updateUIForAction();
    connectWebSocket();
});