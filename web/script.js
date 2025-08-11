document.addEventListener('DOMContentLoaded', () => {
    // State
    let currentAction = 'humanize';
    const WORD_LIMIT = 200;

    // Element Selectors
    const tabs = document.querySelectorAll('.tab-btn');
    const processButton = document.getElementById('processButton');
    const inputText = document.getElementById('inputText');
    const wordCountEl = document.getElementById('wordCount');
    const resultsContainer = document.getElementById('results-container');
    const loader = document.getElementById('loader');
    const errorMessage = document.getElementById('error-message');
    const humanizeOptions = document.getElementById('humanize-options');
    const toneSelect = document.getElementById('tone');
    const complexitySelect = document.getElementById('complexity');
    
    // --- Event Listeners ---
    
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            currentAction = tab.dataset.action;
            updateUIForAction();
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
        });
    });

    inputText.addEventListener('input', () => {
        const words = inputText.value.trim().split(/\s+/).filter(Boolean);
        const count = words.length;
        
        wordCountEl.textContent = `${count} / ${WORD_LIMIT} words`;
        wordCountEl.classList.toggle('limit-exceeded', count > WORD_LIMIT);
        
        // Disable button if text is empty or over the limit
        processButton.disabled = count === 0 || count > WORD_LIMIT;
    });

    processButton.addEventListener('click', handleProcessRequest);

    // --- Functions ---

    function updateUIForAction() {
        processButton.textContent = currentAction.charAt(0).toUpperCase() + currentAction.slice(1);
        humanizeOptions.style.display = currentAction === 'humanize' ? 'flex' : 'none';
        resultsContainer.innerHTML = ''; // Clear previous results
        errorMessage.textContent = '';
        inputText.dispatchEvent(new Event('input')); // Re-validate button state
    }

    async function handleProcessRequest() {
        loader.classList.remove('hidden');
        processButton.disabled = true;
        errorMessage.textContent = '';
        resultsContainer.innerHTML = '';

        const requestBody = {
            text: inputText.value,
            action: currentAction,
            tone: toneSelect.value,
            complexity: complexitySelect.value
        };

        try {
            const response = await fetch('/api/process', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(requestBody),
            });

            const data = await response.json();
            if (!response.ok) throw new Error(data.error || 'An unknown error occurred.');
            
            renderResults(data);

        } catch (error) {
            errorMessage.textContent = error.message;
        } finally {
            loader.classList.add('hidden');
            processButton.disabled = false;
            // Re-enable the button but respect the word count validation
            inputText.dispatchEvent(new Event('input')); 
        }
    }
    
    function renderResults(data) {
        // First, ensure the container is clear
        resultsContainer.innerHTML = '';
        
        switch(data.result_type) {
            case 'humanize':
                // Create a textarea for the humanized text
                const resultTextarea = document.createElement('textarea');
                resultTextarea.readOnly = true;
                resultTextarea.textContent = data.text;
                resultsContainer.appendChild(resultTextarea);
                break;
            case 'detect':
                resultsContainer.innerHTML = createGaugeHTML(data.ai_score);
                break;
            case 'plagiarize':
                resultsContainer.innerHTML = createPlagiarismReportHTML(data.plagiarism_report);
                break;
        }
    }

    function createGaugeHTML(score) {
        let scoreMessage = `Low AI Likelihood`;
        let color = 'var(--green)';
        if (score > 65) { scoreMessage = `High AI Likelihood`; color = 'var(--red)'; }
        else if (score > 35) { scoreMessage = `Moderate AI Likelihood`; color = 'var(--yellow)';}
        
        return `
            <div class="ai-score-container">
                <div class="ai-score-circle" style="--score-color: ${color}; --score-percent: ${score}">
                    <span>${score}%</span>
                </div>
                <p class="ai-score-text">${scoreMessage}</p>
            </div>
        `;
    }

    function createPlagiarismReportHTML(report) {
        if (report.trim().toUpperCase() === 'UNIQUE') {
            return `<div class="plagiarism-unique"><h3>No Significant Plagiarism Found</h3><p>The provided text appears to be unique.</p></div>`;
        }
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

    // Initial setup on page load
    updateUIForAction();
});