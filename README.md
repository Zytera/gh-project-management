# gh-project-management

A GitHub CLI extension for managing projects with hierarchical issues (Epics, User Stories, Tasks).

## Features

- 📦 **Hierarchical Issue Management**: Create and manage Epics → User Stories → Tasks
- 🔄 **Multi-Project Support**: Work with multiple GitHub projects using kubectl-style contexts
- 🔗 **Automatic Linking**: Parent-child relationships and cross-repository references
- 📊 **Custom Fields**: Manage Team, Priority, and Type fields via GitHub Projects v2
- 🚀 **Task Transfer**: Automatically transfer tasks to team repositories

## Requirements

- [GitHub CLI](https://cli.github.com/) v2.0.0 or higher
- GitHub token with the following scopes:
  - `repo` - Access to repositories
  - `read:org` - Read organization data
  - `read:project` - **Required** for GitHub Projects v2 API
  - `write:project` - **Required** to create/update custom fields

### Authentication Setup

Before using this extension, ensure your GitHub CLI token has the necessary scopes:

```bash
# Add required project scopes to your GitHub token
gh auth refresh -s project

# Verify authentication
gh auth status
```

You should see `✓ Token scopes: ... project, read:project ...` in the output.

## Installation

### For Users

```bash
# Install directly from GitHub
gh extension install Zytera/gh-project-management
```

### For Development

```bash
# Clone the repository
git clone https://github.com/Zytera/gh-project-management
cd gh-project-management

# Install dependencies
go mod tidy

# Build and install as gh extension
make dev
```

## Quick Start

### First-Time Setup

Run the interactive setup wizard:

```bash
gh project-managment init
```

This will guide you through:
1. Creating a context name
2. Configuring your GitHub organization
3. Setting up your project
4. Defining team repositories

### Basic Usage

```bash
# Create an Epic
gh project-managment create-epic

# List configured contexts
gh project-managment context list

# Switch to a different project
gh project-managment context use <context-name>

# Show current context
gh project-managment context current
```

## Configuration

The extension uses a global configuration file at `~/.config/gh-project-management/config.yaml`:

```yaml
current-context: project-test

contexts:
  project-test:
    org: Zytera
    project_id: "1"
    project_name: Project Test
    default_repo: project-managment
    team_repos:
      Backend: backend
      App: mobile-app
      Web: web-app
      Auth: auth
```

## Commands

### Context Management

```bash
gh project-managment init                    # Initialize configuration
gh project-managment context list            # List all contexts
gh project-managment context current         # Show current context
gh project-managment context use <name>      # Switch context
gh project-managment context add <name>      # Add new context
gh project-managment context delete <name>   # Delete context
```

### Issue Management

```bash
gh project-managment create-epic             # Create an Epic (implemented)
gh project-managment create-user-story       # Create a User Story (coming soon)
gh project-managment create-task             # Create a Task (coming soon)
```

## Development

### Setup

```bash
# Install dependencies
go mod tidy

# Development build (formats, vets, builds, and installs)
make dev

# Build only
make build

# Format code
make fmt

# Run vet
make vet
```

### Testing

Use the testing repository: https://github.com/Zytera/project-managment-test

```bash
git clone https://github.com/Zytera/project-managment-test
cd project-managment-test
gh project-managment init
```

## Issue Hierarchy

The extension implements a 3-level hierarchy:

```
Epic (in default repository)
  └── User Story (in default repository)
      └── Task (transferred to team repositories)
```

### Issue Types

| Type | Repository | Team Field |
|------|------------|------------|
| Epic | Default repo | No (multi-team) |
| User Story | Default repo | No (may involve multiple teams) |
| Task | Team-specific | Yes |
| Bug | Team-specific | Yes (coming soon) |

## Workflow Example

1. **Create Epic** in your default repository (e.g., `project-managment`)
2. **Create User Stories** linked to the Epic
3. **Create Tasks** for each User Story
4. **Tasks are transferred** to team repositories (backend, mobile-app, etc.)
5. **Custom fields** (Team, Priority, Type) are set automatically
6. **Parent issues updated** with child references

## Architecture

- **CLI Framework**: Cobra
- **TUI Framework**: Huh
- **GitHub API**: go-gh (GraphQL)
- **Configuration**: YAML with kubectl-style contexts

## Troubleshooting

### Token Scope Errors

**Error:** `Your token has not been granted the required scopes... The 'id' field requires ['read:project']`

**Solution:** Your GitHub token is missing the `project` scopes. Run:

```bash
gh auth refresh -s project
```

This will prompt you to authorize the additional scopes in your browser. After authorizing, verify with:

```bash
gh auth status
```

## License

TODO