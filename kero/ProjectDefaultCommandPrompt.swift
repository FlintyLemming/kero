//
//  ProjectDefaultCommandPrompt.swift
//  kero
//

import AppKit

/// Edits a project's default command ("Default Command…" on the project
/// row): the line its new sessions run in the user's shell before the shell
/// itself. An empty line clears it.
@MainActor
enum ProjectDefaultCommandPrompt {
    static func edit(_ project: Project) {
        let alert = NSAlert()
        alert.messageText = String(
            localized: "Default Command for “\(project.name)”",
            comment: "Title of the project default command dialog. The placeholder is a project name."
        )
        alert.informativeText = String(
            localized: "New sessions in this project run it in your shell, then return to the shell when it ends. Leave it empty to start at the shell.",
            comment: "Explanation in the project default command dialog."
        )

        let field = NSTextField(string: project.defaultCommand ?? "")
        field.placeholderString = String(
            localized: "Command, as claude or npm run dev",
            comment: "Placeholder of the project default command field."
        )
        field.usesSingleLineMode = true
        field.cell?.isScrollable = true
        field.frame = NSRect(x: 0, y: 0, width: 320, height: 24)
        alert.accessoryView = field

        let save = alert.addButton(withTitle: String(localized: "Save"))
        save.keyEquivalent = "\r"
        let cancel = alert.addButton(withTitle: String(localized: "Cancel"))
        cancel.keyEquivalent = "\u{1b}"
        alert.window.initialFirstResponder = field

        let apply: (NSApplication.ModalResponse) -> Void = { response in
            guard response == .alertFirstButtonReturn else { return }
            project.defaultCommand = normalized(field.stringValue)
        }
        if let window = NSApp.keyWindow ?? NSApp.mainWindow {
            alert.beginSheetModal(for: window, completionHandler: apply)
        } else {
            apply(alert.runModal())
        }
    }

    /// One line, as the launch script needs it; nil for none.
    static func normalized(_ line: String) -> String? {
        let oneLine = line
            .components(separatedBy: .newlines)
            .joined(separator: " ")
            .trimmingCharacters(in: .whitespaces)
        return oneLine.isEmpty ? nil : oneLine
    }
}
