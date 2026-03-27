# Prompt Repo

Created by @dawillygene

## Overview

Prompt Repo is a terminal-first prompt management platform with:

- A Laravel backend API for authentication, persistence, validation, and sync.
- A Go CLI frontend for fast local operations, interactive shell workflows, and automation.

The system is designed to let a developer manage prompts fully from Linux terminal while keeping the architecture extensible for future web, desktop, and IDE clients.

## Current Architecture

- Backend: Laravel 13 (PHP 8.3+), MySQL, REST API, token auth middleware.
- CLI: Go 1.24.2, Cobra command framework, interactive shell mode, autocomplete support.
- Data interchange: JSON APIs and JSON import/export.

## Repository Structure

```text
.
├── backend/                         # Laravel API service
│   ├── app/                         # Controllers, Models, Requests, Middleware
│   ├── routes/api.php               # API route definitions
│   ├── database/migrations/         # Schema and domain migrations
│   ├── composer.json                # PHP dependencies/scripts
│   └── package.json                 # Frontend build tooling (Vite)
├── cli/                             # Go CLI runtime
│   ├── cmd/                         # Cobra commands
│   ├── internal/                    # API client, shell/TUI, cache, queue, config
│   ├── completions/                 # bash/zsh/fish completion scripts
│   ├── scripts/                     # helper scripts
│   ├── go.mod                       # Go dependencies
│   └── main.go                      # CLI entrypoint
├── docs/                            # Project docs folder
├── scripts/                         # Root utility scripts
├── my_prompt_repository_requirements.json
├── promptrepo-public-import.json
└── .gitignore
```

## Backend Details

### Core Domain Models

- User
- Prompt
- Category
- Tag
- PromptVersion
- AuditLog
- SyncQueue

### API Surface

Public:

- POST /api/register
- POST /api/login

Authenticated (auth.token middleware):

- POST /api/logout
- GET /api/me
- GET /api/prompts
- GET /api/prompts/trash
- POST /api/prompts
- GET /api/prompts/{promptRef}
- PUT /api/prompts/{promptRef}
- DELETE /api/prompts/{promptRef}
- POST /api/prompts/{promptRef}/restore
- DELETE /api/prompts/{promptRef}/force
- POST /api/prompts/{promptRef}/favorite
- POST /api/prompts/{promptRef}/archive
- GET /api/prompts/{promptRef}/versions
- GET /api/categories
- POST /api/categories
- PUT /api/categories/{categoryRef}
- DELETE /api/categories/{categoryRef}
- GET /api/tags
- POST /api/tags
- PUT /api/tags/{tagRef}
- DELETE /api/tags/{tagRef}
- GET /api/export
- POST /api/import
- GET /api/sync/status
- POST /api/sync
- POST /api/sync/resolve

### Validation and Request Layer

Validation is implemented via dedicated request classes in backend/app/Http/Requests for auth, prompts, categories, and tags.

## CLI Details

### Command Name

The active command is:

- prompt

### Main Cobra Commands

- prompt register
- prompt login
- prompt logout
- prompt whoami
- prompt add
- prompt list (alias: ls)
- prompt show [id-or-slug]
- prompt delete [id-or-slug] (alias: rm)
- prompt search <keyword>
- prompt favorite [id-or-slug] (alias: fav)
- prompt archive <id-or-slug> (alias: arch)
- prompt category list|create|update|delete
- prompt tag list|create|update|delete
- prompt export [file]
- prompt import <file>
- prompt sync
- prompt config set <key> <value>
- prompt ui

### Interactive Shell Features

Launching prompt without subcommands starts an interactive shell with command-style operations such as:

- Navigation: ls, cd, pwd, tree
- Prompt operations: cat, touch, add, edit, rm, mv, cp, copy
- Search: find, grep
- Organization: star/fav/favorite, archive, tag
- Sync: export, import, sync
- Auth: login, register, logout, whoami

The shell includes autocomplete behavior, history navigation, and command-centric workflow.

### CLI Local Config

CLI config is stored under user config path using:

- Directory name: prompt
- File name: config.json

Default API base: http://127.0.0.1:8001

## Quick Start

### 1. Start Backend

```bash
cd backend
cp .env.example .env
php artisan key:generate
php artisan migrate
php artisan serve --host=127.0.0.1 --port=8001
```

### 2. Build CLI

```bash
cd cli
go build -o bin/prompt .
```

### 3. Configure CLI and Authenticate

```bash
./bin/prompt config set api_base http://127.0.0.1:8001
./bin/prompt register --name "Your Name" --email you@example.com --password "YourPass123!"
./bin/prompt login --email you@example.com --password "YourPass123!"
```

### 4. Use the CLI

```bash
./bin/prompt list
./bin/prompt add --title "My Prompt" --content "Explain X clearly"
./bin/prompt show 1
./bin/prompt ui
```

## Development Notes

- Backend dependencies and scripts are managed in backend/composer.json.
- Frontend build tooling for Laravel assets is configured via Vite in backend/package.json.
- CLI completions are in cli/completions for bash, zsh, and fish.
- Requirements and import dataset files exist at repository root and are currently ignored by git as configured.

## Security and Privacy Notes

- Keep backend/.env out of version control.
- Never commit real tokens, database passwords, or production keys.
- Rotate credentials immediately if any secret is exposed.

## Project Identity

- Project: Prompt Repo
- Type: Terminal-first prompt management system
- Created by: @dawillygene
