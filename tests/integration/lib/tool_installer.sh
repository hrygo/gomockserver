#!/bin/bash

# 自动工具安装器
# 检测并安装测试所需的工具

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 工具安装配置
TOOLS_CONFIG=(
    # 工具名:安装命令:检测命令:平台:包管理器
    "curl:brew install curl:curl -V:macos:brew"
    "curl:sudo apt-get update && sudo apt-get install -y curl:curl -V:linux:apt"
    "jq:brew install jq:jq --version:macos:brew"
    "jq:sudo apt-get install -y jq:jq --version:linux:apt"
    "wrk:brew install wrk:wrk -V:macos:brew"
    "wrk:sudo apt-get install -y wrk:wrk -V:linux:apt"
    "ab:sudo apt-get install -y apache2-utils:ab -V:linux:apt"
    "websocat:npm install -g websocat:websocat --version:any:npm"
    "python3:brew install python3:python3 --version:macos:brew"
    "python3:sudo apt-get install -y python3:python3 --version:linux:apt"
)

# 全局变量
TOOLS_INSTALLED=0
TOOLS_FAILED=0

# 平台检测
detect_platform() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "unknown"
    fi
}

# 包管理器检测
detect_package_manager() {
    local platform="$1"

    case "$platform" in
        "macos")
            if command -v brew >/dev/null 2>&1; then
                echo "brew"
            else
                echo "homebrew_missing"
            fi
            ;;
        "linux")
            if command -v apt >/dev/null 2>&1; then
                echo "apt"
            elif command -v yum >/dev/null 2>&1; then
                echo "yum"
            elif command -v dnf >/dev/null 2>&1; then
                echo "dnf"
            else
                echo "unknown"
            fi
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# 检查工具是否已安装
check_tool() {
    local tool="$1"
    local check_cmd="$2"

    if command -v "$tool" >/dev/null 2>&1; then
        if [ -n "$check_cmd" ]; then
            eval "$check_cmd" >/dev/null 2>&1
        else
            return 0
        fi
    else
        return 1
    fi
}

# 安装工具
install_tool() {
    local tool="$1"
    local install_cmd="$2"
    local platform="$3"
    local package_manager="$4"

    echo -e "${YELLOW}[安装] $tool${NC}"

    # 特殊处理 Homebrew 安装
    if [[ "$platform" == "macos" ]] && [[ "$package_manager" == "brew" ]]; then
        if ! command -v brew >/dev/null 2>&1; then
            echo -e "${CYAN}正在安装 Homebrew...${NC}"
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" || {
                echo -e "${RED}Homebrew 安装失败${NC}"
                return 1
            }
            # 重新加载环境变量
            eval "$(/opt/homebrew/bin/brew shellenv)"
        fi
    fi

    # 执行安装命令
    echo -e "${CYAN}执行: $install_cmd${NC}"
    if eval "$install_cmd"; then
        echo -e "${GREEN}✓ $tool 安装成功${NC}"
        return 0
    else
        echo -e "${RED}✗ $tool 安装失败${NC}"
        return 1
    fi
}

# 检查并安装单个工具
check_and_install_tool() {
    local tool_info="$1"
    IFS=':' read -r tool install_cmd check_cmd platform package_manager <<< "$tool_info"

    local current_platform=$(detect_platform)

    # 跳过不匹配平台的工具
    if [[ "$platform" != "$current_platform" ]] && [[ "$platform" != "any" ]]; then
        return 0
    fi

    # 检查工具是否已安装
    if check_tool "$tool" "$check_cmd"; then
        echo -e "${GREEN}✓ $tool 已安装${NC}"
        return 0
    fi

    # 获取当前包管理器
    local current_manager=$(detect_package_manager "$current_platform")

    # 检查包管理器是否匹配
    if [[ "$package_manager" == "$current_manager" ]] || [[ "$package_manager" == "any" ]]; then
        if install_tool "$tool" "$install_cmd" "$current_platform" "$current_manager"; then
            TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
            return 0
        else
            TOOLS_FAILED=$((TOOLS_FAILED + 1))
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ 跳过 $tool (包管理器不匹配: 需要 $package_manager, 当前 $current_manager)${NC}"
        return 0
    fi
}

