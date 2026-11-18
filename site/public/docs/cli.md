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
| **macOS and Linux** | `curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh` |
| **Windows** | See [Windows installation](#install-on-windows) |

Setting Planton up from your coding agent? The Setup Prompt in your organization's Getting Started runs this same installer for you. See [Coding Agents](/docs/coding-agents).

---

## The installer (macOS and Linux)

One command installs the latest CLI on macOS or Linux, Intel or Arm, with no `sudo`:

```bash
curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh
```

The installer:
- Picks the binary for your system and processor
- Checks it against the release's published checksums before installing it
- Puts it at `~/.planton/bin/planton`, and never touches another `planton` on your machine
- Prints how to add `~/.planton/bin` to your PATH if it is not there yet

Run it again any time to update to the latest release. `planton upgrade` updates the same file.

To verify the installation:

```bash
~/.planton/bin/planton version
```

---

## Install on macOS

### Homebrew

Homebrew loads a third-party tap only after you trust it, so the install is three commands:

```bash
brew tap plantonhq/tap
brew trust --tap plantonhq/tap
brew install plantonhq/tap/planton
```

Homebrew picks the binary for your Mac (Intel or Apple Silicon), puts it on your PATH, and owns its updates from then on.

To verify the installation:

```bash
planton version
```

### Direct download

To download the binary yourself:

**Apple Silicon (M1/M2/M3):**

```bash
VERSION=$(curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt)
curl -Lo planton "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/planton-${VERSION}-darwin-arm64"
chmod +x planton
sudo mv planton /usr/local/bin/
```

**Intel Mac:**

```bash
VERSION=$(curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt)
curl -Lo planton "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/planton-${VERSION}-darwin-amd64"
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

Inside WSL, [the installer](#the-installer-macos-and-linux) works as on Linux:

```bash
curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh
```

Or download the Linux binary yourself:

**x86_64 (most systems):**

```bash
VERSION=$(curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt)
curl -Lo planton "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/planton-${VERSION}-linux-amd64"
chmod +x planton
sudo mv planton /usr/local/bin/
```

**ARM64:**

```bash
VERSION=$(curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt)
curl -Lo planton "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/planton-${VERSION}-linux-arm64"
chmod +x planton
sudo mv planton /usr/local/bin/
```

### PowerShell (native Windows)

For native Windows without WSL, open PowerShell as Administrator and run:

**Windows x64 (most systems):**

```powershell
# Download the latest CLI
$version = (Invoke-RestMethod -Uri "https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt").Trim()
Invoke-WebRequest -Uri "https://downloads.planton.ai/client-apps/planton/cli/$version/planton-$version-windows-amd64.exe" -OutFile "planton.exe"

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
# Download the latest CLI
$version = (Invoke-RestMethod -Uri "https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt").Trim()
Invoke-WebRequest -Uri "https://downloads.planton.ai/client-apps/planton/cli/$version/planton-$version-windows-arm64.exe" -OutFile "planton.exe"

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

Run this in the same PowerShell session that installed the CLI (it reuses `$version`):

```powershell
# Download the checksums published with that version
Invoke-WebRequest -Uri "https://downloads.planton.ai/client-apps/planton/cli/$version/checksums.txt" -OutFile "checksums.txt"

# Calculate hash of downloaded file
(Get-FileHash -Path "$env:USERPROFILE\bin\planton.exe" -Algorithm SHA256).Hash.ToLower()

# Compare with checksum in checksums.txt
Get-Content checksums.txt | Select-String "windows-amd64"
```

---

## Install on Linux

[The installer](#the-installer-macos-and-linux) is the quickest way. To download the binary yourself, pick your architecture.

### x86_64 (amd64)

Most Linux servers and desktops use x86_64:

```bash
VERSION=$(curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt)
curl -Lo planton "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/planton-${VERSION}-linux-amd64"
chmod +x planton
sudo mv planton /usr/local/bin/
```

### ARM64 (aarch64)

For ARM-based systems like Raspberry Pi 4, AWS Graviton, or ARM cloud instances:

```bash
VERSION=$(curl -fsSL https://downloads.planton.ai/client-apps/planton/cli/latest/version.txt)
curl -Lo planton "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/planton-${VERSION}-linux-arm64"
chmod +x planton
sudo mv planton /usr/local/bin/
```

> **Tip:** Run `uname -m` to check your architecture. `x86_64` means amd64, `aarch64` means arm64.

### Verify checksum on Linux

For security-conscious installations, verify the download. Run this in the same shell that installed the CLI (it reuses `VERSION`):

```bash
# Download the checksums published with that version
curl -fsSLO "https://downloads.planton.ai/client-apps/planton/cli/${VERSION}/checksums.txt"

# Verify (use linux-arm64 if that is the binary you downloaded)
grep "planton-${VERSION}-linux-amd64$" checksums.txt | awk '{print $1 "  /usr/local/bin/planton"}' | sha256sum -c -
```

---

## Verify installation

After installing, verify the CLI works:

```bash
planton version
```

You should see output like:

```
planton version v0.0.145
```

If you get a "command not found" error, ensure the binary location is in your PATH.

---

## Authentication

Before using the CLI, authenticate with your Planton account. A coding agent working from a Setup Prompt signs in with the prompt's code instead (below).

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
planton context get
```

### Sign in from a Setup Prompt

The Setup Prompt you copy from Getting Started carries a sign-in code that works once, within 15 minutes. Your coding agent runs the line the prompt gives it:

```bash
~/.planton/bin/planton login --grant <code> --instance planton.ai
```

The code becomes a personal API key, named for the agent and the machine. It lasts 30 days, reaches every organization you belong to, and is listed (and revocable) on your API Keys page. The command also sets the organization the prompt named as your context. When this CLI is already signed in as you, it keeps that sign-in and leaves the code unused. When it is signed in as someone else, it changes nothing and says how to sign in beside them with `--account`.

---

## Getting Started from the terminal

Your organization's Getting Started steps can be ticked from the CLI, which is how a coding agent records each step it finished:

```bash
# The steps, and which are done
planton setup done

# Tick two steps
planton setup done connect_cloud create_environment
```

Ticking needs the Owner or Infrastructure Admin role in the organization.

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
- Downloads and installs the update over the binary you ran, when it can write there (the installer's `~/.planton/bin/planton`, or a binary you downloaded)
- For a Homebrew install: runs `brew upgrade planton`

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

#### The installer, or a direct download

Run [the installer](#the-installer-macos-and-linux) again, or download the latest version using the same steps above. Check the current version:

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
