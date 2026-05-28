package tools

import (
	"accelerator/internal/domains"
	"encoding/json"
	"fmt"
	"strings"
)

// дополнительная функция, валидирует additionnally prompt

// AdditionalPromptItem – один элемент дополнительного промпта.
type AdditionalPromptItem struct {
	Title  string `json:"title"`
	Prompt string `json:"prompt"`
}

/*
дополнительный промпт должен иметь следующую структуру:
		[
			{
				"title": "Название секции 1",
				"prompt": "Секция 1 текст"
			}
		]

*/

// ValidateAdditionalPrompt проверяет, что raw является корректным JSON-массивом,
// каждый элемент которого содержит непустой title и поле prompt (может быть пустым).
// Возвращает ошибку, если структура не соответствует ожиданиям.
func ValidateAdditionalPrompt(raw json.RawMessage) error {
	if len(raw) == 0 {
		return fmt.Errorf("additional_prompt не может быть пустым")
	}

	var items []AdditionalPromptItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("additional_prompt должен быть массивом объектов с полями title и prompt")
	}

	if len(items) == 0 {
		return fmt.Errorf("additional_prompt не должен быть пустым массивом")
	}

	for i, item := range items {
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("элемент %d: поле title обязательно и не может быть пустым", i)
		}
		// prompt может быть любым (пустым или нет), дополнительных проверок не требуется
	}

	return nil
}

// BuildFullPrompt собирает итоговый промпт для нейросети.

/*
ПРИМЕР

[
	{
		"title": "Название секции 1",
		"prompt": "Секция 1 текст"
	}
]


РЕЗУЛЬТАТ

Основной текст задания...

Дополнительно:
Укажи сроки выполнения

Пожелания:
Избегай технического жаргона
*/
// Основной промпт (summary_prompt) идёт первым, затем форматированные дополнительные инструкции.
// Предполагается, что AdditionalPrompt уже прошёл валидацию (через ValidateAdditionalPrompt).
func BuildFullPrompt(prompts *domains.TaskPatternPrompts) string {
	var parts []string

	// Основной промпт
	if strings.TrimSpace(prompts.Prompt) != "" {
		parts = append(parts, prompts.Prompt)
	}

	// Дополнительные инструкции
	if len(prompts.AdditionalPrompt) > 0 {
		var items []AdditionalPromptItem
		// Ошибка игнорируется, т.к. данные уже проверены валидатором
		if err := json.Unmarshal(prompts.AdditionalPrompt, &items); err == nil {
			formatted := FormatAdditionalPrompt(items)
			if formatted != "" {
				parts = append(parts, formatted)
			}
		}
	}

	return strings.Join(parts, "\n\n")
}

// Она принимает на вход []AdditionalPromptItem и возвращает строку, где каждый элемент оформлен так:
// Заголовок:
// текст prompt
func FormatAdditionalPrompt(items []AdditionalPromptItem) string {
	if len(items) == 0 {
		return ""
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		prompt := strings.TrimSpace(item.Prompt)
		parts = append(parts, fmt.Sprintf("%s:\n%s", title, prompt))
	}
	return strings.Join(parts, "\n\n")
}
