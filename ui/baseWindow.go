package ui

import (
	"github.com/iltoga/ecnotes-go/lib/common"
)

// WindowDefaultOptions default options for windows
type WindowDefaultOptions struct {
	windowAction  common.WindowAction
	windowMode    common.WindowMode
	windowAspect  common.WindowAspect
	windowVisible bool
}

// ParseDefaultOptions ...
func (ui *WindowDefaultOptions) ParseDefaultOptions(options map[string]interface{}) {
	if val := common.GetMapValOrNil(options, common.OPT_WINDOW_ACTION); val != nil {
		if mode, ok := val.(common.WindowAction); ok {
			ui.windowAction = mode
		}
	} else {
		// default action is new
		ui.windowAction = common.WindowAction_New
	}
	if val := common.GetMapValOrNil(options, common.OPT_WINDOW_MODE); val != nil {
		if mode, ok := val.(common.WindowMode); ok {
			ui.windowMode = mode
		}
	} else {
		// default mode is view
		ui.windowMode = common.WindowMode_View
	}
	if val := common.GetMapValOrNil(options, common.OPT_WINDOW_ASPECT); val != nil {
		if aspect, ok := val.(common.WindowAspect); ok {
			ui.windowAspect = aspect
		} else {
			// default aspect is normal
			ui.windowAspect = common.WindowAspect_Normal
		}
	}
}
