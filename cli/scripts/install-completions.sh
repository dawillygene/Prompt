#!/usr/bin/env bash
# Installation script for prompt shell completions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_DIR="$(dirname "$SCRIPT_DIR")"
COMPLETIONS_DIR="$CLI_DIR/completions"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}MyPrompts Shell Completion Setup${NC}"
echo -e "${BLUE}================================${NC}\n"

# Detect shell
detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    elif [ -n "$FISH_VERSION" ]; then
        echo "fish"
    else
        echo "unknown"
    fi
}

CURRENT_SHELL=$(detect_shell)
echo -e "${BLUE}Detected shell:${NC} $CURRENT_SHELL\n"

# Install bash completion
install_bash() {
    echo -e "${BLUE}Installing Bash completion...${NC}"
    
    # Try system-wide location first (requires sudo)
    if [ -w "/etc/bash_completion.d" ]; then
        cp "$COMPLETIONS_DIR/bash/prompt.bash" "/etc/bash_completion.d/prompt"
        echo -e "${GREEN}✓${NC} Installed to /etc/bash_completion.d/prompt"
    elif [ -w "/usr/local/etc/bash_completion.d" ]; then
        cp "$COMPLETIONS_DIR/bash/prompt.bash" "/usr/local/etc/bash_completion.d/prompt"
        echo -e "${GREEN}✓${NC} Installed to /usr/local/etc/bash_completion.d/prompt"
    else
        # Fall back to user directory
        mkdir -p "$HOME/.bash_completion.d"
        cp "$COMPLETIONS_DIR/bash/prompt.bash" "$HOME/.bash_completion.d/prompt"
        echo -e "${GREEN}✓${NC} Installed to $HOME/.bash_completion.d/prompt"
        
        # Add source command to .bashrc if not present
        BASHRC="$HOME/.bashrc"
        SOURCE_LINE="source ~/.bash_completion.d/prompt"
        
        if [ -f "$BASHRC" ] && ! grep -q "$SOURCE_LINE" "$BASHRC"; then
            echo "" >> "$BASHRC"
            echo "# MyPrompts completion" >> "$BASHRC"
            echo "$SOURCE_LINE" >> "$BASHRC"
            echo -e "${GREEN}✓${NC} Added source command to $BASHRC"
        fi
    fi
    
    echo -e "${YELLOW}→${NC} Reload your shell or run: ${BLUE}source ~/.bashrc${NC}\n"
}

# Install zsh completion
install_zsh() {
    echo -e "${BLUE}Installing Zsh completion...${NC}"
    
    # Check for common zsh completion directories
    if [ -n "$ZDOTDIR" ]; then
        ZSH_COMP_DIR="$ZDOTDIR/.zsh/completions"
    else
        ZSH_COMP_DIR="$HOME/.zsh/completions"
    fi
    
    # Try system-wide first
    if [ -w "/usr/local/share/zsh/site-functions" ]; then
        cp "$COMPLETIONS_DIR/zsh/_prompt" "/usr/local/share/zsh/site-functions/_prompt"
        echo -e "${GREEN}✓${NC} Installed to /usr/local/share/zsh/site-functions/_prompt"
    elif [ -w "/usr/share/zsh/site-functions" ]; then
        cp "$COMPLETIONS_DIR/zsh/_prompt" "/usr/share/zsh/site-functions/_prompt"
        echo -e "${GREEN}✓${NC} Installed to /usr/share/zsh/site-functions/_prompt"
    else
        # User directory
        mkdir -p "$ZSH_COMP_DIR"
        cp "$COMPLETIONS_DIR/zsh/_prompt" "$ZSH_COMP_DIR/_prompt"
        echo -e "${GREEN}✓${NC} Installed to $ZSH_COMP_DIR/_prompt"
        
        # Add fpath to .zshrc if not present
        ZSHRC="${ZDOTDIR:-$HOME}/.zshrc"
        FPATH_LINE="fpath=(~/.zsh/completions \$fpath)"
        
        if [ -f "$ZSHRC" ] && ! grep -q "prompt" "$ZSHRC" && ! grep -q "$ZSH_COMP_DIR" "$ZSHRC"; then
            echo "" >> "$ZSHRC"
            echo "# MyPrompts completion" >> "$ZSHRC"
            echo "$FPATH_LINE" >> "$ZSHRC"
            echo "autoload -Uz compinit && compinit" >> "$ZSHRC"
            echo -e "${GREEN}✓${NC} Added fpath and compinit to $ZSHRC"
        fi
    fi
    
    echo -e "${YELLOW}→${NC} Reload your shell or run: ${BLUE}source ~/.zshrc${NC}\n"
}

# Install fish completion
install_fish() {
    echo -e "${BLUE}Installing Fish completion...${NC}"
    
    FISH_COMP_DIR="$HOME/.config/fish/completions"
    mkdir -p "$FISH_COMP_DIR"
    cp "$COMPLETIONS_DIR/fish/prompt.fish" "$FISH_COMP_DIR/prompt.fish"
    echo -e "${GREEN}✓${NC} Installed to $FISH_COMP_DIR/prompt.fish"
    echo -e "${YELLOW}→${NC} Fish will automatically load completions on next start\n"
}

# Main installation
if [ "$#" -eq 0 ]; then
    # Auto-detect and install for current shell
    case "$CURRENT_SHELL" in
        bash)
            install_bash
            ;;
        zsh)
            install_zsh
            ;;
        fish)
            install_fish
            ;;
        *)
            echo -e "${RED}✗${NC} Could not detect shell. Please specify: bash, zsh, or fish"
            echo -e "  Usage: $0 [bash|zsh|fish|all]"
            exit 1
            ;;
    esac
else
    # Install for specified shell(s)
    for shell in "$@"; do
        case "$shell" in
            bash)
                install_bash
                ;;
            zsh)
                install_zsh
                ;;
            fish)
                install_fish
                ;;
            all)
                install_bash
                install_zsh
                install_fish
                ;;
            *)
                echo -e "${RED}✗${NC} Unknown shell: $shell"
                echo -e "  Supported: bash, zsh, fish, all"
                exit 1
                ;;
        esac
    done
fi

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Installation complete!${NC}"
echo -e "${GREEN}================================${NC}\n"

echo -e "Try it out:"
echo -e "  ${BLUE}prompt <TAB>${NC}    - See available commands"
echo -e "  ${BLUE}prompt add --<TAB>${NC} - See available flags"
echo -e ""
echo -e "For dynamic completions (prompt IDs, categories, tags),"
echo -e "ensure your API is configured: ${BLUE}prompt config set api_base <url>${NC}"
