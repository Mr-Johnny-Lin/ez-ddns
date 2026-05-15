#!/bin/bash
# Copyright (c) 2026 Johnny Lin (林展毅)
# 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
# 未经许可不得以本项目名义进行商业宣传

set -e

TAG_NAME="$1"
OUTPUT_FILE="${2:-CHANGELOG.md}"

echo "Generating changelog for tag $TAG_NAME..."

PREVIOUS_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")

if [ -z "$PREVIOUS_TAG" ]; then
  echo "This is the first release. Getting all commits..."
  CHANGELOG=$(git log --pretty=format:"- %s (%h)" --no-merges)
else
  echo "Getting commits from $PREVIOUS_TAG to $TAG_NAME..."
  CHANGELOG=$(git log --pretty=format:"- %s (%h)" --no-merges ${PREVIOUS_TAG}..HEAD)
fi

FEATS=$(echo "$CHANGELOG" | grep -E "^- (feat|feature)" || true)
FIXES=$(echo "$CHANGELOG" | grep -E "^- (fix|bugfix)" || true)
DOCS=$(echo "$CHANGELOG" | grep -E "^- (docs|doc)" || true)
STYLES=$(echo "$CHANGELOG" | grep -E "^- (style|styles)" || true)
REFACTOR=$(echo "$CHANGELOG" | grep -E "^- (refactor|ref)" || true)
PERF=$(echo "$CHANGELOG" | grep -E "^- (perf|performance)" || true)
TEST=$(echo "$CHANGELOG" | grep -E "^- (test|tests)" || true)
CHORE=$(echo "$CHANGELOG" | grep -E "^- (chore|ci|build)" || true)

echo "## 🚀 更新内容" > "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

HAS_CONTENT=false

if [ ! -z "$FEATS" ]; then
  echo "### ✨ 新功能" >> "$OUTPUT_FILE"
  echo "$FEATS" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$FIXES" ]; then
  echo "### 🐛 修复" >> "$OUTPUT_FILE"
  echo "$FIXES" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$DOCS" ]; then
  echo "### 📝 文档" >> "$OUTPUT_FILE"
  echo "$DOCS" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$STYLES" ]; then
  echo "### 💄 样式" >> "$OUTPUT_FILE"
  echo "$STYLES" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$REFACTOR" ]; then
  echo "### ♻️ 重构" >> "$OUTPUT_FILE"
  echo "$REFACTOR" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$PERF" ]; then
  echo "### ⚡ 性能优化" >> "$OUTPUT_FILE"
  echo "$PERF" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$TEST" ]; then
  echo "### ✅ 测试" >> "$OUTPUT_FILE"
  echo "$TEST" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ ! -z "$CHORE" ]; then
  echo "### 🔧 其他" >> "$OUTPUT_FILE"
  echo "$CHORE" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  HAS_CONTENT=true
fi

if [ "$HAS_CONTENT" = false ]; then
  echo "## 🚀 更新内容" > "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  echo "$CHANGELOG" >> "$OUTPUT_FILE"
fi

echo "Generated $OUTPUT_FILE:"
cat "$OUTPUT_FILE"