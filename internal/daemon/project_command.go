package daemon

import (
	"fmt"
	"strings"

	"kero/internal/proto"
)

// The project's default command lives in its own file, as a fork's
// feature, so that rebasing on upstream touches few of upstream's lines.

// setProjectCommand sets the line a project's new sessions run
// (project.command); "" leaves them at the shell.
func (d *Daemon) setProjectCommand(p proto.ProjectParams) error {
	line := strings.TrimSpace(p.Command)
	found := false
	d.mu.Lock()
	for i := range d.projects {
		if d.projects[i].ID == p.ID {
			d.projects[i].Command, found = line, true
		}
	}
	d.mu.Unlock()
	if !found {
		return fmt.Errorf("no project %q", p.ID)
	}
	d.changed()
	return nil
}

// sessionCommandLocked is the command a new session of a project starts
// with: its own, else, when the app asks (ProjectCommand), the project's.
// A browser runs none. d.mu is held.
func (d *Daemon) sessionCommandLocked(project string, p proto.CreateParams) []string {
	if len(p.Command) > 0 || !p.ProjectCommand || p.URL != "" {
		return p.Command
	}
	for _, pr := range d.projects {
		if pr.ID == project && pr.Command != "" {
			// One line, which the shell runs as typed: pipes, && and
			// aliases work as they do at its prompt (shellJoin).
			return []string{pr.Command}
		}
	}
	return nil
}
