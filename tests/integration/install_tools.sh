#!/bin/bash

# MockServer 测试工具安装脚本
# 自动安装所有 E2E 测试所需的工具

set -e

# 脚本目录
SCRIPT_DIR="$(dirname "$0")"
INSTALLER_LIB="$SCRIPT_DIR/lib/tool_installer.sh"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# 显示横幅
show_banner() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   MockServer 测试工具安装器${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""
    echo -e "${CYAN}此脚本将自动安装 E2E 测试所需的所有工具${NC}"
    echo ""
}

# 显示使用说明
show_usage() {
    echo -e "${CYAN}使用方法:${NC}"
    echo -e "  $0 [选项]"
    echo ""
    echo -e "${YELLOW}选项:${NC}"
    echo -e "  --basic      仅安装基础工具 (curl, jq, python3)"
    echo -e "  --stress     仅安装压力测试工具 (wrk, ab)"
    echo -e "  --websocket  仅安装 WebSocket 测试工具 (websocat)"
    echo -e "  --all        安装所有工具 (默认)"
    echo -e "  --check      仅检查工具状态，不安装"
    echo -e "  --help       显示此帮助信息"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo -e "  $0              # 安装所有工具"
    echo -e "  $0 --basic      # 仅安装基础工具"
    echo -e "  $0 --stress     # 仅安装压力测试工具"
    echo -e "  $0 --websocket  # 仅安装 WebSocket 测试工具"
    echo -e "  $0 --check      # 检查工具状态"
    echo ""
}

# 解析命令行参数
parse_args() {
    INSTALL_TYPE="all"
    CHECK_ONLY=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --basic)
                INSTALL_TYPE="basic"
                shift
                ;;
            --stress)
                INSTALL_TYPE="stress"
                shift
                ;;
            --websocket)
                INSTALL_TYPE="websocket"
                shift
                ;;
            --all)
                INSTALL_TYPE="all"
                shift
                ;;
            --check)
                CHECK_ONLY=true
                shift
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                echo -e "${RED}未知选项: $1${NC}"
                echo ""
                show_usage
                exit 1
                ;;
        esac
    done
}

# 检查系统环境
check_system() {
    echo -e "${CYAN}系统环境检查:${NC}"
    echo -e "  操作系统: $(uname -s) $(uname -r)"
    echo -e "  架构: $(uname -m)"
    echo -e "  Shell: $SHELL"
    echo -e "  用户: $(whoami)"
    echo ""

    # 检查包管理器
    if command -v brew >/dev/null 2>&1; then
        echo -e "  包管理器: ${GREEN}Homebrew${NC}"
    elif command -v apt >/dev/null 2>&1; then
        echo -e "  包管理器: ${GREEN}APT (Debian/Ubuntu)${NC}"
    elif command -v yum >/dev/null 2>&1; then
        echo -e "  包管理器: ${GREEN}YUM (CentOS/RHEL)${NC}"
    elif command -v dnf >/dev/null 2>&1; then
        echo -e "  包管理器: ${GREEN}DNF (Fedora)${NC}"
    else
        echo -e "  包管理器: ${YELLOW}未知${NC}"
    fi

    # 检查 Node.js
    if command -v npm >/dev/null 2>&1; then
        echo -e "  Node.js: ${GREEN}已安装${NC}"
    else
        echo -e "  Node.js: ${YELLOW}未安装${NC}"
    fi

    echo ""
}

