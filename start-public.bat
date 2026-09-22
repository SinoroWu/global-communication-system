@echo off
chcp 65001 >nul
title 全渠道通訊整合系統 - 公開網際網路發布 (Public Tunnel)

echo ========================================================
echo   全渠道通訊整合系統 (Global Communication System)
echo   公開網際網路連線 (Cloudflare HTTPS Public Tunnel)
echo ========================================================
echo.

cd /d "%~dp0"

echo [1/2] 正在檢查後端伺服器...
netstat -ano | findstr :8080 >nul
if %errorlevel% neq 0 (
    echo [*] 正在啟動後端伺服器...
    start "Go Backend Server" cmd /c "cd backend && bin\server.exe"
    timeout /t 2 /nobreak >nul
) else (
    echo [*] 後端伺服器已在運行中 (:8080)。
)

echo.
echo [2/2] 正在啟動 Cloudflare 公網加密通道...
echo [*] 系統將生成一組全球可存取的公開 HTTPS 網址 (https://xxxx.trycloudflare.com)
echo [*] 任何人皆可直接透過該網址訪問客服系統與介接 LINE/IG/FB/抖音/TG Webhook！
echo.
tools\cloudflared.exe tunnel --url http://localhost:8080
pause
