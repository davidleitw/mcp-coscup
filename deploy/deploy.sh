#!/bin/bash

# MCP-COSCUP 自動部署腳本
# 用途: 自動建置、推送和部署到 Google Cloud Run

set -e  # 遇到錯誤立即停止

# 載入 .env 檔案（如果存在）
if [ -f ".env" ]; then
    echo "載入 .env 環境變數..."
    export $(grep -v '^#' .env | xargs)
elif [ -f "../.env" ]; then
    echo "載入 ../.env 環境變數..."
    export $(grep -v '^#' ../.env | xargs)
fi

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置變數 - 可透過環境變數覆蓋
PROJECT_ID="${PROJECT_ID:-your-project-id}"
REPOSITORY_NAME="${REPOSITORY_NAME:-mcp-coscup}"
IMAGE_NAME="${IMAGE_NAME:-mcp-coscup}"
REGION="${REGION:-asia-east1}"
SERVICE_NAME="${SERVICE_NAME:-mcp-coscup}"

# 檢查必要的環境變數
if [ "$PROJECT_ID" = "your-project-id" ]; then
    echo -e "${RED}[ERROR]${NC} 請設定 PROJECT_ID 環境變數"
    echo "範例: export PROJECT_ID=your-actual-project-id"
    echo "或者直接執行: PROJECT_ID=your-project-id $0"
    exit 1
fi

IMAGE_TAG="$REGION-docker.pkg.dev/$PROJECT_ID/$REPOSITORY_NAME/$IMAGE_NAME:latest"

# 日誌函數
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# 檢查必要工具
check_prerequisites() {
    log_step "檢查必要工具"
    
    # 檢查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安裝或不在 PATH 中"
        exit 1
    fi
    log_success "Docker 已安裝: $(docker --version)"
    
    # 檢查 gcloud
    if ! command -v gcloud &> /dev/null; then
        log_error "gcloud CLI 未安裝或不在 PATH 中"
        exit 1
    fi
    log_success "gcloud CLI 已安裝: $(gcloud --version | head -1)"
    
    # 檢查 Dockerfile
    if [ ! -f "Dockerfile" ]; then
        log_error "Dockerfile 不存在於當前目錄"
        exit 1
    fi
    log_success "Dockerfile 存在"
}

# 檢查 Google Cloud 認證和專案
check_gcloud_auth() {
    log_step "檢查 Google Cloud 認證和專案設定"
    
    # 檢查是否已登入
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" &> /dev/null; then
        log_error "未登入 Google Cloud，請執行: gcloud auth login"
        exit 1
    fi
    
    local active_account=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
    log_success "已登入帳號: $active_account"
    
    # 檢查當前專案
    local current_project=$(gcloud config get-value project)
    if [ "$current_project" != "$PROJECT_ID" ]; then
        log_warning "當前專案 ($current_project) 與設定專案 ($PROJECT_ID) 不符"
        log_info "設定專案為: $PROJECT_ID"
        gcloud config set project $PROJECT_ID
    fi
    log_success "當前專案: $(gcloud config get-value project)"
    
    # 檢查 Docker 認證
    log_info "檢查 Docker 認證設定..."
    if ! grep -q "asia-east1-docker.pkg.dev" ~/.docker/config.json 2>/dev/null; then
        log_warning "Docker 認證未設定，正在設定..."
        gcloud auth configure-docker asia-east1-docker.pkg.dev --quiet
    fi
    log_success "Docker 認證設定完成"
}

# 建置 Docker 映像
build_image() {
    log_step "建置 Docker 映像"
    
    log_info "開始建置映像: $IMAGE_TAG"
    log_info "建置可能需要幾分鐘，請耐心等待..."
    
    if docker build -t $IMAGE_TAG .; then
        log_success "映像建置完成"
    else
        log_error "映像建置失敗"
        exit 1
    fi
    
    # 驗證映像是否建置成功
    if docker images $IMAGE_TAG --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | grep -v "REPOSITORY"; then
        log_success "映像驗證成功"
    else
        log_error "映像驗證失敗"
        exit 1
    fi
}

