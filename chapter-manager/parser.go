package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// parseAllChapters scans the base directory and parses all chapter HTML files
func parseAllChapters(baseDir string) ([]Chapter, error) {
	var chapters []Chapter

	// Scan for Spanish chapters (capitulo_XX.html)
	esChapters, err := parseChaptersByLanguage(baseDir, Spanish)
	if err != nil {
		return nil, fmt.Errorf("parsing Spanish chapters: %w", err)
	}
	chapters = append(chapters, esChapters...)

	// Scan for English chapters (chapter_XX.html)
	enChapters, err := parseChaptersByLanguage(baseDir, English)
	if err != nil {
		return nil, fmt.Errorf("parsing English chapters: %w", err)
	}
	chapters = append(chapters, enChapters...)

	// Sort by language then number
	sort.Slice(chapters, func(i, j int) bool {
		if chapters[i].Language != chapters[j].Language {
			return chapters[i].Language < chapters[j].Language
		}
		return chapters[i].Number < chapters[j].Number
	})

	return chapters, nil
}

// parseChaptersByLanguage finds and parses all chapters of a given language
func parseChaptersByLanguage(baseDir string, lang Language) ([]Chapter, error) {
	prefix := FilePrefix(lang)
	pattern := filepath.Join(baseDir, prefix+"_*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var chapters []Chapter
	for _, f := range files {
		ch, err := parseChapterFile(f, lang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", f, err)
			continue
		}
		chapters = append(chapters, ch)
	}
	return chapters, nil
}

// parseChapterFile parses a single chapter HTML file
func parseChapterFile(filePath string, lang Language) (Chapter, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Chapter{}, err
	}
	content := string(data)

	ch := LanguageDefaults(lang)
	ch.Language = lang

	// Extract chapter number from filename
	base := filepath.Base(filePath)
	numStr := extractChapterNumber(base)
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return Chapter{}, fmt.Errorf("invalid chapter number in %s", base)
	}
	ch.Number = num

	// Extract title
	ch.Title = extractBetweenTags(content, "<title>", "</title>")

	// Extract h1 heading
	ch.Heading = extractBetweenTags(content, "<h1>", "</h1>")

	// Check for custom CSS (checkbox styles)
	ch.HasCustomCSS = strings.Contains(content, ".checkbox-group")

	// Check for custom conditional JS
	ch.HasCustomJS = strings.Contains(content, "conditional") || strings.Contains(content, "classList.remove('hidden')")

	// Extract chapter name from JS
	chNameRe := regexp.MustCompile(`window\.chapterName\s*=\s*"([^"]+)"`)
	if m := chNameRe.FindStringSubmatch(content); len(m) > 1 {
		ch.ChapterName = m[1]
	} else {
		ch.ChapterName = fmt.Sprintf("%s %d", ChapterLabel(lang), num)
	}

	// Extract questions
	ch.Questions = extractQuestions(content)

	// Detect conditional relationships
	detectConditionals(content, ch.Questions)

	// Extract conversation section
	ch.Conversation = extractConversation(content, lang)

	// Extract email label
	emailLabelRe := regexp.MustCompile(`<label for="email">([^<]+)</label>`)
	if m := emailLabelRe.FindStringSubmatch(content); len(m) > 1 {
		ch.EmailLabel = m[1]
	}

	// Extract email placeholder
	emailPlaceholderRe := regexp.MustCompile(`type="email"[^>]*placeholder="([^"]+)"`)
	if m := emailPlaceholderRe.FindStringSubmatch(content); len(m) > 1 {
		ch.EmailPlaceH = m[1]
	}

	// Extract button texts
	submitRe := regexp.MustCompile(`class="submit-btn"[^>]*>([^<]+)</button>`)
	if m := submitRe.FindStringSubmatch(content); len(m) > 1 {
		ch.SubmitText = m[1]
	}
	resetRe := regexp.MustCompile(`class="reset-btn"[^>]*>([^<]+)</button>`)
	if m := resetRe.FindStringSubmatch(content); len(m) > 1 {
		ch.ResetText = m[1]
	}

	// Extract success message
	successRe := regexp.MustCompile(`class="success-message"[^>]*>\s*(.+?)\s*</div>`)
	if m := successRe.FindStringSubmatch(content); len(m) > 1 {
		ch.SuccessMsg = strings.TrimSpace(m[1])
	}

	// Extract summary title
	summaryTitleRe := regexp.MustCompile(`class="summary-title">([^<]+)</h2>`)
	if m := summaryTitleRe.FindStringSubmatch(content); len(m) > 1 {
		ch.SummaryTitle = m[1]
	}

	return ch, nil
}

