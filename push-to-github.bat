@echo off
chcp 65001 >nul
title 全渠道通訊整合系統 - 推送程式碼至 GitHub

echo ========================================================
echo   推送全渠道通訊整合系統至 GitHub (用於 Render 部署)
echo ========================================================
echo.
echo 步驟說明：
echo 1. 請先前往 https://github.com/new 建立一個新的 GitHub 倉庫 (Repository)。
echo 2. 複製倉庫的 Git 網址 (例如: https://github.com/your-username/global-comm.git)。
echo.
set /p REPO_URL="https://github.com/SinoroWu/global-communication-system.git "

if "%REPO_URL%"=="" (
    echo [!] 未輸入網址，程序已取消。
    pause
    exit /b
)

echo.
echo [*] 正在關聯遠端倉庫與推送 main 分支...
git remote remove origin 2>nul
git remote add origin %REPO_URL%
git branch -M main
git push -u origin main

if %errorlevel% equ 0 (
    echo.
    echo ========================================================
    echo   [V] 程式碼已成功推送到 GitHub！
    echo   接下來請登入 https://dashboard.render.com
    echo   點擊 [New +] -^> [Web Service] 並選擇該倉庫進行部署！
    echo ========================================================
) else (
    echo.
    echo [!] 推送失敗，請確認網址正確且已登入 GitHub 憑證。
)
pause
