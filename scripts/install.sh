#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

REPO="bluegardenproject/github-butler"
INSTALL_DIR="$HOME/.github-butler"
BINARY_NAME="github-butler"

echo -e "${BOLD}${BLUE}github-butler Installer${NC}"
echo -e "Installing to: ${YELLOW}$INSTALL_DIR${NC}"
echo

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *)
        echo -e "${RED}Error: Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

echo -e "Detected: ${GREEN}$OS-$ARCH${NC}"

echo -e "${BLUE}Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"

echo -e "${BLUE}Fetching latest release...${NC}"
RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"
DOWNLOAD_URL=$(curl -fsSL "$RELEASE_URL" | grep -o "https://.*github-butler-$OS-$ARCH[^\"]*" | head -n 1 || true)

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}Error: Could not find binary for $OS-$ARCH${NC}"
    echo -e "${YELLOW}Available releases: https://github.com/$REPO/releases${NC}"
    exit 1
fi

echo -e "Download URL: ${GREEN}$DOWNLOAD_URL${NC}"

echo -e "${BLUE}Downloading github-butler...${NC}"
TEMP_FILE=$(mktemp)
curl -L -o "$TEMP_FILE" "$DOWNLOAD_URL"

echo -e "${BLUE}Installing binary...${NC}"
mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo -e "${BLUE}Adding to PATH...${NC}"

add_path_posix() {
    local rc="$1"
    local marker="# github-butler (auto-added by install.sh)"
    local line="export PATH=\"$INSTALL_DIR:\$PATH\""

    mkdir -p "$(dirname "$rc")"
    [ -f "$rc" ] || touch "$rc"

    if grep -Fq "$INSTALL_DIR" "$rc" 2>/dev/null; then
        echo -e "${YELLOW}$INSTALL_DIR already referenced in $rc${NC}"
        return
    fi

    {
        echo ""
        echo "$marker"
        echo "$line"
    } >> "$rc"
    echo -e "${GREEN}Added $INSTALL_DIR to PATH in $rc${NC}"
}

add_path_fish() {
    local conf_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish/conf.d"
    local conf="$conf_dir/github-butler.fish"

    mkdir -p "$conf_dir"
    if [ -f "$conf" ] && grep -Fq "$INSTALL_DIR" "$conf"; then
        echo -e "${YELLOW}$INSTALL_DIR already referenced in $conf${NC}"
        return
    fi

    cat > "$conf" <<EOF
# github-butler (auto-added by install.sh)
fish_add_path -gP $INSTALL_DIR
EOF
    echo -e "${GREEN}Added $INSTALL_DIR to PATH in $conf${NC}"
}

SHELL_NAME="$(basename "${SHELL:-}")"
SHELL_CONFIG=""

case "$SHELL_NAME" in
    zsh)
        SHELL_CONFIG="${ZDOTDIR:-$HOME}/.zshrc"
        add_path_posix "$SHELL_CONFIG"
        ;;
    bash)
        SHELL_CONFIG="$HOME/.bashrc"
        add_path_posix "$SHELL_CONFIG"
        if [ "$(uname -s)" = "Darwin" ] && [ -f "$HOME/.bash_profile" ]; then
            add_path_posix "$HOME/.bash_profile"
        fi
        ;;
    fish)
        add_path_fish
        SHELL_CONFIG="${XDG_CONFIG_HOME:-$HOME/.config}/fish/conf.d/github-butler.fish"
        ;;
    *)
        SHELL_CONFIG="$HOME/.profile"
        add_path_posix "$SHELL_CONFIG"
        echo -e "${YELLOW}Unrecognized shell '$SHELL_NAME' — wrote to $SHELL_CONFIG.${NC}"
        echo -e "${YELLOW}If your shell doesn't source that file, add $INSTALL_DIR to PATH manually.${NC}"
        ;;
esac

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) export PATH="$INSTALL_DIR:$PATH" ;;
esac

echo -e "${BLUE}Verifying installation...${NC}"
if "$INSTALL_DIR/$BINARY_NAME" --version >/dev/null 2>&1; then
    echo -e "${GREEN}Installation successful!${NC}"
else
    echo -e "${YELLOW}Installation completed, but verification failed${NC}"
    echo -e "${YELLOW}  You may need to restart your terminal${NC}"
fi

echo
echo -e "${BOLD}${GREEN}Installation Complete!${NC}"
echo
echo -e "${BOLD}Usage:${NC}"
echo -e "  ${GREEN}github-butler${NC}            - Start the dashboard"
echo -e "  ${GREEN}github-butler --version${NC}  - Show version information"
echo
echo -e "${YELLOW}Note: You may need to restart your terminal.${NC}"
if [ -n "$SHELL_CONFIG" ]; then
    if [ "$SHELL_NAME" = "fish" ]; then
        echo -e "  ${BLUE}source $SHELL_CONFIG${NC}    # fish"
    else
        echo -e "  ${BLUE}source $SHELL_CONFIG${NC}"
    fi
fi
echo
