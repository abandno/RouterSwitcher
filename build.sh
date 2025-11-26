#!/bin/bash

echo "构建前端资源..."
cd frontend && npm install && npm run build && cd ..

echo "构建可执行文件..."
wails build -tags webkit2_41

echo "构建完成!"