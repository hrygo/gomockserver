#!/bin/bash

# Docker镜像构建脚本
# 支持构建后端和完整栈镜像，包含Redis环境配置

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
PROJECT_NAME="mockserver"
VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo 'dev')}"
BUILD_TIME="${BUILD_TIME:-$(date -u +'%Y-%m-%dT%H:%M:%SZ')}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"

# 构建参数
DOCKER_REGISTRY="${DOCKER_REGISTRY:-}"
CACHE="${CACHE:-}"
NO_CACHE="${NO_CACHE:-false}"
PUSH="${PUSH:-false}"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示横幅
show_banner() {
    echo -e "${CYAN}=========================================="
    echo -e "${CYAN}     MockServer Docker 构建工具"
    echo -e "${CYAN}=========================================="
    echo ""
    echo -e "${CYAN}版本信息:${NC}"
    echo -e "  版本: ${YELLOW}$VERSION${NC}"
    echo -e "  构建时间: ${YELLOW}$BUILD_TIME${NC}"
    echo -e "  Git提交: ${YELLOW}$GIT_COMMIT${NC}"
    echo ""
    echo -e "${CYAN}构建选项:${NC}"
    echo -e "  Docker Registry: ${YELLOW}${DOCKER_REGISTRY:-default}${NC}"
    echo -e "  缓存: ${YELLOW}${CACHE:-enabled}${NC}"
    echo -e "  无缓存: ${YELLOW}${NO_CACHE}${NC}"
    echo -e "  推送镜像: ${YELLOW}${PUSH}${NC}"
    echo ""
}

# 检查Docker环境
check_docker() {
    log_info "检查Docker环境..."

    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker未安装或不在PATH中"
        exit 1
    fi

    if ! docker info >/dev/null 2>&1; then
        log_error "无法连接到Docker守护进程"
        exit 1
    fi

    log_success "Docker环境检查通过"
}

# 构建参数
prepare_build_args() {
    local build_args=""

    build_args="$build_args --build-arg VERSION=$VERSION"
    build_args="$build_args --build-arg BUILD_TIME=$BUILD_TIME"
    build_args="$build_args --build-arg GIT_COMMIT=$GIT_COMMIT"

    if [ "$NO_CACHE" = "true" ]; then
        build_args="$build_args --no-cache"
    fi

    if [ -n "$CACHE" ]; then
        build_args="$build_args --cache-from $CACHE"
    fi

    echo "$build_args"
}

# 构建后端镜像
build_backend() {
    log_info "构建后端Docker镜像..."

    local image_name="${DOCKER_REGISTRY}${PROJECT_NAME}:${VERSION}"
    local build_args=$(prepare_build_args)

    log_info "构建镜像: $image_name"

    if docker build $build_args -f docker/Dockerfile -t "$image_name" .; then
        log_success "后端镜像构建成功: $image_name"

        if [ "$PUSH" = "true" ]; then
            log_info "推送后端镜像..."
            if docker push "$image_name"; then
                log_success "后端镜像推送成功"
            else
                log_error "后端镜像推送失败"
                return 1
            fi
        fi
    else
        log_error "后端镜像构建失败"
        return 1
    fi
}

# 构建完整栈镜像
build_fullstack() {
    log_info "构建完整栈Docker镜像（包含前端）..."

    local image_name="${DOCKER_REGISTRY}${PROJECT_NAME}-fullstack:${VERSION}"
    local build_args=$(prepare_build_args)

    log_info "构建镜像: $image_name"

    if docker build $build_args -f docker/Dockerfile.fullstack -t "$image_name" .; then
        log_success "完整栈镜像构建成功: $image_name"

        # 显示镜像信息
        log_info "镜像信息:"
        docker images "$image_name" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

        if [ "$PUSH" = "true" ]; then
            log_info "推送完整栈镜像..."
            if docker push "$image_name"; then
                log_success "完整栈镜像推送成功"
            else
                log_error "完整栈镜像推送失败"
                return 1
            fi
        fi
    else
        log_error "完整栈镜像构建失败"
        return 1
    fi
}

# 构建测试运行器镜像
build_test_runner() {
    log_info "构建测试运行器Docker镜像..."

    local image_name="${DOCKER_REGISTRY}${PROJECT_NAME}-test:${VERSION}"

    if docker build -f docker/Dockerfile.test-runner -t "$image_name" .; then
        log_success "测试运行器镜像构建成功: $image_name"
    else
        log_error "测试运行器镜像构建失败"
        return 1
    fi
}

# 验证镜像
verify_images() {
    log_info "验证构建的镜像..."

    local images=("$PROJECT_NAME:$VERSION" "${PROJECT_NAME}-fullstack:$VERSION")

    for image in "${images[@]}"; do
        if docker images "$image" --format "{{.Repository}}:{{.Tag}}" | grep -q "$image"; then
            log_success "镜像存在: $image"

            # 运行健康检查（如果有）
            local image_id=$(docker images "$image" --format "{{.ID}}")
            if docker inspect "$image_id" 2>/dev/null | grep -q "HealthCheck"; then
                log_info "镜像 $image 包含健康检查配置"
            fi
        else
            log_warning "镜像不存在: $image"
        fi
    done
}

