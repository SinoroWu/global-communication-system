# ==========================================
# Stage 1: Build Frontend (Vue 3 + TS + Vite)
# ==========================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# ==========================================
# Stage 2: Build Backend (Go 1.24+ Native)
# ==========================================
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum* ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# ==========================================
# Stage 3: Ultra-light Production Runtime (~30MB)
# ==========================================
FROM alpine:3.20
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=backend-builder /app/server /app/server
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

ENV PORT=8080
ENV DB_PATH=/app/data/global_comm.db
VOLUME ["/app/data"]
EXPOSE 8080

CMD ["/app/server"]
