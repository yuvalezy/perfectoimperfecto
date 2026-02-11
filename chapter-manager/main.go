package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var reader *bufio.Reader

func main() {
	reader = bufio.NewReader(os.Stdin)

	// Determine base directory (parent of chapter-manager/)
	execDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Try to find the base directory containing the HTML files
	baseDir := findBaseDir(execDir)
	if baseDir == "" {
		fmt.Println("Could not find the HTML chapter files directory.")
		fmt.Print("Enter the path to the directory containing the chapter HTML files: ")
		baseDir = readLine()
		baseDir = strings.TrimSpace(baseDir)
	}

	fmt.Println("========================================")
	fmt.Println("  Chapter Manager - Perfecto Imperfecto")
	fmt.Println("========================================")
	fmt.Printf("Working directory: %s\n\n", baseDir)

	for {
		showMainMenu()
		choice := readLine()
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			listChapters(baseDir)
		case "2":
			viewChapter(baseDir)
		case "3":
			editChapter(baseDir)
		case "4":
			createChapter(baseDir)
		case "5":
			removeChapter(baseDir)
		case "6":
			previewChapterMenu(baseDir)
		case "7":
			fmt.Println("\nGoodbye!")
			return
		default:
			fmt.Println("\nInvalid option. Please try again.")
		}
	}
}

func showMainMenu() {
	fmt.Println("--- Main Menu ---")
	fmt.Println("1. List all chapters")
	fmt.Println("2. View chapter (questions & answers)")
	fmt.Println("3. Edit chapter")
	fmt.Println("4. Create new chapter")
	fmt.Println("5. Remove chapter")
	fmt.Println("6. Preview & Save chapter")
	fmt.Println("7. Exit")
	fmt.Print("\nSelect an option: ")
}

func readLine() string {
	line, _ := reader.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func readLinePrompt(prompt string) string {
	fmt.Print(prompt)
	return readLine()
}

func findBaseDir(startDir string) string {
	// Check current directory
	if hasChapterFiles(startDir) {
		return startDir
	}
	// Check parent directory
	parent := filepath.Dir(startDir)
	if hasChapterFiles(parent) {
		return parent
	}
	return ""
}

func hasChapterFiles(dir string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, "capitulo_*.html"))
	return len(matches) > 0
}

func selectLanguage() Language {
	fmt.Println("\nSelect language:")
	fmt.Println("1. Spanish (Español)")
	fmt.Println("2. English")
	fmt.Print("Choice: ")
	choice := strings.TrimSpace(readLine())
	if choice == "2" {
		return English
	}
	return Spanish
}

// ─── 1. LIST ALL CHAPTERS ────────────────────────────────────────────

func listChapters(baseDir string) {
	lang := selectLanguage()

	chapters, err := parseChaptersByLanguage(baseDir, lang)
	if err != nil {
		fmt.Printf("\nError: %v\n\n", err)
		return
	}

	if len(chapters) == 0 {
		langName := "Spanish"
		if lang == English {
			langName = "English"
		}
		fmt.Printf("\nNo %s chapters found.\n\n", langName)
		return
	}

	langName := "Spanish"
	if lang == English {
		langName = "English"
	}
	fmt.Printf("\n=== %s Chapters (%d total) ===\n", langName, len(chapters))
	for _, ch := range chapters {
		qCount := len(ch.Questions)
		fmt.Printf("  %s %d - %d questions - Conversation: %q\n",
			ChapterLabel(lang), ch.Number, qCount,
			truncate(ch.Conversation.Label, 60))
	}
	fmt.Println()
}

// ─── 2. VIEW CHAPTER ─────────────────────────────────────────────────

