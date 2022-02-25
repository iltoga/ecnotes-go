package ui

import "fyne.io/fyne/v2"

type WindowInterface interface {
	CreateWindow(title string, width, height float32, visible bool, options map[string]interface{})
	ParseDefaultOptions(options map[string]interface{})
	GetWindow() fyne.Window
}
