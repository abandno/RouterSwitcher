@echo off
echo 构建前端资源...
cd frontend && npm install && npm run build && cd ..

echo 构建Windows可执行文件...
wails build

echo 构建完成!