func viewChapter(baseDir string) {
	lang := selectLanguage()
	num := readChapterNumber()
	if num <= 0 {
		return
	}

	if !chapterExists(baseDir, num, lang) {
		fmt.Printf("\n%s %d does not exist.\n\n", ChapterLabel(lang), num)
		return
	}

	ch, err := parseChapterFile(chapterFilePath(baseDir, num, lang), lang)
	if err != nil {
		fmt.Printf("\nError parsing chapter: %v\n\n", err)
		return
	}

	displayChapter(ch)
}

func displayChapter(ch Chapter) {
	fmt.Printf("\n╔══════════════════════════════════════════╗\n")
	fmt.Printf("║  %s\n", ch.ChapterName)
	fmt.Printf("╚══════════════════════════════════════════╝\n")
	fmt.Printf("  Title:   %s\n", ch.Title)
	fmt.Printf("  Heading: %s\n", ch.Heading)
	fmt.Println()

	for i, q := range ch.Questions {
		condStr := ""
		if q.ConditionalOn != "" {
			condStr = fmt.Sprintf(" [shown when %s = Yes]", q.ConditionalOn)
		}
		fmt.Printf("  ── Question %d (%s%s) ──\n", i+1, q.Type, condStr)
		fmt.Printf("     %s\n", q.Title)
		for j, opt := range q.Options {
			fmt.Printf("       %d) %s\n", j+1, opt.Label)
		}
		fmt.Println()
	}

	fmt.Printf("  ── Conversation Section ──\n")
	fmt.Printf("     Title: %s\n", ch.Conversation.Title)
	fmt.Printf("     Prompt: %s\n", ch.Conversation.Label)
	fmt.Printf("     Placeholder: %s\n", ch.Conversation.Placeholder)
	fmt.Println()
}

// ─── 3. EDIT CHAPTER ─────────────────────────────────────────────────

func editChapter(baseDir string) {
	lang := selectLanguage()
	num := readChapterNumber()
	if num <= 0 {
		return
	}

	if !chapterExists(baseDir, num, lang) {
		fmt.Printf("\n%s %d does not exist.\n\n", ChapterLabel(lang), num)
		return
	}

	ch, err := parseChapterFile(chapterFilePath(baseDir, num, lang), lang)
	if err != nil {
		fmt.Printf("\nError parsing chapter: %v\n\n", err)
		return
	}

	fmt.Printf("\nEditing %s...\n", ch.ChapterName)

	for {
		showEditMenu()
		choice := strings.TrimSpace(readLine())

		switch choice {
		case "1":
			editHeading(&ch)
		case "2":
			editQuestion(&ch)
		case "3":
			addQuestion(&ch)
		case "4":
			removeQuestion(&ch)
		case "5":
			editConversation(&ch)
		case "6":
			editLabels(&ch)
		case "7":
			// Preview
			displayChapter(ch)
		case "8":
			// Save
			if err := saveChapter(baseDir, ch); err != nil {
				fmt.Printf("\nError saving: %v\n", err)
			} else {
				fmt.Printf("\n%s saved successfully to %s\n",
					ch.ChapterName, chapterFilePath(baseDir, ch.Number, ch.Language))
			}
		case "9":
			fmt.Println("\nReturning to main menu.")
			return
		default:
			fmt.Println("\nInvalid option.")
		}
	}
}

func showEditMenu() {
	fmt.Println("\n--- Edit Menu ---")
	fmt.Println("1. Edit heading")
	fmt.Println("2. Edit a question")
	fmt.Println("3. Add a new question")
	fmt.Println("4. Remove a question")
	fmt.Println("5. Edit conversation section")
	fmt.Println("6. Edit labels (email, buttons, etc.)")
	fmt.Println("7. Preview current state")
	fmt.Println("8. Save changes")
	fmt.Println("9. Back to main menu (unsaved changes will be lost)")
	fmt.Print("\nSelect an option: ")
}

func editHeading(ch *Chapter) {
	fmt.Printf("\nCurrent heading: %s\n", ch.Heading)
	newHeading := readLinePrompt("New heading (or press Enter to keep): ")
	if newHeading != "" {
		ch.Heading = newHeading
		fmt.Println("Heading updated.")
	}
}

