package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// A project's menu edits its default command: the dialog shows the one it
// has, Return keeps the new one at once and tells kerod (project.command),
// Escape leaves it be; and the user's new sessions ask kerod to run it.
func TestProjectCommandDialog(t *testing.T) {
	a := testApp(t)
	a.st.Projects[0].Name = "webapp" // not the session's name, to click
	calls := make(chan [2]string, 8)
	cl, stop := projectClient(t, func(method string, raw json.RawMessage) (any, error) {
		calls <- [2]string{method, string(raw)}
		return true, nil
	})
	defer stop()
	call := func() [2]string {
		select {
		case c := <-calls:
			return c
		case <-time.After(time.Second):
			return [2]string{}
		}
	}
	tt := ui.NewTester(a.view, 1280, 720)
	tt.Frame()
	a.local().cl = cl

	if err := tt.RightClick("webapp"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Default Command…"); err != nil {
		t.Fatalf("%v; the menu is %q", err, tt.Menu())
	}
	tt.Frame()
	if a.cmdAsk == nil || !tt.HasText("Default Command for “webapp”") {
		t.Fatalf("no dialog; texts %q", tt.Texts())
	}
	tt.Type("claude --continue")
	tt.Frame()
	shot(t, tt, "project-command")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.cmdAsk != nil {
		t.Error("Return left the dialog open")
	}
	if got := a.st.Projects[0].Command; got != "claude --continue" {
		t.Errorf("the project's command is %q", got)
	}
	if c := call(); c[0] != "project.command" || c[1] != `{"id":"p1","command":"claude --continue"}` {
		t.Errorf("saving called %q", c)
	}

	// Again, it shows the command it has; Escape keeps it.
	a.askProjectCommand(a.local(), "p1")
	tt.Frame()
	if a.cmdAsk == nil || a.cmdAsk.line != "claude --continue" {
		t.Fatalf("the dialog opened with %+v", a.cmdAsk)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.cmdAsk != nil || a.st.Projects[0].Command != "claude --continue" {
		t.Errorf("Escape: dialog %+v, command %q", a.cmdAsk, a.st.Projects[0].Command)
	}

	// A new session asks for the project's command.
	a.newSession("p1")
	if c := call(); c[0] != "create" || !json.Valid([]byte(c[1])) || !strings.Contains(c[1], `"projectCommand":true`) {
		t.Errorf("a new session called %q", c)
	}
}
