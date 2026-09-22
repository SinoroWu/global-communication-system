@echo off
chcp 65001 >nul
title 全渠道通訊整合系統 - 開發模式 (Dev Mode)

echo ========================================================
echo   啟動全渠道通訊整合系統 - 前後端雙熱重載開發模式
echo ========================================================
echo.

cd /d "%~dp0"

echo [1/2] 正在背景啟動高效能 Go 後端 (:8080)...
start "Go Backend Server" cmd /c "cd backend && go run ./cmd/server"

echo [2/2] 正在啟動 Vite 前端開發伺服器 (:5173)...
cd frontend
npm run dev