func editQuestion(ch *Chapter) {
	if len(ch.Questions) == 0 {
		fmt.Println("\nNo questions to edit.")
		return
	}

	fmt.Println("\nQuestions:")
	for i, q := range ch.Questions {
		fmt.Printf("  %d. [%s] %s\n", i+1, q.Type, q.Title)
	}

	idxStr := readLinePrompt("Select question number to edit: ")
	idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
	if err != nil || idx < 1 || idx > len(ch.Questions) {
		fmt.Println("Invalid selection.")
		return
	}
	idx-- // Convert to 0-based

	q := &ch.Questions[idx]

	for {
		fmt.Printf("\n--- Editing %s ---\n", q.ID)
		fmt.Printf("  Title: %s\n", q.Title)
		fmt.Printf("  Type: %s\n", q.Type)
		fmt.Printf("  Options: %d\n", len(q.Options))
		for i, opt := range q.Options {
			fmt.Printf("    %d) %s\n", i+1, opt.Label)
		}
		fmt.Println()
		fmt.Println("a. Edit question title")
		fmt.Println("b. Edit an option")
		fmt.Println("c. Add an option")
		fmt.Println("d. Remove an option")
		fmt.Println("e. Toggle question type (radio/checkbox)")
		fmt.Println("f. Toggle conditional visibility")
		fmt.Println("g. Done editing this question")
		fmt.Print("Choice: ")

		choice := strings.TrimSpace(readLine())
		switch choice {
		case "a":
			newTitle := readLinePrompt("New question title: ")
			if newTitle != "" {
				q.Title = newTitle
				fmt.Println("Title updated.")
			}
		case "b":
			editOption(q)
		case "c":
			addOption(q)
		case "d":
			removeOption(q)
		case "e":
			if q.Type == Radio {
				q.Type = Checkbox
			} else {
				q.Type = Radio
			}
			fmt.Printf("Question type changed to: %s\n", q.Type)
		case "f":
			if q.ConditionalOn != "" {
				q.ConditionalOn = ""
				fmt.Println("Conditional visibility removed (always visible).")
			} else {
				parent := readLinePrompt("Show when which question has 'Yes'? (e.g., q1): ")
				parent = strings.TrimSpace(parent)
				if parent != "" {
					q.ConditionalOn = parent
					fmt.Printf("Will be shown conditionally when %s = Yes.\n", parent)
				}
			}
		case "g":
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}

func editOption(q *Question) {
	if len(q.Options) == 0 {
		fmt.Println("No options to edit.")
		return
	}
	idxStr := readLinePrompt("Option number to edit: ")
	idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
	if err != nil || idx < 1 || idx > len(q.Options) {
		fmt.Println("Invalid selection.")
		return
	}
	idx--

	fmt.Printf("Current label: %s\n", q.Options[idx].Label)
	newLabel := readLinePrompt("New label (or Enter to keep): ")
	if newLabel != "" {
		q.Options[idx].Label = newLabel
		q.Options[idx].Value = newLabel
		fmt.Println("Option updated.")
	}
}

func addOption(q *Question) {
	label := readLinePrompt("New option label: ")
	label = strings.TrimSpace(label)
	if label == "" {
		fmt.Println("Empty label, option not added.")
		return
	}
	q.Options = append(q.Options, Option{Value: label, Label: label})
	fmt.Println("Option added.")
}

func removeOption(q *Question) {
	if len(q.Options) == 0 {
		fmt.Println("No options to remove.")
		return
	}
	for i, opt := range q.Options {
		fmt.Printf("  %d) %s\n", i+1, opt.Label)
	}
	idxStr := readLinePrompt("Option number to remove: ")
	idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
	if err != nil || idx < 1 || idx > len(q.Options) {
		fmt.Println("Invalid selection.")
		return
	}
	idx--
	q.Options = append(q.Options[:idx], q.Options[idx+1:]...)
	fmt.Println("Option removed.")
}

func addQuestion(ch *Chapter) {
	nextNum := len(ch.Questions) + 1
	qID := fmt.Sprintf("q%d", nextNum)

	title := readLinePrompt(fmt.Sprintf("Question title (will be prefixed with Q%d:): ", nextNum))
	title = strings.TrimSpace(title)
	if title == "" {
		fmt.Println("Empty title, question not added.")
		return
	}
	title = fmt.Sprintf("Q%d: %s", nextNum, title)

	fmt.Print("Question type (1=radio, 2=checkbox): ")
	typeChoice := strings.TrimSpace(readLine())
	qType := Radio
	if typeChoice == "2" {
		qType = Checkbox
	}

	q := Question{
		ID:       qID,
		Title:    title,
		Type:     qType,
		Required: true,
	}

	// Add options
	fmt.Println("Add options (enter empty line to stop):")
	optNum := 1
	for {
		label := readLinePrompt(fmt.Sprintf("  Option %d: ", optNum))
		label = strings.TrimSpace(label)
		if label == "" {
			break
		}
		q.Options = append(q.Options, Option{Value: label, Label: label})
		optNum++
	}

	ch.Questions = append(ch.Questions, q)
	fmt.Printf("Question %s added with %d options.\n", qID, len(q.Options))
}

func removeQuestion(ch *Chapter) {
	if len(ch.Questions) == 0 {
		fmt.Println("\nNo questions to remove.")
		return
	}

	fmt.Println("\nQuestions:")
	for i, q := range ch.Questions {
		fmt.Printf("  %d. %s\n", i+1, q.Title)
	}

	idxStr := readLinePrompt("Select question number to remove: ")
	idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
	if err != nil || idx < 1 || idx > len(ch.Questions) {
		fmt.Println("Invalid selection.")
		return
	}
	idx--

	removed := ch.Questions[idx]
	ch.Questions = append(ch.Questions[:idx], ch.Questions[idx+1:]...)

	// Re-number remaining questions
	for i := range ch.Questions {
		newID := fmt.Sprintf("q%d", i+1)
		oldID := ch.Questions[i].ID
		ch.Questions[i].ID = newID

		// Update conditional references
		for j := range ch.Questions {
			if ch.Questions[j].ConditionalOn == oldID {
				ch.Questions[j].ConditionalOn = newID
			}
		}

		// Update title prefix
		ch.Questions[i].Title = replaceQuestionPrefix(ch.Questions[i].Title, i+1)
	}

	fmt.Printf("Removed: %s\n", removed.Title)
}

func editConversation(ch *Chapter) {
	fmt.Printf("\nCurrent conversation section:\n")
	fmt.Printf("  Title: %s\n", ch.Conversation.Title)
	fmt.Printf("  Prompt: %s\n", ch.Conversation.Label)
	fmt.Printf("  Placeholder: %s\n", ch.Conversation.Placeholder)

	newTitle := readLinePrompt("New title (or Enter to keep): ")
	if newTitle != "" {
		ch.Conversation.Title = newTitle
	}

	newLabel := readLinePrompt("New prompt (or Enter to keep): ")
	if newLabel != "" {
		ch.Conversation.Label = newLabel
	}

	newPlaceholder := readLinePrompt("New placeholder (or Enter to keep): ")
	if newPlaceholder != "" {
		ch.Conversation.Placeholder = newPlaceholder
	}

	fmt.Println("Conversation section updated.")
}

func editLabels(ch *Chapter) {
	fmt.Println("\nCurrent labels:")
	fmt.Printf("  1. Email label: %s\n", ch.EmailLabel)
	fmt.Printf("  2. Email placeholder: %s\n", ch.EmailPlaceH)
	fmt.Printf("  3. Submit button: %s\n", ch.SubmitText)
	fmt.Printf("  4. Reset button: %s\n", ch.ResetText)
	fmt.Printf("  5. Success message: %s\n", ch.SuccessMsg)
	fmt.Printf("  6. Summary title: %s\n", ch.SummaryTitle)
	fmt.Printf("  7. Chapter name: %s\n", ch.ChapterName)

	choice := readLinePrompt("Select label to edit (or Enter to go back): ")
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		v := readLinePrompt("New email label: ")
		if v != "" {
			ch.EmailLabel = v
		}
	case "2":
		v := readLinePrompt("New email placeholder: ")
		if v != "" {
			ch.EmailPlaceH = v
		}
	case "3":
		v := readLinePrompt("New submit button text: ")
		if v != "" {
			ch.SubmitText = v
		}
	case "4":
		v := readLinePrompt("New reset button text: ")
		if v != "" {
			ch.ResetText = v
		}
	case "5":
		v := readLinePrompt("New success message: ")
		if v != "" {
			ch.SuccessMsg = v
		}
	case "6":
		v := readLinePrompt("New summary title: ")
		if v != "" {
			ch.SummaryTitle = v
		}
	case "7":
		v := readLinePrompt("New chapter name: ")
		if v != "" {
			ch.ChapterName = v
		}
	}
}

// ─── 4. CREATE NEW CHAPTER ───────────────────────────────────────────

func createChapter(baseDir string) {
	lang := selectLanguage()

	nextNum := getNextChapterNumber(baseDir, lang)
	fmt.Printf("\nNext available chapter number: %d\n", nextNum)
	numStr := readLinePrompt(fmt.Sprintf("Chapter number (or Enter for %d): ", nextNum))
	numStr = strings.TrimSpace(numStr)
	num := nextNum
	if numStr != "" {
		n, err := strconv.Atoi(numStr)
		if err != nil || n <= 0 {
			fmt.Println("Invalid number.")
			return
		}
		num = n
	}

	if chapterExists(baseDir, num, lang) {
		fmt.Printf("\n%s %d already exists. Use edit instead.\n\n", ChapterLabel(lang), num)
		return
	}

	// Start with defaults
	ch := LanguageDefaults(lang)
	ch.Number = num
	ch.Language = lang
	ch.Title = fmt.Sprintf("%s %d - %s", ChapterLabel(lang), num, TitleSuffix(lang))
	ch.ChapterName = fmt.Sprintf("%s %d", ChapterLabel(lang), num)

	// Ask for conversation prompt
	convLabel := readLinePrompt("Conversation/discussion prompt: ")
	if convLabel != "" {
		ch.Conversation.Label = convLabel
	}

	// Add questions interactively
	fmt.Println("\nAdd questions (type 'done' when finished):")
	qNum := 1
	for {
		fmt.Printf("\n--- Adding Question %d ---\n", qNum)
		title := readLinePrompt("Question text (or 'done' to finish): ")
		title = strings.TrimSpace(title)
		if strings.ToLower(title) == "done" || title == "" {
			break
		}

		title = fmt.Sprintf("Q%d: %s", qNum, title)

		fmt.Print("Type (1=radio, 2=checkbox): ")
		typeChoice := strings.TrimSpace(readLine())
		qType := Radio
		if typeChoice == "2" {
			qType = Checkbox
		}

		q := Question{
			ID:       fmt.Sprintf("q%d", qNum),
			Title:    title,
			Type:     qType,
			Required: true,
		}

		fmt.Println("Add options (empty line to stop):")
		optNum := 1
		for {
			label := readLinePrompt(fmt.Sprintf("  Option %d: ", optNum))
			label = strings.TrimSpace(label)
			if label == "" {
				break
			}
			q.Options = append(q.Options, Option{Value: label, Label: label})
			optNum++
		}

		// Ask about conditional visibility
		condChoice := readLinePrompt("Make this conditional on another question? (enter question id like q1, or Enter to skip): ")
		condChoice = strings.TrimSpace(condChoice)
		if condChoice != "" {
			q.ConditionalOn = condChoice
		}

		ch.Questions = append(ch.Questions, q)
		qNum++
	}

	// Preview
	fmt.Println("\n--- Preview ---")
	displayChapter(ch)

	confirm := readLinePrompt("Save this chapter? (y/n): ")
	if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
		if err := saveChapter(baseDir, ch); err != nil {
			fmt.Printf("Error saving: %v\n", err)
		} else {
			fmt.Printf("\n%s %d created successfully at %s\n\n",
				ChapterLabel(lang), num, chapterFilePath(baseDir, num, lang))
		}
	} else {
		fmt.Println("Chapter creation cancelled.")
	}
}

