package ui

import (
	"errors"

	catalogeditor "mcpv/internal/infra/catalog/editor"
)

func mapCatalogError(err error) error {
	if err == nil {
		return nil
	}
	var editorErr *catalogeditor.Error
	if errors.As(err, &editorErr) {
		detail := ""
		if editorErr.Err != nil {
			detail = editorErr.Err.Error()
		}
		switch editorErr.Kind {
		case catalogeditor.ErrorInvalidRequest:
			return NewErrorWithDetails(ErrCodeInvalidRequest, editorErr.Message, detail)
		case catalogeditor.ErrorInvalidConfig:
			return NewErrorWithDetails(ErrCodeInvalidConfig, editorErr.Message, detail)
		default:
			return NewErrorWithDetails(ErrCodeInvalidConfig, editorErr.Message, detail)
		}
	}
	return NewErrorWithDetails(ErrCodeInvalidConfig, "Failed to update configuration", err.Error())
}
