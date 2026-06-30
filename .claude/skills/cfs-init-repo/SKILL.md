---
name: cfs-init-repo
description: Initialize a repository for Code from Spec. Creates the spec directory, downloads tooling, configures the MCP server, and installs the skill and subagent definitions. Run once per project.
---

# Initialize Repository for Code from Spec

One-time setup of a repository to use Code from Spec.

## When invoked

Run this skill when the user asks to initialize Code from
Spec in a project, or when starting a new project that
will use the methodology.

## Algorithm

Each step checks if its target already exists and skips
if so. This makes the skill safe to re-run — it fills
in whatever is missing without overwriting what is
already in place.

1. **Create the spec directory.** If `code-from-spec/`
   does not exist, create it.

2. **Download the methodology file.** Download
   `CODE_FROM_SPEC.md` from
   `https://raw.githubusercontent.com/CodeFromSpec/framework/main/CODE_FROM_SPEC.md`
   and save it to `code-from-spec/.rules/CODE_FROM_SPEC.md`.
   Create the directory if needed.

3. **Download the MCP server.** Detect the platform
   (OS + architecture) and download the appropriate
   `framework-mcp` binary from
   `https://github.com/CodeFromSpec/tool-framework-mcp/releases/latest`
   into `code-from-spec/.tools/`. On Windows, the binary is
   `framework-mcp.exe`.

4. **Configure .gitignore.** Add the following entries
   to `.gitignore`. Create the file if it does not exist.
   Do not duplicate entries that already exist.

   - `/code-from-spec/.tools/`
   - `/code-from-spec/.cache/`
   - `/code-from-spec/.manifest.lock`

5. **Configure the MCP server.** Create or update
   `.mcp.json` in the project root:

   ```json
   {
     "mcpServers": {
       "framework-mcp": {
         "type": "stdio",
         "command": "code-from-spec/.tools/framework-mcp"
       }
     }
   }
   ```

   On Windows, use `code-from-spec/.tools/framework-mcp.exe`
   as the command. If `.mcp.json` already exists and has
   other servers, merge — do not overwrite.

6. **Install subagent definitions.** Download and save
   to `.claude/agents/`. Create the directory if needed.

   - `cfs-artifact-generation` from
     `https://raw.githubusercontent.com/CodeFromSpec/framework/main/subagents/cfs-artifact-generation.md`

7. **Install skills.** Download the following skills and
   save them to `.claude/skills/<name>/SKILL.md`:

   - `cfs-generate` from
     `https://raw.githubusercontent.com/CodeFromSpec/framework/main/skills/cfs-generate/SKILL.md`
   - `cfs-status` from
     `https://raw.githubusercontent.com/CodeFromSpec/framework/main/skills/cfs-status/SKILL.md`
   - `cfs-check-meta-language` from
     `https://raw.githubusercontent.com/CodeFromSpec/framework/main/skills/cfs-check-meta-language/SKILL.md`
   - `cfs-init-session` from
     `https://raw.githubusercontent.com/CodeFromSpec/framework/main/skills/cfs-init-session/SKILL.md`

   Create directories as needed.

8. **Verify.** Ask the user to restart Claude Code (or
   run `/mcp`) so the new MCP server is detected. Once
   reconnected, call `validate_specs` to confirm
   everything is wired up. Expect a clean report.

## Rules

- Do not overwrite existing files without asking.
- If any download fails, report the error and continue
  with the remaining steps.
- Report each step as it completes so the user can see
  progress.
