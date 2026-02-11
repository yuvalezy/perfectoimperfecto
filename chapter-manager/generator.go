package main

import (
	"fmt"
	"html"
	"os"
	"strings"
)

// generateChapterHTML generates the complete HTML file content for a chapter
func generateChapterHTML(ch Chapter) string {
	var b strings.Builder

	langCode := string(ch.Language)
	b.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="survey.css">`, langCode, html.EscapeString(ch.Title)))

	// Add checkbox CSS if needed
	if hasCheckboxQuestions(ch) {
		b.WriteString(`
    <style>
        .checkbox-group {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .checkbox-option {
            display: flex;
            align-items: center;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.3s ease;
        }

        .checkbox-option:hover {
            border-color: #667eea;
            background-color: #f5f7ff;
        }

        .checkbox-option input[type="checkbox"] {
            margin-right: 12px;
            cursor: pointer;
            width: 18px;
            height: 18px;
        }

        .checkbox-option label {
            cursor: pointer;
            flex: 1;
            color: #333;
        }

        .checkbox-option input[type="checkbox"]:checked + label {
            color: #667eea;
            font-weight: 500;
        }

        .hidden {
            display: none;
        }

        .conditional-section {
            transition: opacity 0.3s ease;
        }
    </style>`)
	}

	b.WriteString(fmt.Sprintf(`
</head>
<body>
    <div class="container">
        <h1>%s</h1>
        <form id="pollForm">`, html.EscapeString(ch.Heading)))

	// Generate question sections
	for _, q := range ch.Questions {
		b.WriteString("\n")
		b.WriteString(generateQuestionHTML(q))
	}

	// Conversation section
	b.WriteString(fmt.Sprintf(`

            <!-- Conversation Section -->
            <div class="question-section">
                <div class="question-title">%s</div>
                <div class="text-input-section">
                    <label for="conversation" style="color: #666; font-size: 14px;">%s</label>
                    <textarea id="conversation" name="conversation" placeholder="%s" required></textarea>
                    <div class="error-message" id="textError" style="display: none;"></div>
                </div>
            </div>`,
		html.EscapeString(ch.Conversation.Title),
		html.EscapeString(ch.Conversation.Label),
		html.EscapeString(ch.Conversation.Placeholder)))

	// Email section
	b.WriteString(fmt.Sprintf(`

            <!-- Email Section -->
            <div class="question-section">
                <div class="question-title"><label for="email">%s</label></div>
                <div class="text-input-section">
                    <input type="email" id="email" name="email" placeholder="%s" required>
                    <div class="error-message" id="emailError" style="display: none;"></div>
                </div>
            </div>`,
		html.EscapeString(ch.EmailLabel),
		html.EscapeString(ch.EmailPlaceH)))

	// reCAPTCHA section
	b.WriteString(`

            <!-- reCAPTCHA -->
            <div class="question-section">
                <div class="g-recaptcha" data-sitekey="6LeyoggsAAAAAAgXzEg9PAC9ypZtr-yyc24cAnN_"></div>
                <div class="error-message" id="recaptchaError" style="display: none;"></div>
            </div>`)

	// Buttons
	b.WriteString(fmt.Sprintf(`

            <!-- Buttons -->
            <div class="button-group">
                <button type="submit" class="submit-btn" id="submitBtn">%s</button>
                <button type="reset" class="reset-btn" id="resetBtn">%s</button>
            </div>
        </form>`,
		html.EscapeString(ch.SubmitText),
		html.EscapeString(ch.ResetText)))

	// Success message
	b.WriteString(fmt.Sprintf(`

        <!-- Success Message -->
        <div class="success-message" id="successMessage">
            %s
        </div>`,
		ch.SuccessMsg))

	// Summary section
	b.WriteString(fmt.Sprintf(`

        <!-- Summary Section -->
        <div class="summary-section" id="summarySection" style="display: none;">
            <h2 class="summary-title">%s</h2>
            <div id="summaryContent"></div>
        </div>
    </div>`,
		html.EscapeString(ch.SummaryTitle)))

	// Generate conditional JS if needed
	if hasConditionalQuestions(ch) {
		b.WriteString(generateConditionalJS(ch))
	}

	// Survey questions JS object
	b.WriteString("\n\n    <script>")
	if !hasConditionalQuestions(ch) {
		b.WriteString("\n")
	}
	b.WriteString(`
        // Define questions for this survey
        window.surveyQuestions = {`)

	for i, q := range ch.Questions {
		jsTitle := strings.ReplaceAll(q.Title, `"`, `\"`)
		b.WriteString(fmt.Sprintf(`
            %s: "%s"`, q.ID, jsTitle))
		if i < len(ch.Questions)-1 || ch.Conversation.Label != "" {
			b.WriteString(",")
		}
	}
	if ch.Conversation.Label != "" {
		jsConvLabel := strings.ReplaceAll(ch.Conversation.Label, `"`, `\"`)
		b.WriteString(fmt.Sprintf(`
            conversation: "%s"`, jsConvLabel))
	}
	b.WriteString(`
        };

        // Define the chapter name for this survey
        window.chapterName = "` + strings.ReplaceAll(ch.ChapterName, `"`, `\"`) + `";
    </script>`)

	// EmailJS and reCAPTCHA scripts
	b.WriteString(`

    <!-- EmailJS Library -->
    <script src="https://cdn.jsdelivr.net/npm/emailjs-com@3/dist/email.min.js"></script>
    <script>
        (function() {
            emailjs.init("wZ_Z4F9Y-8CcFzD2g"); // Replace with your EmailJS public key
        })();
    </script>

    <!-- Google reCAPTCHA v2 -->
    <script src="https://www.google.com/recaptcha/api.js" async defer></script>

    <script src="survey.js?v=20250110005"></script>
</body>
</html>
`)

	return b.String()
}

// generateQuestionHTML generates the HTML for a single question
func generateQuestionHTML(q Question) string {
	var b strings.Builder

	// Comment
	b.WriteString(fmt.Sprintf("\n            <!-- %s -->", strings.ToUpper(q.ID)))

	// Opening div with conditional hidden class
	if q.ConditionalOn != "" {
		b.WriteString(fmt.Sprintf(`
            <div class="question-section conditional-section hidden" id="%s-section">`, q.ID))
	} else {
		b.WriteString(`
            <div class="question-section">`)
	}

	// Question title
	b.WriteString(fmt.Sprintf(`
                <div class="question-title">%s</div>`, html.EscapeString(q.Title)))

	if q.Type == Radio {
		b.WriteString(`
                <div class="options">`)
		for i, opt := range q.Options {
			optID := fmt.Sprintf("%s_%d", q.ID, i+1)
			requiredAttr := ""
			if q.Required && i == 0 {
				requiredAttr = " required"
			}
			b.WriteString(fmt.Sprintf(`
                    <div class="option">
                        <input type="radio" id="%s" name="%s" value="%s"%s>
                        <label for="%s">%s</label>
                    </div>`,
				optID, q.ID, html.EscapeString(opt.Value), requiredAttr,
				optID, html.EscapeString(opt.Label)))
		}
		b.WriteString(`
                </div>`)
	} else if q.Type == Checkbox {
		b.WriteString(`
                <div class="checkbox-group">`)
		for i, opt := range q.Options {
			optID := fmt.Sprintf("%s_%d", q.ID, i+1)
			b.WriteString(fmt.Sprintf(`
                    <div class="checkbox-option">
                        <input type="checkbox" id="%s" name="%s" value="%s">
                        <label for="%s">%s</label>
                    </div>`,
				optID, q.ID, html.EscapeString(opt.Value),
				optID, html.EscapeString(opt.Label)))
		}
		b.WriteString(`
                </div>`)
	}

	b.WriteString(`
            </div>`)

	return b.String()
}

// generateConditionalJS generates the JavaScript for conditional show/hide logic
func generateConditionalJS(ch Chapter) string {
	var b strings.Builder

	b.WriteString(`

    <script>
        // Handle conditional questions
        document.addEventListener('DOMContentLoaded', function() {`)

	// Find all conditional pairs
	for _, q := range ch.Questions {
		if q.ConditionalOn == "" {
			continue
		}

		parentID := q.ConditionalOn
		childID := q.ID

		// Determine the "yes" value based on language
		yesValue := "Si"
		if ch.Language == English {
			yesValue = "Yes"
		}

		inputType := "checkbox"
		if q.Type == Checkbox {
			inputType = "checkbox"
		}

		b.WriteString(fmt.Sprintf(`
            const %sRadios = document.querySelectorAll('input[name="%s"]');
            const %sSection = document.getElementById('%s-section');
            const %sInputs = document.querySelectorAll('#%s-section input[type="%s"]');

            // %s -> %s logic
            %sRadios.forEach(radio => {
                radio.addEventListener('change', function() {
                    if (this.value === '%s') {
                        %sSection.classList.remove('hidden');
                    } else {
                        %sSection.classList.add('hidden');
                        // Clear %s selections
                        %sInputs.forEach(input => {
                            input.checked = false;
                        });
                    }
                });
            });`,
			parentID, parentID,
			childID, childID,
			childID, childID, inputType,
			strings.ToUpper(parentID), strings.ToUpper(childID),
			parentID,
			yesValue,
			childID,
			childID,
			strings.ToUpper(childID),
			childID))
	}

	b.WriteString(`
        });
    </script>`)

	return b.String()
}

// hasCheckboxQuestions returns true if any question is a checkbox type
func hasCheckboxQuestions(ch Chapter) bool {
	for _, q := range ch.Questions {
		if q.Type == Checkbox {
			return true
		}
	}
	return false
}

// hasConditionalQuestions returns true if any question has a conditional dependency
func hasConditionalQuestions(ch Chapter) bool {
	for _, q := range ch.Questions {
		if q.ConditionalOn != "" {
			return true
		}
	}
	return false
}

// saveChapter writes a chapter to its HTML file
func saveChapter(baseDir string, ch Chapter) error {
	filePath := chapterFilePath(baseDir, ch.Number, ch.Language)
	content := generateChapterHTML(ch)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// deleteChapterFile removes a chapter HTML file
func deleteChapterFile(baseDir string, num int, lang Language) error {
	filePath := chapterFilePath(baseDir, num, lang)
	return os.Remove(filePath)
}

// previewChapter returns the generated HTML as a string for preview
func previewChapter(ch Chapter) string {
	return generateChapterHTML(ch)
}