// extractChapterNumber extracts the number part from a filename like "capitulo_01.html"
func extractChapterNumber(filename string) string {
	re := regexp.MustCompile(`_(\d+)\.html$`)
	if m := re.FindStringSubmatch(filename); len(m) > 1 {
		return m[1]
	}
	return ""
}

// extractBetweenTags extracts content between simple HTML tags
func extractBetweenTags(content, openTag, closeTag string) string {
	start := strings.Index(content, openTag)
	if start == -1 {
		return ""
	}
	start += len(openTag)
	end := strings.Index(content[start:], closeTag)
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(content[start : start+end])
}

// extractQuestions extracts all questions (radio and checkbox) from HTML content
func extractQuestions(content string) []Question {
	var questions []Question

	// Find question titles and their options
	questionTitleRe := regexp.MustCompile(`<div class="question-title">(Q\d+:[^<]+)</div>`)
	matches := questionTitleRe.FindAllStringSubmatchIndex(content, -1)

	for i, match := range matches {
		titleStart := match[2]
		titleEnd := match[3]
		title := content[titleStart:titleEnd]

		// Extract question ID from title
		qIDRe := regexp.MustCompile(`^Q(\d+):`)
		qNum := ""
		if m := qIDRe.FindStringSubmatch(title); len(m) > 1 {
			qNum = m[1]
		} else {
			continue
		}
		qID := "q" + qNum

		// Determine the section boundary (up to next question-section or end)
		sectionEnd := len(content)
		if i+1 < len(matches) {
			sectionEnd = matches[i+1][0]
		}
		section := content[match[0]:sectionEnd]

		// Check if this section is inside a conditional hidden div
		conditionalOn := ""
		// Look backwards from the match to find if it's in a hidden section
		precedingContent := content[:match[0]]
		hiddenSectionRe := regexp.MustCompile(`<div class="question-section conditional-section hidden" id="` + qID + `-section">`)
		if hiddenSectionRe.MatchString(content) {
			conditionalOn = findConditionalParent(qID)
		}

		// Determine question type based on options
		q := Question{
			ID:            qID,
			Title:         title,
			Required:      true,
			ConditionalOn: conditionalOn,
		}

		// Check for checkbox options
		if strings.Contains(section, `type="checkbox"`) {
			q.Type = Checkbox
			q.Options = extractCheckboxOptions(section, qID)
		} else if strings.Contains(section, `type="radio"`) {
			q.Type = Radio
			q.Options = extractRadioOptions(section, qID)
		}

		// Check if required
		if !strings.Contains(section, "required") {
			q.Required = false
		}

		_ = precedingContent
		questions = append(questions, q)
	}

	return questions
}

// extractRadioOptions extracts radio button options from a section
func extractRadioOptions(section, qID string) []Option {
	var options []Option
	optRe := regexp.MustCompile(`<input type="radio"[^>]*name="` + qID + `"[^>]*value="([^"]*)"[^>]*>\s*<label[^>]*>([^<]+)</label>`)
	matches := optRe.FindAllStringSubmatch(section, -1)
	for _, m := range matches {
		options = append(options, Option{
			Value: m[1],
			Label: m[2],
		})
	}
	return options
}