// ─── 5. REMOVE CHAPTER ──────────────────────────────────────────────

func removeChapter(baseDir string) {
	lang := selectLanguage()
	num := readChapterNumber()
	if num <= 0 {
		return
	}

	if !chapterExists(baseDir, num, lang) {
		fmt.Printf("\n%s %d does not exist.\n\n", ChapterLabel(lang), num)
		return
	}

	filePath := chapterFilePath(baseDir, num, lang)
	fmt.Printf("\nThis will permanently delete: %s\n", filePath)

	confirm := readLinePrompt("Are you sure? (type 'yes' to confirm): ")
	if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
		fmt.Println("Deletion cancelled.")
		return
	}

	if err := deleteChapterFile(baseDir, num, lang); err != nil {
		fmt.Printf("Error deleting: %v\n", err)
	} else {
		fmt.Printf("%s %d deleted successfully.\n\n", ChapterLabel(lang), num)
	}
}

// ─── 6. PREVIEW & SAVE ──────────────────────────────────────────────

func previewChapterMenu(baseDir string) {
	lang := selectLanguage()
	num := readChapterNumber()
	if num <= 0 {
		return
	}

	if !chapterExists(baseDir, num, lang) {
		fmt.Printf("\n%s %d does not exist.\n\n", ChapterLabel(lang), num)
		return
	}

	ch, err := parseChapterFile(chapterFilePath(baseDir, num, lang), lang)
	if err != nil {
		fmt.Printf("\nError parsing chapter: %v\n\n", err)
		return
	}

	for {
		fmt.Println("\n--- Preview & Save Menu ---")
		fmt.Println("1. Show chapter summary")
		fmt.Println("2. Show generated HTML")
		fmt.Println("3. Save (overwrite HTML file)")
		fmt.Println("4. Back to main menu")
		fmt.Print("Choice: ")

		choice := strings.TrimSpace(readLine())
		switch choice {
		case "1":
			displayChapter(ch)
		case "2":
			html := previewChapter(ch)
			fmt.Println("\n--- Generated HTML ---")
			fmt.Println(html)
			fmt.Println("--- End of HTML ---")
		case "3":
			filePath := chapterFilePath(baseDir, num, lang)
			fmt.Printf("This will overwrite: %s\n", filePath)
			confirm := readLinePrompt("Confirm save? (y/n): ")
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				if err := saveChapter(baseDir, ch); err != nil {
					fmt.Printf("Error saving: %v\n", err)
				} else {
					fmt.Printf("%s saved successfully.\n", ch.ChapterName)
				}
			}
		case "4":
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}

// ─── HELPERS ────────────────────────────────────────────────────────

func readChapterNumber() int {
	numStr := readLinePrompt("Enter chapter number: ")
	numStr = strings.TrimSpace(numStr)
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		fmt.Println("Invalid chapter number.")
		return 0
	}
	return num
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func replaceQuestionPrefix(title string, newNum int) string {
	// Replace Q<old>: with Q<new>:
	idx := strings.Index(title, ":")
	if idx == -1 {
		return title
	}
	prefix := title[:idx]
	if strings.HasPrefix(prefix, "Q") {
		return fmt.Sprintf("Q%d%s", newNum, title[idx:])
	}
	return title
}
