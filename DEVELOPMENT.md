# Development Guide

This document contains technical information for developers working on `gh-project-management`.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Contributing](#contributing)

## Development Setup

### Prerequisites

- Go 1.25.5 or higher
- [GitHub CLI](https://cli.github.com/) v2.0.0 or higher
- Git

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/Zytera/gh-project-management
cd gh-project-management

# Install dependencies
go mod tidy

# Build and install as gh extension
make dev
```

### Development Commands

```bash
# Development workflow (format, vet, build, and install as gh extension)
make dev

# Individual commands
make build      # Build binary only
make fmt        # Format code
make vet        # Run go vet
make install    # Build and install as gh extension
```

After running `make dev` or `make install`, the extension will be available as:
```bash
gh project-management [command]
```

## Project Structure

```
gh-project-management/
├── cmd/                      # CLI command definitions (Cobra)
│   ├── root.go              # Root command and context injection
│   ├── context.go           # Context management commands
│   ├── issue_create.go      # Unified issue creation command
│   ├── field.go             # Custom field management command
│   ├── link.go              # Parent-child relationship management
│   ├── dependencies.go      # Issue dependency management
│   └── transfer.go          # Issue transfer command
│
├── internal/                # Private application code
│   ├── config/             # Configuration management
│   │   └── main.go         # Config loading, saving, validation
│   │
│   ├── gh/                 # GitHub API client wrapper (GraphQL & REST)
│   │   ├── types.go        # Data types (Organization, Project, Field, etc.)
│   │   ├── organization.go # Organization queries
│   │   ├── project.go      # Project and custom field management
│   │   ├── repository.go   # Repository queries
│   │   ├── issue.go        # Issue creation
│   │   ├── templates.go    # Issue template management (repo & defaults)
│   │   ├── issue_types.go  # GitHub organization issue types API
│   │   ├── subissues.go    # Parent-child linking via tasklist API
│   │   ├── dependencies.go # Blocked-by relationships
│   │   ├── transfer.go     # Issue transfer via GraphQL API
│   │   └── list_issues.go  # List recent issues for interactive prompts
│   │
│   └── tui/                # Terminal UI components (Huh forms)
│       ├── context/        # Context management forms
│       └── create-issue/   # Issue creation forms
│
├── pkg/                     # Public/reusable packages
│   ├── context/            # Context operations
│   │   └── operations.go   # Add, delete, switch, list contexts
│   │
│   ├── issue/              # Issue operations
│   │   └── create.go       # Dynamic issue creation with templates
│   │
│   └── templates/          # Template system
│       └── *.go            # Default embedded templates
│
├── main.go                  # Application entry point
├── Makefile                # Build automation
└── go.mod                  # Go module definition
```

### Package Responsibilities

- **cmd/**: Cobra command definitions. Each file defines CLI commands and flags.
- **internal/config/**: Manages YAML configuration at `~/.config/gh-project-management/config.yaml`
- **internal/gh/**: GitHub GraphQL API wrapper using `go-gh` library
- **internal/tui/**: Interactive terminal UI forms using Charm's Huh library
- **pkg/**: Business logic that orchestrates internal packages

## Architecture

### Context System

The extension uses a multi-context configuration system similar to `kubectl`:

```yaml
current-context: project-test

contexts:
  project-test:
    owner_type: org          # "org" or "user"
    owner: Zytera
    project_id: "1"          # Project number (not node ID)
    project_name: Project Test
    default_repo: project-management
    team_repos:
      Backend: backend
      App: mobile-app
      Web: web-app
      Auth: auth
```

#### Context Flow

1. **Command Execution** (`cmd/root.go:Execute()`):
   - Checks if command requires configuration
   - Loads current context from global config
   - Injects context into command via `context.WithValue()`

2. **Context Management** (`pkg/context/operations.go`):
   - `AddContext()`: Creates new context, verifies/creates custom fields, sets as current if first
   - `SwitchContext()`: Changes current-context in global config
   - `DeleteContext()`: Removes context and clears current if needed
   - `ListContexts()`: Returns all configured contexts

### GitHub API Integration

The extension uses both **GraphQL** and **REST API** via `github.com/cli/go-gh/v2/pkg/api`:

- **GraphQL**: Project queries, custom fields, dependencies, issue transfer
- **REST API**: Issue body updates (for sub-issues), template fetching

**No external CLI dependencies**: All GitHub operations use direct API calls.

#### Project Custom Fields & Issue Types

Custom field management is one of the core features. The system manages two project custom fields and uses GitHub's native issue types:

**Project Custom Fields:**
1. **Team Field**: Single-select field with team options from config
2. **Priority Field**: Single-select field with fixed levels (Critical/High/Medium/Low)

**GitHub Native Issue Types:**
3. **Issue Type**: GitHub organization issue type (automatically created via GraphQL if it doesn't exist)

**Field Creation Flow** (`internal/gh/project.go`):

```
EnsureTeamField/EnsurePriorityField
  ↓
GetProjectFields (query all existing fields)
  ↓
FindFieldByName (search for Team/Priority)
  ↓
If not found:
  → CreateSingleSelectField (creates field with initial options)

If found:
  → Compare existing options with required options
  → If missing options:
    → AddOptionsToField (merges existing + new options)
```

**Important Implementation Details**:

- When updating fields, we must include **ALL** options (existing + new)
- The system preserves user's custom options to avoid data loss
- Fields are verified/created when adding or modifying a context
- Project ID in config is the project **number**, not the GraphQL node ID

#### Issue Creation

Issues are created with predefined templates:

```go
// internal/gh/issue.go
CreateIssue(ctx, owner, repo, title, body)
  ↓
1. Query repository ID via GraphQL
2. Create issue via createIssue mutation
3. Return issue with ID, number, URL
```

Templates are defined in `internal/gh/templates.go` with Spanish/English markdown.

#### Parent-Child Relationships (Sub-Issues)

Parent-child relationships are created using GitHub's tasklist system (`internal/gh/subissues.go`):

```go
// AddSubIssue adds a child to a parent via REST API
AddSubIssue(ctx, owner, repo, parentNumber, childNumber)
  ↓
1. Fetch parent issue body via REST API
2. Add tasklist item: "- [ ] #childNumber"
3. Update parent issue body via REST PATCH
```

**How it works:**
- GitHub automatically recognizes tasklist items with issue references (`- [ ] #123`) as parent-child relationships
- The extension updates the parent issue's body directly via REST API
- No external CLI dependencies (replaced `gh sub-issue` extension)

#### Issue Dependencies

Blocked-by relationships are created using GraphQL mutations (`internal/gh/dependencies.go`):

```go
// AddBlockedBy creates a dependency
AddBlockedBy(ctx, owner, repo, blockedIssue, blockingIssue)
  ↓
1. Get node IDs for both issues via GraphQL
2. Execute addIssueDependency mutation
3. Create blocked-by relationship
```

**Important:** Both issues must be in the same repository.

#### Issue Transfer

Issue transfer is handled via GraphQL API (`internal/gh/transfer.go`):

```go
// TransferIssue transfers an issue to another repository
TransferIssue(ctx, issueNumber, targetOwner, targetRepo, sourceRepo)
  ↓
1. Get issue node ID via GraphQL
2. Get target repository node ID via GraphQL
3. Execute transferIssue mutation
4. Return new issue number
```

**Auto-transfer behavior:**
- Triggered when Team field is set (via `--team` flag)
- Can be prevented with `--no-transfer` flag
- Returns the new issue number after transfer

### Issue Hierarchy

```
Epic (default repository)
  └── User Story (default repository)
      └── Task (team-specific repository)
```

**Design Decisions**:

- **Epics** and **User Stories** stay in the default repository (coordination layer)
- **Tasks** are transferred to team repositories (execution layer)
- Only Tasks have a Team field (single owner)
- Epics and User Stories can involve multiple teams

### TUI Forms

Interactive forms use [Charm's Huh library](https://github.com/charmbracelet/huh):

- `internal/tui/context/forms.go`: Context configuration forms
- `internal/tui/create-issue/forms.go`: Issue creation forms

Forms provide:
- Input validation
- Multi-step wizards
- Selection lists
- Confirmation prompts

## Development Workflow

### Adding a New Command

1. Create command file in `cmd/`:
```go
// cmd/create-user-story.go
package cmd

import (
    "github.com/spf13/cobra"
)

var createUserStoryCmd = &cobra.Command{
    Use:   "create-user-story",
    Short: "Create a user story",
    RunE:  runCreateUserStory,
}

func init() {
    rootCmd.AddCommand(createUserStoryCmd)
}

func runCreateUserStory(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

2. Add business logic in `pkg/` or `internal/`
3. Build and test: `make dev`

### Adding a New GraphQL Query

1. Define query in relevant `internal/gh/*.go` file:
```go
func ListIssues(ctx context.Context, owner, repo string) ([]Issue, error) {
    client, err := api.DefaultGraphQLClient()
    if err != nil {
        return nil, err
    }

    query := `query($owner: String!, $repo: String!) {
        repository(owner: $owner, name: $repo) {
            issues(first: 100) {
                nodes {
                    id
                    number
                    title
                }
            }
        }
    }`

    // Execute query...
}
```

2. Add corresponding types to `internal/gh/types.go` if needed

### Modifying Configuration Schema

1. Update structs in `internal/config/main.go`
2. Update validation in `Validate()` methods
3. Update sample YAML in documentation
4. Consider migration path for existing configs

## Testing

### Manual Testing

Use the test repository: https://github.com/Zytera/project-managment-test

```bash
# Clone test repo
git clone https://github.com/Zytera/project-managment-test
cd project-managment-test

# Build and install your changes
cd /path/to/gh-project-management
make dev

# Test commands
cd /path/to/project-managment-test
gh project-management context add test-context
gh project-management create-epic
```

### Testing Custom Fields

To test custom field creation:

1. Create a new GitHub Project (https://github.com/orgs/YOUR_ORG/projects)
2. Add a new context pointing to that project
3. The extension will create Team and Priority fields automatically
4. Verify fields exist in the project settings

### Debugging

```bash
# Enable GitHub CLI debug mode
GH_DEBUG=api gh project-management [command]

# This will show all GraphQL queries and responses
```

## Authentication

The GitHub CLI token must have these scopes:

- `repo` - Repository access
- `read:org` - Organization data
- `read:project` - **Required** for GitHub Projects v2 API
- `write:project` - **Required** to create/update custom fields

```bash
# Add required scopes
gh auth refresh -s project

# Verify
gh auth status
```

## Technology Stack

- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **TUI Framework**: [Huh](https://github.com/charmbracelet/huh)
- **GitHub API**: [go-gh](https://github.com/cli/go-gh) (GraphQL)
- **Configuration**: [yaml.v3](https://github.com/go-yaml/yaml)

## Contributing

### Code Style

- Run `make fmt` before committing
- Run `make vet` to check for common issues
- Follow standard Go naming conventions

### Commit Messages

Use conventional commits format:

```
feat: add user story creation command
fix: resolve custom field verification issue
docs: update installation instructions
refactor: simplify context switching logic
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run `make dev` to verify build
5. Commit your changes
6. Push to your fork
7. Open a Pull Request

## Common Issues

### "Your token has not been granted the required scopes"

**Problem**: GitHub token missing `project` scopes.

**Solution**:
```bash
gh auth refresh -s project
gh auth status
```

### "Context not found"

**Problem**: No current context set or context was deleted.

**Solution**:
```bash
gh project-management context list
gh project-management context use <context-name>
```

### "Repository not found"

**Problem**: Repository name mismatch or insufficient permissions.

**Solution**: Verify repository exists and token has `repo` scope.

## Resources

- [GitHub Projects V2 API](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects)
- [GitHub CLI Extension Development](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions)
- [GraphQL Explorer](https://docs.github.com/en/graphql/overview/explorer)
- [Cobra Documentation](https://cobra.dev/)
- [Huh Documentation](https://github.com/charmbracelet/huh)