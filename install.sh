#!/bin/bash

# Scrago 安装脚本
echo "🚀 开始安装 Scrago..."

# 检查 Go 是否已安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go 编译器"
    echo "请先安装 Go: https://golang.org/dl/"
    exit 1
fi

echo "✅ 检测到 Go 版本: $(go version)"

# 构建项目
echo "🔨 构建 scrago 命令行工具..."
if go build -o scrago ./cmd/scrago; then
    echo "✅ 构建成功！"
else
    echo "❌ 构建失败"
    exit 1
fi

# 创建安装目录
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# 复制可执行文件
echo "📦 安装 scrago 到 $INSTALL_DIR..."
cp scrago "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/scrago"

# 检查 PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠️  警告: $INSTALL_DIR 不在 PATH 中"
    echo "请将以下行添加到你的 shell 配置文件 (~/.bashrc, ~/.zshrc 等):"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
    echo "或者运行以下命令临时添加到 PATH:"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo "🎉 Scrago 安装完成！"
echo ""
echo "📖 使用方法:"
echo "  scrago help           # 显示帮助"
echo "  scrago list           # 列出可用爬虫"
echo "  scrago crawl <name>   # 运行爬虫"
echo "  scrago startproject <name>  # 创建新项目"
echo ""
echo "🔗 更多信息: https://github.com/your-repo/scrago"