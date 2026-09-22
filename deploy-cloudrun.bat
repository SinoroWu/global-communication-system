@echo off
chcp 65001 >nul
title 全渠道通訊整合系統 - 部署至 Google Cloud Run

echo ========================================================
echo   部署全渠道通訊整合系統至 Google Cloud Run (全球無伺服器託管)
echo ========================================================
echo.

cd /d "%~dp0"

echo 正在透過 Google Cloud SDK 構建容器並部署至 Cloud Run (台灣機房 asia-east1)...
gcloud run deploy global-comm-system --source . --region asia-east1 --allow-unauthenticated

if %errorlevel% equ 0 (
    echo.
    echo [*] 部署成功！您已取得永久運行的 Cloud Run 公開 HTTPS 網址！
) else (
    echo.
    echo [!] 部署過程中發生錯誤，請確認 GCP 專案已啟用 Cloud Run 與 Cloud Build API。
)
pause