# 生成构建报告
generate_report() {
    local report_file="docker-build-report-${VERSION}.txt"

    log_info "生成构建报告: $report_file"

    {
        echo "=========================================="
        echo "MockServer Docker 构建报告"
        echo "=========================================="
        echo ""
        echo "构建时间: $(date)"
        echo "版本: $VERSION"
        echo "Git提交: $GIT_COMMIT"
        echo "构建主机: $(hostname)"
        echo "Docker版本: $(docker --version)"
        echo ""
        echo "构建的镜像:"
        docker images --filter "reference=${PROJECT_NAME}*" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
        echo ""
        echo "镜像标签:"
        if [ -n "$DOCKER_REGISTRY" ]; then
            echo "- ${DOCKER_REGISTRY}${PROJECT_NAME}:$VERSION"
            echo "- ${DOCKER_REGISTRY}${PROJECT_NAME}-fullstack:$VERSION"
        else
            echo "- ${PROJECT_NAME}:$VERSION"
            echo "- ${PROJECT_NAME}-fullstack:$VERSION"
        fi
        echo ""
        echo "构建日志请查看控制台输出"
        echo "=========================================="
    } > "$report_file"

    log_success "构建报告已生成: $report_file"
}

# 清理旧镜像
cleanup_old_images() {
    log_info "清理旧的Docker镜像..."

    # 删除悬空镜像
    local dangling_images=$(docker images -f "dangling=true" -q)
    if [ -n "$dangling_images" ]; then
        log_info "删除 $echo "$dangling_images" | wc -w 个悬空镜像"
        docker rmi $dangling_images 2>/dev/null || true
    fi

    # 删除旧版本的镜像（保留最近3个）
    local old_images=$(docker images "${PROJECT_NAME}*" --format "{{.Repository}}:{{.Tag}}" | sort -V | head -n -4)
    if [ -n "$old_images" ]; then
        log_info "删除旧版本镜像..."
        echo "$old_images" | xargs docker rmi 2>/dev/null || true
    fi

    log_success "镜像清理完成"
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -v, --version VERSION    设置版本标签"
    echo "  -r, --registry REGISTRY   设置Docker registry"
    " echo "  -c, --cache CACHE        设置构建缓存"
    echo "  --no-cache               禁用构建缓存"
    echo "  -p, --push               构建后推送镜像"
    echo "  -b, --backend            仅构建后端镜像"
    echo "  -f, --fullstack           构建完整栈镜像（默认）"
    echo "  -t, --test                构建测试镜像"
    echo "  --cleanup               构建后清理旧镜像"
    echo "  --verify                构建后验证镜像"
    echo "  -h, --help               显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                      # 构建完整栈镜像"
    echo "  $0 -v 1.0.0 -p         # 构建版本1.0.0并推送"
    echo "  $0 --backend --push    # 构建后端镜像并推送"
    echo "  $0 --no-cache --cleanup # 无缓存构建并清理"
    echo ""
}

# 主函数
main() {
    local build_type="fullstack"

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -r|--registry)
                DOCKER_REGISTRY="$2"
                shift 2
                ;;
            -c|--cache)
                CACHE="$2"
                shift 2
                ;;
            --no-cache)
                NO_CACHE="true"
                shift
                ;;
            -p|--push)
                PUSH="true"
                shift
                ;;
            -b|--backend)
                build_type="backend"
                shift
                ;;
            -f|--fullstack)
                build_type="fullstack"
                shift
                ;;
            -t|--test)
                build_type="test"
                shift
                ;;
            --cleanup)
                cleanup_old_images
                exit 0
                ;;
            --verify)
                verify_images
                exit 0
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 显示横幅
    show_banner

    # 检查Docker环境
    check_docker

    # 执行构建
    case $build_type in
        backend)
            build_backend
            ;;
        fullstack)
            build_fullstack
            ;;
        test)
            build_test_runner
            ;;
        *)
            log_error "未知构建类型: $build_type"
            exit 1
            ;;
    esac

    # 验证镜像
    verify_images

    # 生成报告
    generate_report

    # 清理旧镜像
    if [ "$NO_CACHE" = "true" ]; then
        cleanup_old_images
    fi

    echo ""
    echo -e "${GREEN}🎉 Docker镜像构建完成！${NC}"
    echo ""
    echo -e "${CYAN}使用示例:${NC}"
    echo -e "  docker run -p 8080:8080 -p 9090:9090 ${DOCKER_REGISTRY}${PROJECT_NAME}-fullstack:${VERSION}"
    echo -e ""
}

# 运行主函数
main "$@"