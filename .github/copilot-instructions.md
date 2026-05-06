# Repository AI Instructions

- Respond in Korean by default.
- Use verified repository evidence before making claims about code, files, commands, tests, or Git state.
- Do not invent tickets, validation results, logs, command output, or business context.
- For commit message generation, follow [git-commit-instructions.md](./git-commit-instructions.md).
- Before generating a commit message, inspect the selected commit scope or staged diff first.
- If the current branch contains a JIRA key, include it as the commit message `Ref:`.
- If the commit scope is unclear, suggest logical commit groups instead of forcing one message.
- Keep commit-message rules isolated to commit-message tasks; do not apply them to unrelated code generation or review tasks.
