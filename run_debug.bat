@echo off
echo 正在启动开发模式...

if not exist frontend/node_modules (
    echo 安装前端依赖...
    cd frontend && npm install && cd ..
)

echo 启动Wails开发服务器...
wails dev