# 推送映像到 Artifact Registry
push_image() {
    log_step "推送映像到 Artifact Registry"
    
    log_info "推送映像: $IMAGE_TAG"
    log_info "推送過程可能需要幾分鐘，請耐心等待..."
    
    if docker push $IMAGE_TAG; then
        log_success "映像推送完成"
    else
        log_error "映像推送失敗"
        exit 1
    fi
    
    # 驗證映像是否在 Artifact Registry 中
    log_info "驗證映像是否已上傳到 Artifact Registry..."
    sleep 2  # 給 Artifact Registry 一些時間同步
    
    # 使用更簡單的驗證方式
    if gcloud artifacts docker images list asia-east1-docker.pkg.dev/$PROJECT_ID/$REPOSITORY_NAME --limit=1 &> /dev/null; then
        log_success "映像已成功上傳到 Artifact Registry"
        # 顯示映像資訊
        gcloud artifacts docker images list asia-east1-docker.pkg.dev/$PROJECT_ID/$REPOSITORY_NAME --limit=1
    else
        log_warning "無法驗證映像上傳狀態，但推送似乎成功了"
        log_info "繼續部署流程..."
    fi
}

# 部署到 Cloud Run
deploy_to_cloud_run() {
    log_step "部署到 Cloud Run"
    
    log_info "部署服務: $SERVICE_NAME"
    log_info "映像: $IMAGE_TAG"
    log_info "區域: $REGION"
    
    if gcloud run deploy $SERVICE_NAME \
        --image=$IMAGE_TAG \
        --platform=managed \
        --region=$REGION \
        --allow-unauthenticated \
        --port=8080 \
        --memory=512Mi \
        --cpu=1 \
        --min-instances=0 \
        --max-instances=3 \
        --timeout=300 \
        --concurrency=80 \
        --quiet; then
        log_success "Cloud Run 部署完成"
    else
        log_error "Cloud Run 部署失敗"
        exit 1
    fi
}

# 驗證部署
verify_deployment() {
    log_step "驗證部署"
    
    # 取得服務 URL
    local service_url=$(gcloud run services describe $SERVICE_NAME \
        --region=$REGION \
        --format="value(status.url)")
    
    if [ -z "$service_url" ]; then
        log_error "無法取得服務 URL"
        exit 1
    fi
    
    log_success "服務 URL: $service_url"
    
    # 測試健康檢查端點
    log_info "測試健康檢查端點..."
    if curl -s -f "$service_url/health" > /dev/null; then
        log_success "健康檢查端點正常"
    else
        log_warning "健康檢查端點測試失敗，但服務可能仍在啟動中"
    fi
    
    # 測試 MCP 初始化端點
    log_info "測試 MCP 初始化端點..."
    local mcp_response=$(curl -s -X POST "$service_url/mcp/initialize" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}')
    
    if echo "$mcp_response" | grep -q "COSCUP Schedule Planner"; then
        log_success "MCP 端點正常運作"
    else
        log_warning "MCP 端點測試失敗，請手動檢查"
    fi
    
    # 顯示服務資訊
    log_info "取得服務詳細資訊..."
    gcloud run services describe $SERVICE_NAME --region=$REGION --format="table(
        metadata.name,
        status.url,
        status.conditions[0].type:label=READY,
        metadata.labels,
        spec.template.spec.containers[0].image
    )"
}

# 清理函數
cleanup() {
    log_info "清理暫存資源..."
    # 這裡可以添加清理邏輯，例如刪除舊的映像等
}

# 主要執行流程
main() {
    log_step "MCP-COSCUP 自動部署開始"
    
    # 顯示配置資訊
    log_info "部署配置:"
    log_info "  專案 ID: $PROJECT_ID"
    log_info "  映像標籤: $IMAGE_TAG"
    log_info "  服務名稱: $SERVICE_NAME"
    log_info "  區域: $REGION"
    
    # 執行部署步驟
    check_prerequisites
    check_gcloud_auth
    build_image
    push_image
    deploy_to_cloud_run
    verify_deployment
    
    log_step "部署完成"
    log_success "MCP-COSCUP 服務已成功部署到 Google Cloud Run!"
    
    # 顯示最終資訊
    local service_url=$(gcloud run services describe $SERVICE_NAME \
        --region=$REGION \
        --format="value(status.url)")
    
    echo -e "\n${GREEN}=== 部署資訊 ===${NC}"
    echo -e "服務 URL: ${BLUE}$service_url${NC}"
    echo -e "健康檢查: ${BLUE}$service_url/health${NC}"
    echo -e "MCP 端點: ${BLUE}$service_url/mcp/*${NC}"
    echo -e "\n可以開始使用你的 COSCUP MCP 服務了！"
}

# 錯誤處理
trap cleanup EXIT

# 執行主程式
main "$@"