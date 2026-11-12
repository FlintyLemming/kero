package daemon

import (
	"encoding/json"
	"slices"
	"testing"

	"kero/internal/proto"
)

func setCommand(t *testing.T, d *Daemon, id, line string) error {
	t.Helper()
	raw, err := json.Marshal(proto.ProjectParams{ID: id, Command: line})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.call(nil, "project.command", raw)
	return err
}

// project.command keeps a project's line, trimmed, and "" clears it; a
// project that isn't there is an error.
func TestProjectCommandSet(t *testing.T) {
	d := testDaemon(t)
	if err := setCommand(t, d, "pr", "  claude --continue \n"); err != nil {
		t.Fatal(err)
	}
	if got := d.projects[0].Command; got != "claude --continue" {
		t.Fatalf("command %q", got)
	}
	if err := setCommand(t, d, "pr", ""); err != nil {
		t.Fatal(err)
	}
	if got := d.projects[0].Command; got != "" {
		t.Fatalf("cleared command %q", got)
	}
	if err := setCommand(t, d, "gone", "claude"); err == nil {
		t.Fatal("set the command of no project")
	}
}

// A new session runs its project's command only when the app asks for
// it, and never in place of a command of its own or in a browser.
func TestProjectCommandStartsSessions(t *testing.T) {
	d := testDaemon(t)
	d.projects[0].Command = "echo hi && echo there"
	cases := []struct {
		name string
		p    proto.CreateParams
		want []string
	}{
		{"app's new session", proto.CreateParams{Project: "pr", ProjectCommand: true}, []string{"echo hi && echo there"}},
		{"kero new", proto.CreateParams{Project: "pr"}, nil},
		{"own command", proto.CreateParams{Project: "pr", ProjectCommand: true, Command: []string{"top"}}, []string{"top"}},
		{"browser", proto.CreateParams{Project: "pr", ProjectCommand: true, URL: "localhost:3000"}, nil},
	}
	for _, tc := range cases {
		d.mu.Lock()
		got := d.sessionCommandLocked("pr", tc.p)
		d.mu.Unlock()
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: command %q, want %q", tc.name, got, tc.want)
		}
	}

	// And a real one starts with it: what the pane tells it ran.
	ses, err := d.create(proto.CreateParams{Project: "pr", ProjectCommand: true})
	if err != nil {
		t.Fatal(err)
	}
	defer d.removeSession(ses.ID)
	if got := ses.Panes[0].Command; !slices.Equal(got, []string{"echo hi && echo there"}) {
		t.Fatalf("session's command %q", got)
	}
}

// A project with no command leaves its sessions at the shell, asked or not.
func TestProjectCommandNone(t *testing.T) {
	d := testDaemon(t)
	d.mu.Lock()
	got := d.sessionCommandLocked("pr", proto.CreateParams{Project: "pr", ProjectCommand: true})
	d.mu.Unlock()
	if got != nil {
		t.Fatalf("command %q", got)
	}
}