# 安装压力测试工具
install_stress_tools() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   安装压力测试工具${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""

    local platform=$(detect_platform)
    echo -e "${CYAN}检测到平台: $platform${NC}"

    local manager=$(detect_package_manager "$platform")
    echo -e "${CYAN}检测到包管理器: $manager${NC}"
    echo ""

    # 安装 wrk (首选)
    echo -e "${YELLOW}优先安装 wrk (推荐的压力测试工具)${NC}"
    if [[ "$platform" == "macos" ]]; then
        check_and_install_tool "wrk:brew install wrk:wrk -V:macos:brew"
    elif [[ "$platform" == "linux" ]]; then
        if [[ "$manager" == "apt" ]]; then
            check_and_install_tool "wrk:sudo apt-get install -y wrk:wrk -V:linux:apt"
        elif [[ "$manager" == "yum" ]]; then
            check_and_install_tool "wrk:sudo yum install -y wrk:wrk -V:linux:yum"
        elif [[ "$manager" == "dnf" ]]; then
            check_and_install_tool "wrk:sudo dnf install -y wrk:wrk -V:linux:dnf"
        else
            echo -e "${YELLOW}⚠ 无法自动安装 wrk，请手动安装${NC}"
        fi
    fi

    # 安装 ab (备用方案)
    echo -e "${YELLOW}安装 ab (Apache Bench，备用压力测试工具)${NC}"
    if [[ "$platform" == "linux" ]]; then
        if [[ "$manager" == "apt" ]]; then
            check_and_install_tool "ab:sudo apt-get install -y apache2-utils:ab -V:linux:apt"
        elif [[ "$manager" == "yum" ]]; then
            check_and_install_tool "ab:sudo yum install -y httpd-tools:ab -V:linux:yum"
        elif [[ "$manager" == "dnf" ]]; then
            check_and_install_tool "ab:sudo dnf install -y httpd-tools:ab -V:linux:dnf"
        fi
    fi

    echo ""
}

# 安装 WebSocket 测试工具
install_websocket_tools() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   安装 WebSocket 测试工具${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""

    # 检查 npm
    if command -v npm >/dev/null 2>&1; then
        check_and_install_tool "websocat:npm install -g websocat:websocat --version:any:npm"

        # 修复 npm 全局包的符号链接问题（特别是 macOS 上）
        if ! command -v websocat >/dev/null 2>&1; then
            echo -e "${YELLOW}修复 websocat 符号链接...${NC}"
            local npm_prefix=$(npm config get prefix 2>/dev/null || echo "")
            if [ -n "$npm_prefix" ] && [ -f "$npm_prefix/lib/node_modules/websocat/websocat_mac" ]; then
                mkdir -p "$npm_prefix/bin"
                ln -sf "$npm_prefix/lib/node_modules/websocat/websocat_mac" "$npm_prefix/bin/websocat" 2>/dev/null || true
            fi
        fi
    else
        echo -e "${YELLOW}⚠ npm 未安装，跳过 websocat 安装${NC}"
        echo -e "${CYAN}请先安装 Node.js 和 npm，然后运行: npm install -g websocat${NC}"
    fi

    echo ""
}

# 安装基础工具
install_basic_tools() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   安装基础工具${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""

    local platform=$(detect_platform)

    # 安装 curl (所有平台都需要)
    if ! check_tool "curl" "curl -V"; then
        echo -e "${YELLOW}curl 未安装，正在尝试安装...${NC}"
        if [[ "$platform" == "macos" ]]; then
            # macOS 通常自带 curl，这里只是备用方案
            echo -e "${YELLOW}macOS 系统应该自带 curl，请检查系统配置${NC}"
        elif [[ "$platform" == "linux" ]]; then
            local manager=$(detect_package_manager "$platform")
            if [[ "$manager" == "apt" ]]; then
                check_and_install_tool "curl:sudo apt-get update && sudo apt-get install -y curl:curl -V:linux:apt"
            elif [[ "$manager" == "yum" ]]; then
                check_and_install_tool "curl:sudo yum install -y curl:curl -V:linux:yum"
            elif [[ "$manager" == "dnf" ]]; then
                check_and_install_tool "curl:sudo dnf install -y curl:curl -V:linux:dnf"
            fi
        fi
    else
        echo -e "${GREEN}✓ curl 已安装${NC}"
    fi

    # 安装 jq (JSON 处理工具)
    if ! check_tool "jq" "jq --version"; then
        echo -e "${YELLOW}安装 jq (JSON 处理工具)${NC}"
        if [[ "$platform" == "macos" ]]; then
            check_and_install_tool "jq:brew install jq:jq --version:macos:brew"
        elif [[ "$platform" == "linux" ]]; then
            local manager=$(detect_package_manager "$platform")
            if [[ "$manager" == "apt" ]]; then
                check_and_install_tool "jq:sudo apt-get install -y jq:jq --version:linux:apt"
            elif [[ "$manager" == "yum" ]]; then
                check_and_install_tool "jq:sudo yum install -y jq:jq --version:linux:yum"
            elif [[ "$manager" == "dnf" ]]; then
                check_and_install_tool "jq:sudo dnf install -y jq:jq --version:linux:dnf"
            fi
        fi
    else
        echo -e "${GREEN}✓ jq 已安装${NC}"
    fi

    # 安装 python3
    if ! check_tool "python3" "python3 --version"; then
        echo -e "${YELLOW}安装 python3${NC}"
        if [[ "$platform" == "macos" ]]; then
            check_and_install_tool "python3:brew install python3:python3 --version:macos:brew"
        elif [[ "$platform" == "linux" ]]; then
            local manager=$(detect_package_manager "$platform")
            if [[ "$manager" == "apt" ]]; then
                check_and_install_tool "python3:sudo apt-get install -y python3:python3 --version:linux:apt"
            elif [[ "$manager" == "yum" ]]; then
                check_and_install_tool "python3:sudo yum install -y python3:python3 --version:linux:yum"
            elif [[ "$manager" == "dnf" ]]; then
                check_and_install_tool "python3:sudo dnf install -y python3:python3 --version:linux:dnf"
            fi
        fi
    else
        echo -e "${GREEN}✓ python3 已安装${NC}"
    fi

    echo ""
}

