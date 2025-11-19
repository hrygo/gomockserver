#!/bin/bash

# Docker健康检查脚本
# 用于检查Docker守护进程状态和容器健康状况

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 Docker Health Check${NC}"

# 检查Docker守护进程
check_docker_daemon() {
    if ! docker info >/dev/null 2>&1; then
        echo -e "${RED}❌ Docker daemon is not running${NC}"
        return 1
    fi
    echo -e "${GREEN}✅ Docker daemon is running${NC}"
    return 0
}

# 检查MongoDB容器
check_mongodb() {
    if docker ps --format '{{.Names}}' | grep -q '^mongodb$'; then
        echo -e "${GREEN}✅ MongoDB container is running${NC}"
        # 检查MongoDB是否可以连接
        if docker exec mongodb mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ MongoDB is ready${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠️ MongoDB container is running but not ready${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ MongoDB container is not running${NC}"
        return 1
    fi
}

# 检查Redis容器
check_redis() {
    if docker ps --format '{{.Names}}' | grep -q '^mockserver-redis$'; then
        echo -e "${GREEN}✅ Redis container is running${NC}"
        # 检查Redis是否可以连接
        if docker exec mockserver-redis redis-cli ping >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Redis is ready${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠️ Redis container is running but not ready${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Redis container is not running${NC}"
        return 1
    fi
}

# 主检查函数
main() {
    local failed=0

    check_docker_daemon || failed=1

    if [ $failed -eq 0 ]; then
        check_mongodb || failed=1
        check_redis || failed=1
    fi

    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 All Docker services are healthy!${NC}"
        exit 0
    else
        echo -e "${RED}💥 Some Docker services need attention${NC}"
        exit 1
    fi
}

# 如果直接运行此脚本
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi