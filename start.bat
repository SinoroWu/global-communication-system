@echo off
chcp 65001 >nul
title 全渠道通訊整合系統 - Global Communication System

echo ========================================================
echo   全渠道通訊整合系統 (Global Communication System)
echo   整合 LINE | Instagram | Facebook | 抖音 | Telegram
echo ========================================================
echo.
echo [1/2] 正在啟動高效能 Go 原生後端 (埠號 :8080)...
cd /d "%~dp0backend"

if not exist "bin\server.exe" (
    echo [*] 正在編譯原生二進制執行檔...
    "C:\Program Files\Go\bin\go.exe" build -o bin\server.exe .\cmd\server
)

start "" http://localhost:8080
echo [*] 已為您在預設瀏覽器開啟: http://localhost:8080
echo.
.\bin\server.exe
pause
