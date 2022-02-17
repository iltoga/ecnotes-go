package ui

type WindowInterface interface {
	CreateWindow(title string, width, height float32, visible bool, options map[string]interface{})
	ParseDefaultOptions(options map[string]interface{})
}
