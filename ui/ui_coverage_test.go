package ui

import (
	"net/url"
	"sync"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeApp struct {
	runCalled        bool
	quitCalled       bool
	notifications    []*fyne.Notification
}

func (f *fakeApp) NewWindow(title string) fyne.Window { return nil }
func (f *fakeApp) OpenURL(_ *url.URL) error           { return nil }
func (f *fakeApp) Icon() fyne.Resource                { return nil }
func (f *fakeApp) SetIcon(fyne.Resource)              {}
func (f *fakeApp) Run()                               { f.runCalled = true }
func (f *fakeApp) Quit()                              { f.quitCalled = true }
func (f *fakeApp) Driver() fyne.Driver                { return nil }
func (f *fakeApp) UniqueID() string                   { return "fake-app" }
func (f *fakeApp) SendNotification(n *fyne.Notification) {
	f.notifications = append(f.notifications, n)
}
func (f *fakeApp) Settings() fyne.Settings                 { return nil }
func (f *fakeApp) Preferences() fyne.Preferences           { return nil }
func (f *fakeApp) Storage() fyne.Storage                   { return nil }
func (f *fakeApp) Lifecycle() fyne.Lifecycle               { return nil }
func (f *fakeApp) Metadata() fyne.AppMetadata              { return fyne.AppMetadata{} }
func (f *fakeApp) CloudProvider() fyne.CloudProvider       { return nil }
func (f *fakeApp) SetCloudProvider(fyne.CloudProvider)     {}

var _ fyne.App = (*fakeApp)(nil)

func TestWindowDefaultOptions_ParseDefaultOptions(t *testing.T) {
	var opts WindowDefaultOptions

	opts.ParseDefaultOptions(nil)
	assert.Equal(t, common.WindowAction_New, opts.windowAction)
	assert.Equal(t, common.WindowMode_View, opts.windowMode)
	assert.Equal(t, common.WindowAspect(0), opts.windowAspect)

	opts.ParseDefaultOptions(map[string]interface{}{
		common.OPT_WINDOW_ACTION: common.WindowAction_Update,
		common.OPT_WINDOW_MODE:   common.WindowMode_Edit,
		common.OPT_WINDOW_ASPECT:  common.WindowAspect_FullScreen,
	})
	assert.Equal(t, common.WindowAction_Update, opts.windowAction)
	assert.Equal(t, common.WindowMode_Edit, opts.windowMode)
	assert.Equal(t, common.WindowAspect_FullScreen, opts.windowAspect)

	opts.ParseDefaultOptions(map[string]interface{}{
		common.OPT_WINDOW_ASPECT: "wrong-type",
	})
	assert.Equal(t, common.WindowAspect_Normal, opts.windowAspect)
}

func TestNoteDetailsWindowImpl_ParseDefaultOptions(t *testing.T) {
	win := &NoteDetailsWindowImpl{}

	win.ParseDefaultOptions(map[string]interface{}{
		common.OPT_WINDOW_ACTION: common.WindowAction_Delete,
		common.OPT_WINDOW_MODE:   common.WindowMode_Edit,
		common.OPT_WINDOW_ASPECT:  common.WindowAspect_FullScreen,
	})

	assert.Equal(t, common.WindowAction_Delete, win.windowAction)
	assert.Equal(t, common.WindowMode_Edit, win.windowMode)
	assert.Equal(t, common.WindowAspect_FullScreen, win.windowAspect)
}

func TestUImpl_WindowAndWidgetHelpers(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	ui := NewUI(app, nil, nil, nil, nil, &observer.ObserverImpl{})
	win := app.NewWindow("test")
	entry := widget.NewEntry()
	label := widget.NewLabel("read only")

	ui.AddWindow("main", win)
	ui.AddWidget("entry", entry)
	ui.AddWidget("label", label)

	gotWin, err := ui.GetWindow("main")
	require.NoError(t, err)
	assert.Same(t, win, gotWin)

	gotWidget, err := ui.GetWidget("entry")
	require.NoError(t, err)
	assert.Same(t, entry, gotWidget)

	require.NoError(t, ui.SetWidgetEnabled("entry", false))
	require.NoError(t, ui.SetWidgetEnabled("entry", true))
	require.Error(t, ui.SetWidgetEnabled("label", true))
	require.Error(t, ui.SetWidgetEnabled("missing", true))

	require.NoError(t, ui.SetWidgetVisibility("entry", false))
	require.NoError(t, ui.SetWidgetVisibility("entry", true))
	require.Error(t, ui.SetWidgetVisibility("missing", true))

	require.NoError(t, ui.SetWindowVisibility("main", false))
	require.NoError(t, ui.SetWindowVisibility("main", true))
	require.Error(t, ui.SetWindowVisibility("missing", true))

	initialFullScreen := win.FullScreen()
	ui.ToggleFullScreen(win)
	assert.NotEqual(t, initialFullScreen, win.FullScreen())

	ui.SetFocusOnWidget(win, entry)
}

func TestUImpl_RunStopAndShowNotification(t *testing.T) {
	app := &fakeApp{}
	ui := NewUI(app, nil, nil, nil, nil, &observer.ObserverImpl{})

	ui.Run()
	assert.True(t, app.runCalled)

	ui.Stop()
	assert.True(t, app.quitCalled)

	ui.ShowNotification("Title", "Content")
	require.Len(t, app.notifications, 1)
	assert.Equal(t, "Title", app.notifications[0].Title)
	assert.Equal(t, "Content", app.notifications[0].Content)
}

func TestUImpl_Getters(t *testing.T) {
	obs := &observer.ObserverImpl{}
	ui := NewUI(&fakeApp{}, nil, nil, nil, nil, obs)

	assert.Nil(t, ui.GetNoteService())
	assert.Nil(t, ui.GetKeyService())
	assert.Same(t, obs, ui.GetObserver())
}

func TestUImpl_GetWindowAndWidgetMissing(t *testing.T) {
	ui := NewUI(&fakeApp{}, nil, nil, nil, nil, &observer.ObserverImpl{})

	_, err := ui.GetWindow("missing")
	require.Error(t, err)
	_, err = ui.GetWidget("missing")
	require.Error(t, err)
}

func TestUImpl_AddWindowAndWidgetAreThreadSafeEnoughForSimpleUse(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	ui := NewUI(app, nil, nil, nil, nil, &observer.ObserverImpl{})
	win := app.NewWindow("secondary")
	entry := widget.NewEntry()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ui.AddWindow("secondary", win)
	}()
	go func() {
		defer wg.Done()
		ui.AddWidget("secondary-entry", entry)
	}()
	wg.Wait()

	gotWin, err := ui.GetWindow("secondary")
	require.NoError(t, err)
	assert.Same(t, win, gotWin)
	gotWidget, err := ui.GetWidget("secondary-entry")
	require.NoError(t, err)
	assert.Same(t, entry, gotWidget)
}
