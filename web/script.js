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
            const data = JSON.parse(event.data);
            if (data.type === 'stats') {
                document.getElementById('stat-humanize').textContent = data.humanize_count;
                document.getElementById('stat-detect').textContent = data.detect_count;
                document.getElementById('stat-plagiarize').textContent = data.plagiarize_count;
                document.getElementById('stat-research').textContent = data.research_count;
            }
        };
        ws.onclose = () => { setTimeout(connectWebSocket, 3000); };
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
        
        resultsContainer.innerHTML = '';
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
                resultsContainer.innerHTML = `<div class="humanize-result">${escapeHtml(data.text)}</div>`;
                break;
            case 'detect':
                const highlightedHTML = createHighlightedTextHTML(inputText.value, data.detection_result.sentences);
                const gaugeHTML = createGaugeHTML(data.detection_result.overall_score);
                resultsContainer.innerHTML = `<div class="detect-header">${gaugeHTML}</div><div class="highlighted-text-container">${highlightedHTML}</div>`;
                break;
            case 'plagiarize':
                resultsContainer.innerHTML = createPlagiarismReportHTML(data.plagiarism_report);
                break;
            case 'research':
                const researchHTML = marked.parse(data.research_result);
                resultsContainer.innerHTML = `<div class="research-result">${researchHTML}</div>`;
                break;
        }
    }

    function escapeHtml(unsafe) { return unsafe.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;"); }
    function createGaugeHTML(score) {
        let scoreMessage = `Low AI Likelihood`;
        let color = 'var(--green)';
        if (score > 65) { scoreMessage = `High AI Likelihood`; color = 'var(--red)'; }
        else if (score > 35) { scoreMessage = `Moderate AI Likelihood`; color = 'var(--yellow)';}
        return `<div class="ai-score-container"><div class="ai-score-circle" style="--score-color: ${color}; --score-percent: ${score}"><span>${score}%</span></div><p class="ai-score-text">${scoreMessage}</p></div>`;
    }
    function createPlagiarismReportHTML(report) {
        if (report.trim().toUpperCase() === 'UNIQUE') { return `<div class="plagiarism-unique"><h3>No Significant Plagiarism Found</h3><p>The provided text appears to be unique.</p></div>`; }
        const matches = report.split('\n').filter(line => line.includes('MATCH:'));
        let html = '<h3>Potential Matches Found</h3><ul class="plagiarism-list">';
        matches.forEach(match => {
            const parts = match.split('|');
            const phrase = parts[0].replace('MATCH:', '').trim();
            const source = parts[1] ? parts[1].replace('SOURCE:', '').trim() : '#';
            html += `<li><p>"${phrase}"</p><a href="${source}" target="_blank" rel="noopener noreferrer">Possible Source</a></li>`;
        });
        html += '</ul>';
        return html;
    }

    // --- Initial Setup ---
    updateUIForAction();
    connectWebSocket();
});