# gh-project-management

A GitHub CLI extension for managing projects with hierarchical issues (Epics, User Stories, Tasks).

## Features

- 📦 **Hierarchical Issue Management**: Create and manage Epics → User Stories → Tasks
- 🔄 **Multi-Project Support**: Work with multiple GitHub projects using kubectl-style contexts
- 🎨 **Interactive TUI**: Beautiful terminal UI powered by Bubble Tea
- 🔗 **Automatic Linking**: Parent-child relationships and cross-repository references
- 📊 **Custom Fields**: Manage Team, Priority, and Type fields via GitHub Projects v2
- 🚀 **Task Transfer**: Automatically transfer tasks to team repositories

## Requirements

- [Go](https://golang.org/dl/) (for development)
- [GitHub CLI](https://cli.github.com/)

## Installation

```bash
# Clone the repository
git clone https://github.com/Zytera/gh-project-management
cd gh-project-management

# Build and install
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
current-context: medapsis

contexts:
  medapsis:
    org: Zytera
    project_id: "3"
    project_name: Medapsis
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
- **TUI Framework**: Bubble Tea + Huh
- **GitHub API**: go-gh (GraphQL)
- **Configuration**: YAML with kubectl-style contexts

For detailed architecture information, see [CLAUDE.md](CLAUDE.md).

## Contributing

This project is in active development. See [CLAUDE.md](CLAUDE.md) for implementation details and roadmap.

## License

TODO

## Reference

Based on the project management workflow documented in:
- [Zytera Project Management Repository](https://github.com/Zytera/project-managment)