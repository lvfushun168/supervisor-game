# Supervisor Game

本项目是一个本地单机网页服务骨架：

- Go 单 exe 本地服务
- Gin HTTP API
- GORM + MySQL 8.0
- Vite + Vue 3 前端
- 生产构建后由 Go exe 直接托管 `frontend/dist`

## 快速开始

```bash
cp .env.example .env
docker compose up -d mysql
go run .
```

前端开发模式：

```bash
cd frontend
npm install
npm run dev
```

Vite 开发服务器会把 `/api` 代理到 `http://localhost:8080`。

## 构建单 exe

```bash
make build
./bin/supervisor-game
```

构建会先生成 `frontend/dist`，再编译 Go 程序并把前端静态资源嵌入可执行文件。

## 常用接口

- `GET /api/health`
- `GET /api/admin/status`
