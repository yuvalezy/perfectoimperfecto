package main

// Language represents the supported languages
type Language string

const (
	Spanish Language = "es"
	English Language = "en"
)

// QuestionType represents the type of input for a question
type QuestionType string

const (
	Radio    QuestionType = "radio"
	Checkbox QuestionType = "checkbox"
)

// Option represents a single answer option for a question
type Option struct {
	Value string
	Label string
}

// Question represents a survey question with its options
type Question struct {
	ID            string       // e.g. "q1", "q2"
	Title         string       // The question text
	Type          QuestionType // radio or checkbox
	Options       []Option
	Required      bool
	ConditionalOn string // If set, this question only shows when the referenced question has value "Si"/"Yes"
}

// ConversationSection represents the discussion/conversation prompt
type ConversationSection struct {
	Title       string // e.g. "Conversen:" or "Discuss:"
	Label       string // The prompt text
	Placeholder string
}

// Chapter represents a complete survey chapter
type Chapter struct {
	Number       int
	Language     Language
	Title        string // HTML <title> content
	Heading      string // <h1> content
	Questions    []Question
	Conversation ConversationSection
	EmailLabel   string // "Correo Electrónico" or "Email Address"
	EmailPlaceH  string // email placeholder
	SubmitText   string // Submit button text
	ResetText    string // Reset button text
	SuccessMsg   string // Success message text
	SummaryTitle string // Summary section title
	ChapterName  string // e.g. "Capítulo 1" or "Chapter 1"
	HasCustomCSS bool   // Whether the chapter has inline <style> (e.g. checkbox chapters)
	HasCustomJS  bool   // Whether the chapter has conditional show/hide logic
}

// LanguageDefaults returns default strings for a given language
func LanguageDefaults(lang Language) Chapter {
	if lang == English {
		return Chapter{
			Language:     English,
			Heading:      "After watching the video, complete this exercise:",
			EmailLabel:   "Email Address",
			EmailPlaceH:  "example@email.com",
			SubmitText:   "Submit Survey",
			ResetText:    "Clear",
			SuccessMsg:   "✓ Survey submitted successfully!",
			SummaryTitle: "Summary of your answers",
			Conversation: ConversationSection{
				Title:       "Discuss:",
				Placeholder: "Write your answer here...",
			},
		}
	}
	return Chapter{
		Language:     Spanish,
		Heading:      "Después de ver el video, realiza este ejercicio:",
		EmailLabel:   "Correo Electrónico",
		EmailPlaceH:  "ejemplo@correo.com",
		SubmitText:   "Enviar Encuesta",
		ResetText:    "Limpiar",
		SuccessMsg:   "✓ ¡Encuesta enviada correctamente!",
		SummaryTitle: "Resumen de tus respuestas",
		Conversation: ConversationSection{
			Title:       "Conversen:",
			Placeholder: "Escribe tu respuesta aquí...",
		},
	}
}

// FilePrefix returns the filename prefix for a language
func FilePrefix(lang Language) string {
	if lang == English {
		return "chapter"
	}
	return "capitulo"
}

// ChapterLabel returns the chapter label for a language
func ChapterLabel(lang Language) string {
	if lang == English {
		return "Chapter"
	}
	return "Capítulo"
}

// TitleSuffix returns the title suffix for a language
func TitleSuffix(lang Language) string {
	if lang == English {
		return "After watching the video"
	}
	return "Después de ver el video"
}