# 显示安装结果
show_installation_results() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   工具安装结果${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""

    echo -e "${CYAN}安装统计:${NC}"
    echo -e "  新安装工具: ${GREEN}$TOOLS_INSTALLED${NC}"
    echo -e "  安装失败: ${RED}$TOOLS_FAILED${NC}"

    if [ $TOOLS_FAILED -eq 0 ]; then
        echo ""
        echo -e "${GREEN}🎉 所有工具安装成功！${NC}"
        echo -e "${GREEN}✅ 系统准备就绪，可以运行 E2E 测试${NC}"
    else
        echo ""
        echo -e "${YELLOW}⚠ 部分工具安装失败${NC}"
        echo -e "${YELLOW}💡 建议手动安装失败的工具${NC}"
    fi

    echo ""
}

# 主要安装函数
install_required_tools() {
    local install_type="$1"  # "basic", "stress", "websocket", "all"

    echo -e "${CYAN}开始安装测试所需工具...${NC}"
    echo ""

    case "$install_type" in
        "basic")
            install_basic_tools
            ;;
        "stress")
            install_stress_tools
            ;;
        "websocket")
            install_websocket_tools
            ;;
        "all"|*)
            install_basic_tools
            install_stress_tools
            install_websocket_tools
            ;;
    esac

    show_installation_results
}

# 静默安装（不输出）
install_required_tools_silent() {
    local install_type="$1"

    # 重定向输出到 /dev/null
    {
        case "$install_type" in
            "basic")
                install_basic_tools
                ;;
            "stress")
                install_stress_tools
                ;;
            "websocket")
                install_websocket_tools
                ;;
            "all"|*)
                install_basic_tools
                install_stress_tools
                install_websocket_tools
                ;;
        esac
    } >/dev/null 2>&1

    return $([ $TOOLS_FAILED -eq 0 ])
}

# 检查工具是否就绪
check_tools_ready() {
    local required_tools="$1"  # "basic", "stress", "websocket", "all"

    local missing_tools=()

    case "$required_tools" in
        "basic")
            ! check_tool "curl" "curl -V" && missing_tools+=("curl")
            ! check_tool "python3" "python3 --version" && missing_tools+=("python3")
            ! check_tool "jq" "jq --version" && missing_tools+=("jq")
            ;;
        "stress")
            ! check_tool "wrk" "wrk -V" && ! check_tool "ab" "ab -V" && missing_tools+=("压力测试工具 (wrk 或 ab)")
            ;;
        "websocket")
            ! check_tool "websocat" "websocat --version" && missing_tools+=("websocat")
            ;;
        "all"|*)
            ! check_tool "curl" "curl -V" && missing_tools+=("curl")
            ! check_tool "python3" "python3 --version" && missing_tools+=("python3")
            ! check_tool "jq" "jq --version" && missing_tools+=("jq")
            ! check_tool "wrk" "wrk -V" && ! check_tool "ab" "ab -V" && missing_tools+=("压力测试工具 (wrk 或 ab)")
            ! check_tool "websocat" "websocat --version" && missing_tools+=("websocat")
            ;;
    esac

    if [ ${#missing_tools[@]} -eq 0 ]; then
        return 0
    else
        echo -e "${YELLOW}缺失的工具: ${missing_tools[*]}${NC}"
        return 1
    fi
}

# 导出函数
export -f detect_platform detect_package_manager check_tool install_tool
export -f check_and_install_tool install_stress_tools install_websocket_tools
export -f install_basic_tools show_installation_results
export -f install_required_tools install_required_tools_silent check_tools_ready

echo -e "${GREEN}工具安装器已加载${NC}"