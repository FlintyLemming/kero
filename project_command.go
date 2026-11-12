package main

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"kero/internal/proto"

	"kero/locales"
)

// The project's default command lives in its own files, as a fork's
// feature, so that rebasing on upstream touches few of upstream's lines.

// projectCommandAsk edits the line a project's new sessions run in their
// shell (proto.Project.Command).
type projectCommandAsk struct {
	open bool
	line string
	host *host
	id   string
	name string
}

// askProjectCommand opens the dialog of a project's default command.
func (a *App) askProjectCommand(h *host, id string) {
	pr := a.project(id)
	if pr == nil {
		return
	}
	if h.cl == nil {
		a.toast(locales.HostNotConnected(h.label()))
		return
	}
	a.cmdAsk = &projectCommandAsk{open: true, line: pr.Command, host: h, id: id, name: pr.Name}
}

func (a *App) projectCommandView(c *ui.Context, pal *palette) {
	w := a.cmdAsk
	if w == nil {
		return
	}
	save := false
	ui.DialogBase(c.Key("project-command"), &w.open, func(backdrop, panel ui.Element) {
		backdrop.Background(pal.overlay)
		title := locales.ProjectCommandTitle(w.name)
		floating(panel.Width(440).Padding(22).Gap(8).Radius(14).Background(pal.popover), pal).Label(title)
		ui.Text(c, title).FontSize(15).FontWeight(650).TextColor(pal.text)
		ui.Text(c, locales.ProjectCommandDetail()).FontSize(13).TextColor(pal.textMuted).LineHeight(1.4)
		ui.Row(c).Margin(8, 0, 0, 0).Children(func() {
			in := textField(c, "command", &w.line, locales.ProjectCommandPlaceholder(), 0, pal)
			in.AutoFocus()
			if in.Submitted() {
				save = true
			}
		})
		ui.Row(c).Gap(8).Justify(ui.End).Margin(14, 0, 0, 0).Children(func() {
			if button(c, "cancel", locales.CommonCancel(), "esc", false, pal) {
				w.open = false
			}
			if button(c, "ok", locales.CommonSave(), "↩", true, pal) {
				save = true
			}
		})
	})
	if save {
		w.open = false
		a.setProjectCommand(w.host, w.id, strings.TrimSpace(w.line))
	}
	if !w.open {
		a.cmdAsk = nil
		a.focusTerminal()
	}
}

// setProjectCommand stores a project's default command in its host's
// kerod, which keeps it with the project. A kerod of upstream's build
// knows no project.command, which the toast of its error tells.
func (a *App) setProjectCommand(h *host, id, line string) {
	pr := a.project(id)
	if pr == nil || pr.Command == line {
		return
	}
	pr.Command = line // at once; the daemon's state follows
	a.callOn(h, "project.command", proto.ProjectParams{ID: id, Command: line}, nil)
}
