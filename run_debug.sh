#!/bin/bash

echo "正在启动开发模式..."

if [ ! -d "frontend/node_modules" ]; then
    echo "安装前端依赖..."
    cd frontend && npm install && cd ..
fi

echo "启动Wails开发服务器..."
wails dev