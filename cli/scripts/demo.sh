#!/usr/bin/env bash
# Demo script for MyPrompts CLI auto-completion features

set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}MyPrompts CLI - Auto-Completion Demo${NC}"
echo -e "${BLUE}======================================${NC}\n"

echo -e "${GREEN}✓ CLI Build:${NC}"
echo "  Location: $(pwd)/bin/prompt"
echo "  Version: $(./bin/prompt --version 2>&1 | head -1)"
echo ""

echo -e "${GREEN}✓ Available Commands:${NC}"
./bin/prompt --help | grep -A 20 "Available Commands:" | tail -n +2
echo ""

echo -e "${GREEN}✓ Command Aliases:${NC}"
echo "  list → ls"
echo "  delete → rm"
echo "  favorite → fav"
echo "  archive → arch"
echo ""

echo -e "${GREEN}✓ Completion Scripts Generated:${NC}"
ls -lh completions/bash/prompt.bash | awk '{print "  Bash:  " $9 " (" $5 ")"}'
ls -lh completions/zsh/_prompt | awk '{print "  Zsh:   " $9 " (" $5 ")"}'
ls -lh completions/fish/prompt.fish | awk '{print "  Fish:  " $9 " (" $5 ")"}'
echo ""

echo -e "${GREEN}✓ Sample Commands:${NC}"
echo "  # Basic usage"
echo "  ./bin/prompt list"
echo "  ./bin/prompt add --title 'My Prompt' --content 'Prompt content'"
echo ""
echo "  # With aliases"
echo "  ./bin/prompt ls --search 'code'"
echo "  ./bin/prompt fav my-prompt"
echo ""
echo "  # JSON output"
echo "  ./bin/prompt list --json"
echo "  ./bin/prompt whoami --json"
echo ""

echo -e "${YELLOW}📦 To Install Completions:${NC}"
echo "  ./scripts/install-completions.sh"
echo ""
echo "  Or manually:"
echo "  # Bash"
echo "  source completions/bash/prompt.bash"
echo ""
echo "  # Zsh"
echo "  cp completions/zsh/_prompt ~/.zsh/completions/"
echo "  # Then add to ~/.zshrc:"
echo "  fpath=(~/.zsh/completions \$fpath)"
echo "  autoload -Uz compinit && compinit"
echo ""
echo "  # Fish"
echo "  cp completions/fish/prompt.fish ~/.config/fish/completions/"
echo ""

echo -e "${YELLOW}🎮 Try Tab Completion:${NC}"
echo "  prompt <TAB>           # See all commands"
echo "  prompt add --<TAB>     # See all flags"
echo "  prompt show <TAB>      # See your prompts (requires API)"
echo "  prompt category <TAB>  # See category subcommands"
echo ""

echo -e "${YELLOW}📚 More Info:${NC}"
echo "  README: ./README-COMPLETIONS.md"
echo "  Install: ./scripts/install-completions.sh --help"
echo ""

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}✨ Setup Complete!${NC}"
echo -e "${GREEN}======================================${NC}"
