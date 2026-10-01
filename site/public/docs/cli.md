---
title: "CLI Installation"
description: "Install the Planton CLI on macOS, Linux, or Windows to manage cloud infrastructure from your terminal"
icon: terminal
order: 2
tags:
  - CLI
  - Installation
  - Getting Started
---

# Planton CLI

The Planton CLI provides terminal access to all platform capabilities. Deploy infrastructure, manage resources, and automate workflows from the command line.

## Why use the CLI?

- **Scriptable deployments** - Integrate Planton into CI/CD pipelines and automation scripts
- **Faster workflows** - Power users can move faster with keyboard-driven commands
- **Local validation** - Validate configurations before deploying to catch errors early
- **Offline editing** - Work with YAML manifests locally, deploy when ready

## Quick install

Choose your platform:

| Platform | Command |
|----------|---------|
| **macOS** | `brew install plantonhq/tap/planton` |
| **Windows** | See [Windows installation](#install-on-windows) |
| **Linux** | See [Linux installation](#install-on-linux) |

---

## Install on macOS

### Homebrew (recommended)

The recommended way to install on macOS is via Homebrew:

```bash
brew install plantonhq/tap/planton
```

Homebrew automatically:
- Downloads the correct binary for your Mac (Intel or Apple Silicon)
- Places it in your PATH
- Removes the macOS quarantine attribute (no Gatekeeper warnings)

To verify the installation:

```bash
planton version
```

### Direct download

If you prefer not to use Homebrew, download the binary directly:

**Apple Silicon (M1/M2/M3):**

```bash
curl -Lo planton https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-darwin-arm64
chmod +x planton
sudo mv planton /usr/local/bin/
```

**Intel Mac:**

```bash
curl -Lo planton https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-darwin-amd64
chmod +x planton
sudo mv planton /usr/local/bin/
```

**Remove Gatekeeper quarantine:**

macOS flags downloaded binaries as untrusted. Remove the quarantine attribute:

```bash
xattr -dr com.apple.quarantine /usr/local/bin/planton
```

> **Tip:** Not sure which Mac you have? Run `uname -m`. If it shows `arm64`, you have Apple Silicon. If it shows `x86_64`, you have Intel.

---

## Install on Windows

### Windows Subsystem for Linux (WSL)

If you use WSL, install the Linux binary using curl:

**x86_64 (most systems):**

```bash
curl -Lo planton https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-linux-amd64
chmod +x planton
sudo mv planton /usr/local/bin/
```

**ARM64:**

```bash
curl -Lo planton https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-linux-arm64
chmod +x planton
sudo mv planton /usr/local/bin/
```

### PowerShell (native Windows)

For native Windows without WSL, open PowerShell as Administrator and run:

**Windows x64 (most systems):**

```powershell
# Download the CLI
Invoke-WebRequest -Uri "https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-windows-amd64.exe" -OutFile "planton.exe"

# Move to a directory in your PATH (create if needed)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"
Move-Item -Path "planton.exe" -Destination "$env:USERPROFILE\bin\planton.exe"

# Add to PATH (run once)
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\bin", "User")
}
```

**Windows ARM64:**

```powershell
# Download the CLI
Invoke-WebRequest -Uri "https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-windows-arm64.exe" -OutFile "planton.exe"

# Move to a directory in your PATH (create if needed)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"
Move-Item -Path "planton.exe" -Destination "$env:USERPROFILE\bin\planton.exe"

# Add to PATH (run once)
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\bin", "User")
}
```

After adding to PATH, restart your terminal for changes to take effect.

> **Tip:** Check your architecture in PowerShell with `$env:PROCESSOR_ARCHITECTURE`. `AMD64` is x64, `ARM64` is ARM.

### Verify checksum on Windows

```powershell
# Download checksums
Invoke-WebRequest -Uri "https://downloads.planton.ai/cli/v0.0.3/checksums.txt" -OutFile "checksums.txt"

# Calculate hash of downloaded file
(Get-FileHash -Path "planton.exe" -Algorithm SHA256).Hash.ToLower()

# Compare with checksum in checksums.txt
Get-Content checksums.txt | Select-String "windows-amd64"
```

---

## Install on Linux

Download the binary directly for your architecture.

### x86_64 (amd64)

Most Linux servers and desktops use x86_64:

```bash
curl -Lo planton https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-linux-amd64
chmod +x planton
sudo mv planton /usr/local/bin/
```

### ARM64 (aarch64)

For ARM-based systems like Raspberry Pi 4, AWS Graviton, or ARM cloud instances:

```bash
curl -Lo planton https://downloads.planton.ai/cli/v0.0.3/planton-v0.0.3-linux-arm64
chmod +x planton
sudo mv planton /usr/local/bin/
```

> **Tip:** Run `uname -m` to check your architecture. `x86_64` means amd64, `aarch64` means arm64.

### Verify checksum on Linux

For security-conscious installations, verify the download:

```bash
# Download checksums
curl -LO https://downloads.planton.ai/cli/v0.0.3/checksums.txt

# Verify (replace filename with your download)
sha256sum -c checksums.txt --ignore-missing
```

---

## Verify installation

After installing, verify the CLI works:

```bash
planton version
```

You should see output like:

```
planton version v0.0.3
```

If you get a "command not found" error, ensure the binary location is in your PATH.

---

## Authentication

Before using the CLI, authenticate with your Planton account:

### Step 1: Log in

```bash
planton auth login
```

This opens your browser to sign in through the Planton console. The CLI gets a sign-in of its own, checks it with the platform before it reports success, and stores it locally; it renews itself, and when it no longer can, every command that needs a sign-in tells you to run `planton auth login` again (commands that need none, like browsing the public catalog, keep working).

### Step 2: Set your organization context

```bash
planton context set --org your-org-name
```

This sets the default organization for all CLI commands. Replace `your-org-name` with your actual organization ID.

### Step 3: (Optional) Set environment context

```bash
planton context set --org your-org-name --env prod
```

Setting an environment makes it the default for resource operations, so you don't need to specify `--env` with every command.

### Verify context

```bash
planton context show
```

---

## Shell completion

Enable tab-completion for all commands, subcommands, and flags. This is a one-time setup per machine.

### Zsh (default on macOS)

Add this line to your `~/.zshrc`:

```bash
source <(planton completion zsh)
```

Then reload your shell:

```bash
source ~/.zshrc
```

### Bash

Add this line to your `~/.bashrc`:

```bash
source <(planton completion bash)
```

> **Note:** Bash completion requires the `bash-completion` package. On macOS: `brew install bash-completion@2`. On Linux it is usually pre-installed.

### Fish

```bash
planton completion fish | source
```

To persist across sessions, save it to the Fish completions directory:

```bash
planton completion fish > ~/.config/fish/completions/planton.fish
```

### PowerShell

Add this line to your PowerShell profile:

```powershell
planton completion powershell | Out-String | Invoke-Expression
```

### Verify it works

Open a new terminal and type:

```bash
planton <Tab>
```

You should see a list of available commands. Pressing Tab after flags like `--` shows available flag names.

---

## Upgrading

### Self-Update (Recommended)

The CLI includes a built-in upgrade command that works on all platforms:

```bash
planton upgrade
```

This automatically:
- Checks for the latest version
- Downloads and installs the update
- On macOS with Homebrew: runs `brew upgrade planton`
- On other platforms: downloads the binary directly

**Additional options:**

```bash
# Check for updates without installing
planton upgrade --check

# Force upgrade even if already on latest
planton upgrade --force
```

### Manual Upgrade

#### macOS (Homebrew)

```bash
brew update && brew upgrade planton
```

#### Linux and Windows (direct download)

Download the latest version using the same installation steps above. Check the current version:

```bash
planton version
```

---

## Troubleshooting

### "command not found" after installation

The binary isn't in your PATH. Either:
- Move the binary to a directory already in your PATH (like `/usr/local/bin`)
- Add the binary's directory to your PATH

### macOS: "cannot verify developer" dialog

Run this to remove the quarantine flag:

```bash
xattr -dr com.apple.quarantine /usr/local/bin/planton
```

Or if you used Homebrew, reinstall:

```bash
brew reinstall planton
```

### Permission denied when installing

Use `sudo` to move the binary to system directories:

```bash
sudo mv planton /usr/local/bin/
```

Or install to a user directory that doesn't require root:

```bash
mkdir -p ~/.local/bin
mv planton ~/.local/bin/
# Add ~/.local/bin to your PATH in .bashrc or .zshrc
```

### Windows: "execution policy" error

If PowerShell blocks script execution, run:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

---

## What's next?

Now that you have the CLI installed:

1. **[Getting Started Guide](/docs/getting-started)** - Deploy your first resource
2. **[Platform Overview](/docs/platform)** - Understand how Planton works
3. **[Infrastructure](/docs/infrastructure)** - Deploy cloud infrastructure via CLI
4. **[Secrets](/docs/secrets)** - Manage secrets from the terminal (`planton secret set|get|describe|history|snippet|access`, `planton secret backend verify`)
5. **[Coding Agents](/docs/coding-agents)** - Install the Planton skills so Cursor, Claude Code, or any coding agent drives this CLI for you

For CLI command reference:

```bash
planton --help
```
