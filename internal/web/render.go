package web

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"path/filepath"
)

// templatesDir — папка с шаблонами (относительно корня проекта).
const templatesDir = "web/templates"

// Renderer — рендерит HTML-страницы через layout + content.
type Renderer struct {
	logger *slog.Logger
}

func NewRenderer(logger *slog.Logger) *Renderer {
	return &Renderer{logger: logger}
}

// RenderPage — рендерит страницу.
//
// layoutFile — например, "layout.html".
// pageFile   — например, "login.html".
// data       — данные для шаблона.
//
// Как работает: ParseFiles объединяет layout и content в одну коллекцию,
// затем ExecuteTemplate рендерит "layout.html", который через {{template "content" .}}
// вставляет содержимое page-файла.
func (r *Renderer) RenderPage(w io.Writer, layoutFile, pageFile string, data any) error {
	layoutPath := filepath.Join(templatesDir, layoutFile)
	pagePath := filepath.Join(templatesDir, pageFile)

	tmpl, err := template.ParseFiles(layoutPath, pagePath)
	if err != nil {
		r.logger.Error("failed to parse templates", "layout", layoutFile, "page", pageFile, "error", err)
		return fmt.Errorf("parse templates: %w", err)
	}

	if err := tmpl.ExecuteTemplate(w, layoutFile, data); err != nil {
		r.logger.Error("failed to execute template", "layout", layoutFile, "error", err)
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
