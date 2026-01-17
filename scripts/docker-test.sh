#!/bin/bash
# Test against specific package manager versions using Docker
#
# Usage:
#   ./scripts/docker-test.sh                    # test all versions
#   ./scripts/docker-test.sh npm                # test all npm versions
#   ./scripts/docker-test.sh npm 10             # test npm 10 only
#   ./scripts/docker-test.sh --list             # list available versions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_DIR/docker"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

list_versions() {
    echo "Available versions:"
    echo ""
    for manager_dir in "$DOCKER_DIR"/*/; do
        manager=$(basename "$manager_dir")
        versions=""
        for dockerfile in "$manager_dir"Dockerfile.*; do
            if [ -f "$dockerfile" ]; then
                version=$(basename "$dockerfile" | sed 's/Dockerfile\.//')
                versions="$versions $version"
            fi
        done
        if [ -n "$versions" ]; then
            echo "  $manager:$versions"
        fi
    done
}

build_and_test() {
    local manager=$1
    local version=$2
    local dockerfile="$DOCKER_DIR/$manager/Dockerfile.$version"
    local image_name="managers-$manager-$version"

    if [ ! -f "$dockerfile" ]; then
        echo -e "${RED}Dockerfile not found: $dockerfile${NC}"
        return 1
    fi

    echo -e "${YELLOW}Building $image_name...${NC}"
    docker build -t "$image_name" -f "$dockerfile" "$PROJECT_DIR" || return 1

    echo -e "${YELLOW}Testing with $manager $version...${NC}"
    if docker run --rm "$image_name"; then
        echo -e "${GREEN}✓ $manager $version passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $manager $version failed${NC}"
        return 1
    fi
}

test_manager() {
    local manager=$1
    local specific_version=$2
    local manager_dir="$DOCKER_DIR/$manager"

    if [ ! -d "$manager_dir" ]; then
        echo -e "${RED}Unknown manager: $manager${NC}"
        return 1
    fi

    local failed=0

    if [ -n "$specific_version" ]; then
        build_and_test "$manager" "$specific_version" || failed=1
    else
        for dockerfile in "$manager_dir"/Dockerfile.*; do
            if [ -f "$dockerfile" ]; then
                version=$(basename "$dockerfile" | sed 's/Dockerfile\.//')
                build_and_test "$manager" "$version" || failed=1
            fi
        done
    fi

    return $failed
}

test_all() {
    local failed=0

    for manager_dir in "$DOCKER_DIR"/*/; do
        manager=$(basename "$manager_dir")
        test_manager "$manager" || failed=1
    done

    return $failed
}

# Main
case "${1:-}" in
    --list|-l)
        list_versions
        ;;
    --help|-h)
        echo "Usage: $0 [manager] [version]"
        echo ""
        echo "Options:"
        echo "  --list, -l    List available versions"
        echo "  --help, -h    Show this help"
        echo ""
        echo "Examples:"
        echo "  $0                    # test all versions"
        echo "  $0 npm                # test all npm versions"
        echo "  $0 npm 10             # test npm 10 only"
        ;;
    "")
        test_all
        ;;
    *)
        test_manager "$1" "${2:-}"
        ;;
esac
