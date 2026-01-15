#!/bin/bash
# Multi-architecture Docker build script for goPlayer
# Prerequisites:
#   - QEMU emulation: docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
#   - buildx builder: docker buildx create --use --name=crossplat

docker buildx build --platform linux/amd64 -t dunglex/goplayer-builder:x64 -f Dockerfile.x64 --push .
docker buildx build --platform linux/386 -t dunglex/goplayer-builder:x86 -f Dockerfile.x86 --push .
docker buildx build --platform linux/arm64 -t dunglex/goplayer-builder:arm64 -f Dockerfile.arm64 --push .
docker buildx build --platform linux/arm/v7 -t dunglex/goplayer-builder:armv7 -f Dockerfile.armv7 --push .
docker buildx imagetools create -t dunglex/goplayer-builder:latest dunglex/goplayer-builder:x64 dunglex/goplayer-builder:x86 dunglex/goplayer-builder:arm64 dunglex/goplayer-builder:armv7