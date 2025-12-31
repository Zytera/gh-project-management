# gh-project-management

A GitHub CLI extension for managing hierarchical projects with Epics, User Stories, and Tasks using GitHub Projects V2.

## Features

- 📦 **Hierarchical Issue Management**: Create and manage Epics → User Stories → Tasks → Subtasks
- 🎯 **Flexible Templates**: Use default templates or custom repository templates
- 🔍 **Dynamic Field Discovery**: See available fields for any template with `--show-fields`
- 🔗 **Issue Linking**: Parent-child relationships using GitHub's sub-issue system
- 🚫 **Dependency Management**: Block issues with blocked-by relationships
- 📊 **Custom Fields**: Manage Team, Priority, and Type fields via GitHub Projects V2
- 🚀 **Smart Transfer**: Automatically transfer issues to team repositories based on Team field
- 🔄 **Multi-Project Support**: Work with multiple GitHub projects using kubectl-style contexts

## Requirements

- [GitHub CLI](https://cli.github.com/) v2.0.0 or higher
- A GitHub account with access to:
  - At least one repository (for creating issues)
  - A GitHub Project V2 (organization or personal project)
  - Appropriate organization permissions (if using organization projects)

## Installation

### Install the Extension

```bash
# Install from GitHub
gh extension install Zytera/gh-project-management
```

### Setup Authentication

Add required scopes to your GitHub token:

```bash
gh auth refresh -s project
```

Verify authentication:

```bash
gh auth status
```

You should see `project`, `read:project`, and `write:project` in the token scopes.

## Quick Start

### First-Time Setup

Run the interactive setup wizard:

```bash
gh project-management init
```

This will guide you through:
1. Creating a context name
2. Configuring your GitHub organization
3. Setting up your project
4. Defining team repositories

### Create Your First Issue

```bash
# See available fields for epic type
gh project-management issue create --type epic --show-fields

# Create an epic
gh project-management issue create --type epic \
  --title "User Authentication System" \
  --field description="Complete authentication system" \
  --field objective="Secure user authentication" \
  --field stories="Login, Registration, Password Reset" \
  --field acceptance_criteria="All flows tested" \
  --field teams="Backend, Frontend"
```

## Commands

### Issue Management

#### Create Issues

Create issues using templates (repository or defaults) with support for custom fields, dependencies, and parent-child relationships:

```bash
# Show available fields for a type
gh project-management issue create --type <type> --show-fields

# Create an issue with template fields
gh project-management issue create --type <type> \
  --title "Issue Title" \
  --field field_name="value" \
  --field another_field="value"

# Create an issue with custom fields, parent, and dependencies
gh project-management issue create --type <type> \
  --title "Issue Title" \
  --field field_name="value" \
  --team Backend \
  --priority High \
  --type-field Task \
  --parent 44 \
  --depends-on 45 \
  --depends-on 46
```

**Available default types:**
- `epic` - Project epics
- `user_story` - User stories
- `task` - Technical tasks
- `bug` - Bug reports
- `feature` - Feature requests

**Custom types:**
You can use any custom template from your repository's `.github/ISSUE_TEMPLATE/` directory.

**Available flags:**
- `--type` - Issue type/template to use (required)
- `--title` - Issue title (required)
- `--field` - Template field values (repeatable)
- `--team` - Team field (Backend, App, Web, Auth, etc.)
- `--priority` - Priority field (Critical, High, Medium, Low)
- `--type-field` - Type field (Epic, User Story, Story, Task, Bug, Feature)
- `--parent` - Parent issue number to link to
- `--depends-on` - Issue numbers that block this issue (repeatable)
- `--no-transfer` - Prevent automatic transfer when Team is set
- `--show-fields` - Show available template fields and exit

**Interactive prompts:**
If `--parent` or `--depends-on` flags are not provided, the CLI will interactively prompt you to select from recent open issues.

**Auto-transfer behavior:**
When a Team field is set, the issue is automatically transferred to the corresponding team repository **unless** `--no-transfer` is specified. This happens after the issue is created and all fields are set.

**Examples:**

```bash
# Create an epic
gh project-management issue create --type epic \
  --title "User Management" \
  --field description="Complete user management system" \
  --field objective="Enable user registration" \
  --field stories="Login, Registration, Password Reset"

# Create a task with custom fields and auto-transfer
gh project-management issue create --type task \
  --title "Implement login API" \
  --field description="Create REST endpoint for login" \
  --field checklist="- [ ] Create endpoint\n- [ ] Add validation\n- [ ] Write tests" \
  --team Backend \
  --priority High \
  --type-field Task \
  --parent 44
# This will automatically transfer to the Backend repository

# Create a task with custom fields but prevent auto-transfer
gh project-management issue create --type task \
  --title "Implement login API" \
  --field description="Create REST endpoint for login" \
  --team Backend \
  --priority High \
  --no-transfer
# This will NOT transfer automatically

# Create with dependencies
gh project-management issue create --type task \
  --title "Implement login validation" \
  --field description="Add validation to login endpoint" \
  --depends-on 48 \
  --depends-on 49
# Issue #48 and #49 must be completed before this task

# Create a bug
gh project-management issue create --type bug \
  --title "Login fails with special characters" \
  --field description="Users cannot login with @ in password" \
  --field steps_to_reproduce="1. Enter email\n2. Enter password with @" \
  --field expected_behavior="User logged in" \
  --field actual_behavior="Error displayed" \
  --field severity="High" \
  --priority Critical

# Use a custom template from your repo
gh project-management issue create --type my-custom-type \
  --title "Custom Issue" \
  --field custom_field="value"

# Interactive mode (prompts for parent and dependencies)
gh project-management issue create --type task \
  --title "New Task" \
  --field description="Task description"
# Will prompt: "Do you want to link this issue to a parent? (y/N):"
# Will prompt: "Do you want to add dependencies? (y/N):"
```

### Custom Fields Management

Set Team, Priority, and Type fields for existing issues in GitHub Projects:

```bash
# Set team field (automatically transfers to team repository)
gh project-management field set <issue-number> --team Backend

# Set multiple fields (automatically transfers)
gh project-management field set 48 --team Backend --priority High --type Task

# Set fields but prevent automatic transfer
gh project-management field set 48 --team Backend --priority Critical --no-transfer
```

**Available values:**
- **Team**: Backend, App, Web, Auth (from your config)
- **Priority**: Critical, High, Medium, Low
- **Type**: Epic, User Story, Story, Task, Bug, Feature

**Auto-transfer behavior:**
When setting the Team field, the issue is **automatically transferred** to the corresponding team repository unless `--no-transfer` is specified.

**Auto-transfer mapping:**
- Backend → `backend` repository
- App → `mobile-app` repository
- Web → `web-app` repository
- Auth → `auth` repository

**Note:** Custom fields are now best set during issue creation using the `issue create` command with `--team`, `--priority`, and `--type-field` flags.

### Issue Linking (Parent-Child)

Create hierarchical relationships between issues using GitHub's tasklist-based parent-child system:

```bash
# Link child issue to parent
gh project-management link add <parent-issue> <child-issue>

# Remove link
gh project-management link remove <parent-issue> <child-issue>
```

**How it works:**
The extension uses GitHub's API to update the parent issue's body with a tasklist item (`- [ ] #123`) that GitHub automatically recognizes as a parent-child relationship.

**Examples:**

```bash
# Link User Story #45 to Epic #44
gh project-management link add #44 #45

# Link Task #48 to User Story #45
gh project-management link add 45 48

# Remove link
gh project-management link remove #44 #45
```

**Note:** Both issues must be in the same repository. For cross-repository hierarchies, use the parent issue body to manually reference the child with the full format `Owner/Repo#123`.

### Dependency Management

Establish blocked-by relationships between issues:

```bash
# Issue A is blocked by issue B
gh project-management dependency add <blocked-issue> <blocking-issue>
```

**Examples:**

```bash
# Issue #46 is blocked by issue #45
gh project-management dependency add #46 #45

# Cross-repository dependency (both must be in same repo)
gh project-management dependency add #50 Zytera/backend#25
```

**Important:** Both issues must be in the same repository for dependencies to work.

### Issue Transfer

Transfer issues between repositories using GitHub's GraphQL API:

```bash
# Transfer issue to another repository
gh project-management transfer issue <issue-number> --target <repo-name>
```

**Example:**

```bash
# Transfer issue #48 to backend repository
gh project-management transfer issue 48 --target backend
```

**Auto-transfer:**
Issues are automatically transferred when the Team field is set (either during creation with `issue create --team` or afterwards with `field set --team`) unless the `--no-transfer` flag is used.

**Note:** The issue number will change after transfer. The new issue number is returned by the command. Update parent issue references accordingly with the format `Owner/Repo#NewNumber`.

### Context Management

Manage multiple GitHub Projects:

```bash
gh project-management init                    # Initialize configuration
gh project-management context list            # List all contexts
gh project-management context current         # Show current context
gh project-management context use <name>      # Switch context
gh project-management context add <name>      # Add new context
gh project-management context delete <name>   # Delete context
```

## Complete Workflow Example

This example shows the **modern integrated approach** using the `issue create` command with all features in a single command.

### 1. Create an Epic

```bash
# First, see what fields are available
gh project-management issue create --type epic --show-fields

# Create the epic
gh project-management issue create --type epic \
  --title "User Authentication System" \
  --field description="Complete authentication system with login, registration, and password reset" \
  --field objective="Provide secure user authentication" \
  --field stories="Login, Registration, Password Reset, Two-Factor Auth" \
  --field acceptance_criteria="All authentication flows work securely" \
  --field teams="Backend, Frontend, Auth"

# Creates issue #44
```

### 2. Create a User Story and Link to Epic

```bash
# Create user story with parent link in one command
gh project-management issue create --type user_story \
  --title "User Login" \
  --field description="As a user, I want to log in to access my account" \
  --field tasks="Backend API, Frontend form, Session management, Tests" \
  --field acceptance_criteria="Users can login with email and password" \
  --field teams="Backend, Frontend" \
  --parent 44

# Creates issue #45 and automatically links it to Epic #44
```

### 3. Create a Task with All Metadata (Modern Approach)

```bash
# Create task with parent, custom fields, dependencies, and auto-transfer - all in one command!
gh project-management issue create --type task \
  --title "Implement Login API Endpoint" \
  --field description="Create REST endpoint for user authentication" \
  --field checklist="- [ ] Create POST /api/auth/login endpoint\n- [ ] Add JWT token generation\n- [ ] Add validation\n- [ ] Write unit tests\n- [ ] Write integration tests" \
  --parent 45 \
  --team Backend \
  --priority High \
  --type-field Task \
  --depends-on 47

# This single command will:
# 1. Create the issue (#48)
# 2. Link it to User Story #45 as a child
# 3. Set Team, Priority, and Type fields
# 4. Add dependency on issue #47
# 5. Automatically transfer to Backend repository
# 6. Return new issue number (e.g., backend#27)
```

### Alternative: Step-by-Step Approach

If you prefer the traditional step-by-step approach:

#### 3a. Create the Task

```bash
gh project-management issue create --type task \
  --title "Implement Login API Endpoint" \
  --field description="Create REST endpoint for user authentication" \
  --field checklist="- [ ] Create endpoint\n- [ ] Add tests"

# Creates issue #48
```

#### 3b. Link Task to User Story

```bash
gh project-management link add #45 #48
```

#### 3c. Add Dependencies

```bash
gh project-management dependency add #48 #47
```

#### 3d. Set Custom Fields and Auto-Transfer

```bash
# Automatically transfers when Team is set
gh project-management field set 48 --team Backend --priority High --type Task

# Or prevent auto-transfer
gh project-management field set 48 --team Backend --priority High --no-transfer
```

### 4. Update Parent with Cross-Repo Reference

After transfer, the issue number changes (e.g., `#48` → `backend#27`). Update the parent User Story body with the cross-repository reference:

```markdown
### Zytera/backend#27 - [Backend] Implement Login API Endpoint
```

## Workflow Recommendations

### Best Practice: Integrated Approach

Use the `issue create` command with all flags for the most efficient workflow:

```bash
gh project-management issue create --type task \
  --title "Task Title" \
  --field description="Description" \
  --parent <parent-issue> \
  --team <team> \
  --priority <priority> \
  --type-field <type> \
  --depends-on <blocking-issue>
```

This approach:
- ✅ Creates everything in one command
- ✅ Automatically handles parent linking
- ✅ Sets all custom fields before transfer
- ✅ Adds dependencies before transfer
- ✅ Auto-transfers to the correct repository
- ✅ Returns the new issue number immediately

### When to Use Step-by-Step

Use the traditional commands when:
- You need to update existing issues
- You want to review the issue before transferring
- You need to prevent automatic transfer temporarily

## Templates

### Default Templates

The extension includes default English templates for:
- **Epic**: Project epics with objectives and stories
- **User Story**: User stories with tasks and acceptance criteria
- **Task**: Technical tasks with checklists
- **Bug**: Bug reports with reproduction steps and severity
- **Feature**: Feature requests with problem statements and solutions

### Custom Repository Templates

You can override defaults by creating templates in your repository:

```
.github/
└── ISSUE_TEMPLATE/
    ├── epic.yml
    ├── custom_type.yml
    └── ...
```

The extension will automatically use repository templates if available, falling back to defaults.

### Discovering Template Fields

Use `--show-fields` to see available fields for any template:

```bash
gh project-management issue create --type epic --show-fields
```

Output:
```
Template: Epic
Type: Epic
Source: default embedded template
Description: Template for creating a new project epic

Required fields:
  --field description="..."  # 📝 Description
      General description of the epic and what functionalities it includes
  --field objective="..."  # 🎯 Objective
      Main objective to be achieved with this epic
  ...

Optional fields:
  --field estimation="..."  # 📊 Estimation
      Estimation of stories, tasks and complexity
  ...

Example usage:
  gh project-management issue create --type Epic \
    --title "Issue Title" \
    --field description="..." \
    --field objective="..."
```

## Configuration

### Configuration File

Configuration is stored at `~/.config/gh-project-management/config.yaml`:

```yaml
current-context: project-test

contexts:
  project-test:
    owner_type: org                    # "org" or "user"
    owner: Zytera                      # Organization or username
    project_id: "1"                    # GitHub Project number
    project_name: Project Test         # Display name
    default_repo: project-management   # Main repository for Epics/Stories
    team_repos:                        # Team-specific repositories
      Backend: backend
      App: mobile-app
      Web: web-app
      Auth: auth
```

### Configuration Fields

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `owner_type` | Yes | Type of project owner | `org` or `user` |
| `owner` | Yes | GitHub organization or username | `Zytera` |
| `project_id` | Yes | Project number (from project URL) | `1` |
| `project_name` | Yes | Human-readable project name | `Project Test` |
| `default_repo` | Yes | Repository for Epics and User Stories | `project-management` |
| `team_repos` | Yes | Map of team names to repositories | `Backend: backend` |

### Finding Your Project ID

The project ID is the number in your GitHub Project URL:
- Organization: `https://github.com/orgs/YOUR_ORG/projects/1` → Project ID is `1`
- User: `https://github.com/users/YOUR_USERNAME/projects/5` → Project ID is `5`

## Issue Hierarchy

The extension implements a multi-level hierarchy:

```
Epic (project-management repo)
  └── User Story (project-management repo)
      └── Task (backend/mobile-app/web-app/auth repos)
          └── Subtask (optional)
```

### Workflow Approaches

**Modern Approach (Recommended):**
Use the integrated `issue create` command to do everything in one step:

```bash
gh project-management issue create --type task \
  --title "Task Title" \
  --field description="..." \
  --parent <parent-issue> \
  --team <team> \
  --priority <priority> \
  --type-field <type> \
  --depends-on <blocking-issue>
```

This automatically handles:
1. ✅ Creating the issue
2. ✅ Linking to parent
3. ✅ Setting custom fields
4. ✅ Adding dependencies
5. ✅ Auto-transferring to team repository
6. ✅ Returning the new issue number

**Traditional Step-by-Step Approach:**
If needed, follow this order:

1. ✅ **Create issue** in source repository
2. ✅ **Link to parent** with `link add`
3. ✅ **Set custom fields** with `field set` (especially Team)
4. ✅ **Set dependencies** with `dependency add` (same repo required)
5. ✅ **Auto-transfer** happens when Team is set (or use `--no-transfer` to prevent)
6. ✅ **Update parent** with cross-repo reference

**Important Notes:**
- Custom fields should be set before transfer to determine target repository
- Dependencies require both issues to be in the same repository
- After transfer, the issue number changes and parent references must be updated manually

## Custom Fields

The extension automatically manages three custom fields in GitHub Projects:

### Team Field
- **Type**: Single-select dropdown
- **Options**: Automatically populated from `team_repos` configuration
- **Usage**: Determines which repository to transfer tasks to
- **Auto-transfer**: When set (via `--team` flag), automatically moves issue to team repository unless `--no-transfer` is specified
- **Set during creation**: `--team Backend`
- **Set after creation**: `field set <issue> --team Backend`

### Priority Field
- **Type**: Single-select dropdown
- **Options**: Critical 🔴, High 🟠, Medium 🟡, Low ⚪
- **Colors**: Automatically set (Red, Orange, Yellow, Gray)
- **Set during creation**: `--priority High`
- **Set after creation**: `field set <issue> --priority High`

### Type Field
- **Type**: Single-select dropdown
- **Options**: Epic, User Story, Story, Task, Bug, Feature
- **Source**: GitHub organization issue types
- **Management**: Automatically ensures types exist
- **Set during creation**: `--type-field Task`
- **Set after creation**: `field set <issue> --type Task`

## Permissions

### Required GitHub Token Scopes

| Scope | Purpose |
|-------|---------|
| `repo` | Create and manage issues |
| `read:org` | Read organization data |
| `read:project` | Query GitHub Projects V2 |
| `write:project` | Create/update custom fields |

### Organization Permissions

For organization projects, ensure:
- Member access to the organization
- Write access to repositories
- Admin or write access to the GitHub Project

## Troubleshooting

### Common Issues

**Missing scopes:**
```bash
gh auth refresh -s project
gh auth status
```

**No configuration found:**
```bash
gh project-management context add my-project
```

**Dependencies fail after transfer:**
- Dependencies must be set BEFORE transferring issues
- Both issues must be in the same repository

**Custom fields not appearing:**
- Verify `write:project` permission
- Check project access permissions
- Re-add context to trigger field creation

**Template not found:**
- Check template name matches file name
- Use `--show-fields` to verify template exists
- Templates should be in `.github/ISSUE_TEMPLATE/*.yml`

### Debug Mode

Enable debug output:

```bash
GH_DEBUG=api gh project-management <command>
```

## Examples

### Complete Epic to Task Flow (Modern Integrated Approach)

```bash
# 1. Create Epic
gh project-management issue create --type epic \
  --title "E-commerce Checkout" \
  --field description="Complete checkout system" \
  --field objective="Enable users to purchase products" \
  --field stories="Cart, Checkout, Payment" \
  --field acceptance_criteria="Users can complete purchases"
# Creates issue #100

# 2. Create User Story with parent link
gh project-management issue create --type user_story \
  --title "Shopping Cart" \
  --field description="As a user, I want to manage my cart" \
  --field tasks="Add items, Remove items, Update quantities" \
  --field acceptance_criteria="Users can manage cart items" \
  --parent 100
# Creates issue #101 and automatically links to Epic #100

# 3. Create Task with everything in one command
gh project-management issue create --type task \
  --title "Implement Cart API" \
  --field description="Create REST endpoints for cart management" \
  --field checklist="- [ ] POST /cart/items\n- [ ] DELETE /cart/items/:id\n- [ ] PUT /cart/items/:id" \
  --parent 101 \
  --team Backend \
  --priority High \
  --type-field Task
# Creates issue, links to #101, sets fields, and auto-transfers to backend repo!
# Returns new issue number (e.g., backend#27)
```

### Complete Epic to Task Flow (Traditional Step-by-Step)

```bash
# 1. Create Epic
gh project-management issue create --type epic \
  --title "E-commerce Checkout" \
  --field description="Complete checkout system" \
  --field objective="Enable users to purchase products" \
  --field stories="Cart, Checkout, Payment" \
  --field acceptance_criteria="Users can complete purchases"

# 2. Create User Story
gh project-management issue create --type user_story \
  --title "Shopping Cart" \
  --field description="As a user, I want to manage my cart" \
  --field tasks="Add items, Remove items, Update quantities" \
  --field acceptance_criteria="Users can manage cart items"

# 3. Link User Story to Epic
gh project-management link add #100 #101

# 4. Create Task
gh project-management issue create --type task \
  --title "Implement Cart API" \
  --field description="Create REST endpoints for cart management" \
  --field checklist="- [ ] POST /cart/items\n- [ ] DELETE /cart/items/:id\n- [ ] PUT /cart/items/:id"

# 5. Link Task to User Story
gh project-management link add #101 #102

# 6. Set fields (auto-transfers)
gh project-management field set 102 --team Backend --priority High --type Task

# Task is now in backend repository with all fields set!
```

### Using Custom Templates

```bash
# Create .github/ISSUE_TEMPLATE/spike.yml in your repo
# Then use it:
gh project-management issue create --type spike --show-fields
gh project-management issue create --type spike \
  --title "Research GraphQL Performance" \
  --field question="How to optimize GraphQL queries?" \
  --field approach="Benchmark different approaches"
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for development setup and contribution guidelines.

## Support

If you encounter issues:
1. Check this README and troubleshooting section
2. Review [GitHub Issues](https://github.com/Zytera/gh-project-management/issues)
3. Open a new issue with detailed information

## License

TODO