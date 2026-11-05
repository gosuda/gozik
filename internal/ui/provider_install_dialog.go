package ui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gg582/gozik/internal/plugininstaller"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func (mw *MainWindow) openPluginManager(parent *gtk.Dialog) {
	dialog, err := gtk.DialogNewWithButtons("Music Plugins", parent, gtk.DIALOG_MODAL, []interface{}{"Close", gtk.RESPONSE_CLOSE})
	if err != nil {
		return
	}
	defer dialog.Destroy()
	dialog.SetDefaultSize(650, 520)
	content, _ := dialog.GetContentArea()
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)
	content.PackStart(box, true, true, 0)

	intro, _ := gtk.LabelNew("Install the official prebuilt plugins. Spotify includes yt-dlp; YouTube Music includes Python and Node.js. No compiler, Xcode, or developer toolkit is required.")
	intro.SetLineWrap(true)
	intro.SetHAlign(gtk.ALIGN_START)
	box.PackStart(intro, false, false, 0)

	logView, _ := gtk.TextViewNew()
	logView.SetEditable(false)
	logView.SetCursorVisible(false)
	logBuffer, _ := logView.GetBuffer()
	scrolled, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolled.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrolled.SetSizeRequest(-1, 100)
	scrolled.Add(logView)
	appendLog := func(message string) {
		glib.IdleAdd(func() bool {
			end := logBuffer.GetEndIter()
			logBuffer.Insert(end, strings.TrimSpace(message)+"\n")
			return false
		})
	}

	for _, plugin := range plugininstaller.Plugins {
		mw.addPluginRow(box, dialog, plugin, appendLog)
	}
	box.PackStart(scrolled, true, true, 0)

	guide, _ := gtk.LabelNew("After installation, click Configure. Spotify requires your own Client ID and the redirect URI http://127.0.0.1:43827/callback. YouTube Music supports OAuth device login and browser-cookie login.")
	guide.SetLineWrap(true)
	guide.SetHAlign(gtk.ALIGN_START)
	guide.SetSelectable(true)
	box.PackStart(guide, false, false, 0)
	box.ShowAll()
	dialog.Run()
}

func (mw *MainWindow) addPluginRow(parent *gtk.Box, dialog *gtk.Dialog, plugin plugininstaller.Plugin, appendLog func(string)) {
	frame, _ := gtk.FrameNew(plugin.Name)
	row, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	row.SetMarginTop(8)
	row.SetMarginBottom(8)
	row.SetMarginStart(8)
	row.SetMarginEnd(8)
	frame.Add(row)
	description, _ := gtk.LabelNew(plugin.Description)
	description.SetLineWrap(true)
	description.SetHAlign(gtk.ALIGN_START)
	row.PackStart(description, false, false, 0)

	actions, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	status, _ := gtk.LabelNew("")
	status.SetHAlign(gtk.ALIGN_START)
	installed := mw.pluginRunning(plugin.ID)
	if installed {
		status.SetText("Running")
	} else {
		status.SetText("Not detected")
	}
	installButton, _ := gtk.ButtonNewWithLabel("Install / Update")
	configureButton, _ := gtk.ButtonNewWithLabel("Configure")
	configureButton.SetSensitive(installed)
	sourceButton, _ := gtk.ButtonNewWithLabel("Project / Manual Install")
	actions.PackStart(status, true, true, 0)
	actions.PackEnd(sourceButton, false, false, 0)
	actions.PackEnd(configureButton, false, false, 0)
	actions.PackEnd(installButton, false, false, 0)
	row.PackStart(actions, false, false, 0)
	parent.PackStart(frame, false, false, 0)

	sourceButton.Connect("clicked", func() { openPluginURL(plugin.Repository) })
	configureButton.Connect("clicked", func() { openPluginURL(plugin.WebUI) })
	installButton.Connect("clicked", func() {
		installButton.SetSensitive(false)
		configureButton.SetSensitive(false)
		status.SetText("Installing...")
		appendLog("Starting " + plugin.Name + " installation. Follow any system permission prompts.")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			err := plugininstaller.Install(ctx, plugin.ID, appendLog)
			glib.IdleAdd(func() bool {
				installButton.SetSensitive(true)
				if err != nil {
					status.SetText("Installation failed")
					appendLog(fmt.Sprintf("Error: %v\nUse Project / Manual Install if no package is available for this system.", err))
					return false
				}
				status.SetText("Installed; waiting for service")
				appendLog(plugin.Name + " installed successfully.")
				go func() {
					time.Sleep(3 * time.Second)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					mw.providerMgr.DiscoverServices(ctx)
					cancel()
					glib.IdleAdd(func() bool {
						if mw.pluginRunning(plugin.ID) {
							status.SetText("Running")
							configureButton.SetSensitive(true)
						} else {
							status.SetText("Installed; restart Gozik if not detected")
						}
						return false
					})
				}()
				return false
			})
		}()
	})
}

func (mw *MainWindow) pluginRunning(id string) bool {
	for _, item := range mw.providerMgr.Providers() {
		if item.ProviderID == id || (id == "ytmusic" && item.ProviderID == "yt-music") {
			return true
		}
	}
	return false
}

func openPluginURL(url string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	_ = command.Start()
}
