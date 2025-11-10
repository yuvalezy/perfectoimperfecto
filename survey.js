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

        // Validate email field
        const emailField = document.getElementById('email');
        const emailError = document.getElementById('emailError');
        if (emailField) {
            const emailValue = emailField.value.trim();
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

            if (emailValue.length === 0) {
                if (emailError) {
                    emailError.textContent = 'Por favor ingresa tu correo electrónico.';
                    emailError.style.display = 'block';
                }
                emailField.scrollIntoView({ behavior: 'smooth', block: 'center' });
                return;
            }

            if (!emailRegex.test(emailValue)) {
                if (emailError) {
                    emailError.textContent = 'Por favor ingresa un correo electrónico válido.';
                    emailError.style.display = 'block';
                }
                emailField.scrollIntoView({ behavior: 'smooth', block: 'center' });
                return;
            }
        }

        // Validate checkbox groups - at least one must be selected in each visible group
        const checkboxGroups = document.querySelectorAll('.checkbox-group');
        for (const group of checkboxGroups) {
            // Skip hidden sections
            const section = group.closest('.question-section');
            if (section && section.classList.contains('hidden')) {
                continue;
            }

            const checkboxes = group.querySelectorAll('input[type="checkbox"]');
            if (checkboxes.length > 0) {
                const isAnyChecked = Array.from(checkboxes).some(cb => cb.checked);
                if (!isAnyChecked) {
                    // Find or create error message element
                    let errorMsg = group.parentElement.querySelector('.checkbox-error');
                    if (!errorMsg) {
                        errorMsg = document.createElement('div');
                        errorMsg.className = 'error-message checkbox-error';
                        errorMsg.style.color = '#e74c3c';
                        errorMsg.style.fontSize = '14px';
                        errorMsg.style.marginTop = '8px';
                        group.parentElement.appendChild(errorMsg);
                    }
                    errorMsg.textContent = 'Por favor selecciona al menos una opción.';
                    errorMsg.style.display = 'block';

                    // Scroll to the error
                    section.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    return;
                }
            }
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
                // Multiple values - join them with pipe
                responses[name] = allValues.join(' | ');
            } else if (allValues.length === 1) {
                // Single value
                responses[name] = allValues[0];
            }
        });

        console.log('Respuestas finales:', responses);

        // Generate answers_html for email template
        const answersHtml = generateAnswersHtml(responses);

        // Get user email
        const userEmail = responses.email || '';

        // Send email via EmailJS
        sendEmailJS(answersHtml, userEmail);

        // Show summary
        displaySummary(responses);

        // Hide form and show summary
        form.style.display = 'none';
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

            // Check if value contains multiple selections (has pipe separator)
            let answerHTML;
            if (typeof value === 'string' && value.includes(' | ')) {
                const values = value.split(' | ');

                // Filter out "Todos" or "Todas las anteriores" if present
                const filteredValues = values.filter(v =>
                    v !== 'Todos' &&
                    v !== 'Todas las anteriores'
                );

                // If more than 1 option (after filtering), show as bullet list
                if (filteredValues.length > 1) {
                    answerHTML = '<ul>' + filteredValues.map(v => `<li>${v}</li>`).join('') + '</ul>';
                } else {
                    answerHTML = filteredValues[0] || value;
                }
            } else {
                answerHTML = value;
            }

            summaryHTML += `
                <div class="summary-item">
                    <div class="summary-question">${question}</div>
                    <div class="summary-answer">${answerHTML}</div>
                </div>
            `;
        }

        summaryContent.innerHTML = summaryHTML;
    }

    function generateAnswersHtml(responses) {
        const questions = window.surveyQuestions || {};
        let answersHtml = '';

        for (const [key, value] of Object.entries(responses)) {
            // Skip email field in the answers table
            if (key === 'email') continue;

            const questionText = questions[key] || key;

            // Format the value (handle multiple selections)
            let formattedValue = value;
            if (typeof value === 'string' && value.includes(' | ')) {
                const values = value.split(' | ');
                // Filter out "Todos" or "Todas las anteriores" if present
                const filteredValues = values.filter(v =>
                    v !== 'Todos' &&
                    v !== 'Todas las anteriores'
                );
                formattedValue = filteredValues.join(', ');
            }

            answersHtml += `
                <tr>
                    <td style="padding: 12px; border: 1px solid #ddd; background-color: #f9f9f9; font-weight: 600;">${questionText}</td>
                    <td style="padding: 12px; border: 1px solid #ddd;">${formattedValue}</td>
                </tr>
            `;
        }

        return answersHtml;
    }

    function sendEmailJS(answersHtml, userEmail) {
        // Check if EmailJS is available
        if (typeof emailjs === 'undefined') {
            console.error('EmailJS no está cargado');
            alert('Error: El servicio de correo no está disponible.');
            return;
        }

        const chapterName = window.chapterName || 'Capítulo';

        const templateParams = {
            chapter: chapterName,
            answers_html: answersHtml,
            to_email: userEmail,
            user_email: userEmail
        };

        emailjs.send("service_i5l7p4t", "template_arwcoqm", templateParams)
            .then(() => {
                console.log('Correo enviado exitosamente');
                successMessage.textContent = '✓ ¡Encuesta enviada correctamente!';
                successMessage.style.display = 'block';
            })
            .catch((error) => {
                console.error('Error al enviar el correo:', error);
                successMessage.textContent = '✓ Respuestas guardadas (Error al enviar correo)';
                successMessage.style.display = 'block';
                successMessage.style.background = '#ff9800';
            });
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
    const conversationField = document.getElementById('conversation');
    if (conversationField) {
        conversationField.addEventListener('input', function() {
            if (textError) {
                textError.style.display = 'none';
            }
        });
    }

    // Clear email error on input
    const emailField = document.getElementById('email');
    const emailError = document.getElementById('emailError');
    if (emailField && emailError) {
        emailField.addEventListener('input', function() {
            emailError.style.display = 'none';
        });
    }

    // Clear checkbox error messages when any checkbox in a group is clicked
    const allCheckboxes = document.querySelectorAll('.checkbox-group input[type="checkbox"]');
    allCheckboxes.forEach(checkbox => {
        checkbox.addEventListener('change', function() {
            const group = this.closest('.checkbox-group');
            if (group) {
                const errorMsg = group.parentElement.querySelector('.checkbox-error');
                if (errorMsg) {
                    errorMsg.style.display = 'none';
                }
            }
        });
    });
});