// extractCheckboxOptions extracts checkbox options from a section
func extractCheckboxOptions(section, qID string) []Option {
	var options []Option
	optRe := regexp.MustCompile(`<input type="checkbox"[^>]*name="` + qID + `"[^>]*value="([^"]*)"[^>]*>\s*<label[^>]*>([^<]+)</label>`)
	matches := optRe.FindAllStringSubmatch(section, -1)
	for _, m := range matches {
		options = append(options, Option{
			Value: decodeHTMLEntities(m[1]),
			Label: m[2],
		})
	}
	return options
}

// decodeHTMLEntities decodes basic HTML entities
func decodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return s
}

// encodeHTMLEntities encodes special chars for HTML attribute values
func encodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// findConditionalParent determines the parent question for a conditional question
// Convention: q2 depends on q1, q4 depends on q3, etc. (previous odd-numbered question)
func findConditionalParent(qID string) string {
	numStr := strings.TrimPrefix(qID, "q")
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 1 {
		return ""
	}
	return fmt.Sprintf("q%d", num-1)
}

// detectConditionals sets ConditionalOn field based on HTML content patterns
func detectConditionals(content string, questions []Question) {
	// Look for patterns like: id="q2-section" with class "hidden"
	for i := range questions {
		sectionID := questions[i].ID + "-section"
		hiddenRe := regexp.MustCompile(`id="` + sectionID + `"[^>]*class="[^"]*hidden[^"]*"`)
		classFirstRe := regexp.MustCompile(`class="[^"]*hidden[^"]*"[^>]*id="` + sectionID + `"`)
		if hiddenRe.MatchString(content) || classFirstRe.MatchString(content) {
			questions[i].ConditionalOn = findConditionalParent(questions[i].ID)
		}
	}
}

// extractConversation extracts the conversation/discussion section
func extractConversation(content string, lang Language) ConversationSection {
	defaults := LanguageDefaults(lang)
	conv := defaults.Conversation

	// Extract conversation title (e.g. "Conversen:" or "Discuss:")
	convTitleRe := regexp.MustCompile(`<div class="question-title">((?:Conversen|Discuss|Hablen|Talk)[^<]*)</div>`)
	if m := convTitleRe.FindStringSubmatch(content); len(m) > 1 {
		conv.Title = m[1]
	}

	// Extract conversation label/prompt
	convLabelRe := regexp.MustCompile(`<label for="conversation"[^>]*>([^<]+)</label>`)
	if m := convLabelRe.FindStringSubmatch(content); len(m) > 1 {
		conv.Label = m[1]
	}

	// Extract placeholder
	convPlaceholderRe := regexp.MustCompile(`<textarea[^>]*id="conversation"[^>]*placeholder="([^"]+)"`)
	if m := convPlaceholderRe.FindStringSubmatch(content); len(m) > 1 {
		conv.Placeholder = m[1]
	}

	return conv
}

// chapterFilePath returns the expected file path for a chapter
func chapterFilePath(baseDir string, num int, lang Language) string {
	prefix := FilePrefix(lang)
	return filepath.Join(baseDir, fmt.Sprintf("%s_%02d.html", prefix, num))
}

// chapterExists checks if a chapter file exists
func chapterExists(baseDir string, num int, lang Language) bool {
	_, err := os.Stat(chapterFilePath(baseDir, num, lang))
	return err == nil
}

// getNextChapterNumber finds the next available chapter number
func getNextChapterNumber(baseDir string, lang Language) int {
	prefix := FilePrefix(lang)
	pattern := filepath.Join(baseDir, prefix+"_*.html")
	files, _ := filepath.Glob(pattern)

	maxNum := 0
	for _, f := range files {
		numStr := extractChapterNumber(filepath.Base(f))
		if num, err := strconv.Atoi(numStr); err == nil && num > maxNum {
			maxNum = num
		}
	}
	return maxNum + 1
}