# 检查工具状态
check_tools_status() {
    echo -e "${MAGENTA}工具状态检查:${NC}"
    echo ""

    # 基础工具
    echo -e "${CYAN}基础工具:${NC}"
    local basic_tools=(
        "curl:curl -V"
        "jq:jq --version"
        "python3:python3 --version"
    )

    for tool_info in "${basic_tools[@]}"; do
        IFS=':' read -r tool cmd <<< "$tool_info"
        if command -v "$tool" >/dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $tool"
            if [ -n "$cmd" ]; then
                echo -e "    版本: $(eval "$cmd" 2>/dev/null | head -1 || "未知")"
            fi
        else
            echo -e "  ${RED}✗${NC} $tool (未安装)"
        fi
    done

    echo ""

    # 压力测试工具
    echo -e "${CYAN}压力测试工具:${NC}"
    local stress_tools=(
        "wrk:wrk -V"
        "ab:ab -V"
    )

    local stress_available=false
    for tool_info in "${stress_tools[@]}"; do
        IFS=':' read -r tool cmd <<< "$tool_info"
        if command -v "$tool" >/dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $tool"
            stress_available=true
            if [ -n "$cmd" ]; then
                echo -e "    版本: $(eval "$cmd" 2>/dev/null | head -1 || "未知")"
            fi
        else
            echo -e "  ${RED}✗${NC} $tool (未安装)"
        fi
    done

    if [ "$stress_available" = false ]; then
        echo -e "  ${YELLOW}⚠ 缺少压力测试工具${NC}"
    fi

    echo ""

    # WebSocket 测试工具
    echo -e "${CYAN}WebSocket 测试工具:${NC}"
    if command -v websocat >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} websocat"
        echo -e "    版本: $(websocat --version 2>/dev/null | head -1 || "未知")"
    else
        echo -e "  ${RED}✗${NC} websocat (未安装)"
    fi

    echo ""
}

# 主函数
main() {
    show_banner

    # 解析参数
    parse_args "$@"

    # 检查系统环境
    check_system

    # 如果只是检查状态
    if [ "$CHECK_ONLY" = true ]; then
        check_tools_status
        echo -e "${GREEN}工具状态检查完成${NC}"
        exit 0
    fi

    # 检查安装器是否存在
    if [ ! -f "$INSTALLER_LIB" ]; then
        echo -e "${RED}错误: 找不到工具安装器 $INSTALLER_LIB${NC}"
        echo -e "${YELLOW}请确保文件存在且有执行权限${NC}"
        exit 1
    fi

    # 加载安装器
    source "$INSTALLER_LIB"

    # 显示即将安装的工具类型
    case "$INSTALL_TYPE" in
        "basic")
            echo -e "${YELLOW}即将安装基础工具...${NC}"
            echo -e "  - curl (HTTP 客户端)"
            echo -e "  - jq (JSON 处理)"
            echo -e "  - python3 (脚本支持)"
            ;;
        "stress")
            echo -e "${YELLOW}即将安装压力测试工具...${NC}"
            echo -e "  - wrk (HTTP 压力测试)"
            echo -e "  - ab (Apache Bench)"
            ;;
        "websocket")
            echo -e "${YELLOW}即将安装 WebSocket 测试工具...${NC}"
            echo -e "  - websocat (WebSocket 客户端)"
            ;;
        "all"|*)
            echo -e "${YELLOW}即将安装所有测试工具...${NC}"
            echo -e "  - curl, jq, python3 (基础工具)"
            echo -e "  - wrk, ab (压力测试工具)"
            echo -e "  - websocat (WebSocket 测试工具)"
            ;;
    esac

    echo ""
    echo -e "${CYAN}开始安装...${NC}"
    echo ""

    # 执行安装
    install_required_tools "$INSTALL_TYPE"

    # 安装后检查
    echo -e "${CYAN}安装完成，正在验证...${NC}"
    echo ""
    check_tools_status

    # 显示完成信息
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   安装完成${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""
    echo -e "${GREEN}🎉 测试工具安装完成！${NC}"
    echo -e "${GREEN}✅ 现在可以运行 E2E 测试了${NC}"
    echo ""
    echo -e "${CYAN}运行测试示例:${NC}"
    echo -e "  ./tests/integration/e2e_test.sh"
    echo -e "  ./tests/integration/run_all_e2e_tests.sh"
    echo ""
}

# 错误处理
trap 'echo -e "\n${RED}安装被中断${NC}"; exit 1' INT TERM

# 执行主函数
main "$@"