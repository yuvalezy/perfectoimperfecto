// Survey form functionality
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('pollForm');
    const submitBtn = document.getElementById('submitBtn');
    const resetBtn = document.getElementById('resetBtn');
    const successMessage = document.getElementById('successMessage');
    const summarySection = document.getElementById('summarySection');
    const summaryContent = document.getElementById('summaryContent');
    const textError = document.getElementById('textError');

    // Get questions data from the page
    const questions = window.surveyQuestions || {};

    form.addEventListener('submit', function(e) {
        e.preventDefault();

        // Check for required textarea fields
        const conversationField = document.getElementById('conversation');
        if (conversationField && conversationField.value.trim().length === 0) {
            const textErr = document.getElementById('textError') || textError;
            if (textErr) {
                textErr.textContent = 'Por favor completa este campo.';
                textErr.style.display = 'block';
            }
            return;
        }

        const formData = new FormData(form);
        const responses = {};

        // Get all unique field names from the form
        const fieldNames = [...new Set(formData.keys())];

        // For each field, collect all its values (handles multiple checkboxes with same name)
        fieldNames.forEach(name => {
            const allValues = formData.getAll(name);
            console.log(`Field "${name}" has ${allValues.length} value(s):`, allValues);
            if (allValues.length > 1) {
                // Multiple values - join them with comma
                responses[name] = allValues.join(', ');
            } else if (allValues.length === 1) {
                // Single value
                responses[name] = allValues[0];
            }
        });

        console.log('Respuestas finales:', responses);

        // Show summary
        displaySummary(responses);

        // Hide form and show summary
        form.style.display = 'none';
        successMessage.style.display = 'block';
        summarySection.style.display = 'block';

        // Scroll to top smoothly
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    });

    function displaySummary(responses) {
        let summaryHTML = '';

        for (const [key, value] of Object.entries(responses)) {
            const question = questions[key];
            summaryHTML += `
                <div class="summary-item">
                    <div class="summary-question">${question}</div>
                    <div class="summary-answer">${value}</div>
                </div>
            `;
        }

        summaryContent.innerHTML = summaryHTML;
    }

    function lockForm() {
        const inputs = form.querySelectorAll('input, textarea, button');
        inputs.forEach(input => {
            if (input.type !== 'reset') {
                input.disabled = true;
            }
        });
        form.classList.add('locked');
        resetBtn.disabled = true;
    }

    // Clear error on input
    document.getElementById('conversation').addEventListener('input', function() {
        textError.style.display = 'none';
    });
});
