# Mock Server Frontend

Mock Server 的 Web 管理界面，基于 React + TypeScript + Ant Design 构建。

## 技术栈

- **框架**: React 18.x
- **语言**: TypeScript 5.x
- **UI 组件**: Ant Design 5.x
- **构建工具**: Vite 5.x
- **路由**: React Router 6.x
- **状态管理**: Zustand 4.x
- **数据请求**: TanStack Query (React Query) 5.x
- **HTTP 客户端**: Axios 1.x
- **代码编辑器**: Monaco Editor
- **图表**: ECharts 5.x

## 开发环境要求

- Node.js >= 18.0.0
- npm >= 9.0.0

## 快速开始

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

开发服务器将运行在 `http://localhost:5173`

### 构建生产版本

```bash
npm run build
```

构建产物将输出到 `../dist` 目录（供后端静态托管使用）

### 预览生产构建

```bash
npm run preview
```

## 可用命令

- `npm run dev` - 启动开发服务器
- `npm run build` - 构建生产版本
- `npm run preview` - 预览生产构建
- `npm run lint` - 运行 ESLint 检查
- `npm run lint:fix` - 自动修复 ESLint 问题
- `npm run format` - 格式化代码（Prettier）
- `npm run test` - 运行单元测试
- `npm run test:coverage` - 运行测试并生成覆盖率报告

## 项目结构

```
src/
├── api/              # API 接口层
├── components/       # 通用组件
├── pages/            # 页面组件
├── hooks/            # 自定义 Hooks
├── store/            # 状态管理
├── types/            # TypeScript 类型定义
├── utils/            # 工具函数
├── styles/           # 样式文件
├── App.tsx           # 根组件
└── main.tsx          # 入口文件
```

## 开发规范

### 代码风格

- 使用 ESLint 进行代码检查
- 使用 Prettier 进行代码格式化
- 组件命名使用 PascalCase
- 函数命名使用 camelCase
- 常量命名使用 UPPER_SNAKE_CASE

### TypeScript

- 启用严格模式
- 明确定义所有类型
- 避免使用 `any` 类型

### 组件开发

- 优先使用函数组件和 Hooks
- 组件职责单一
- 提取可复用组件
- 使用 React.memo 优化性能

## API 代理配置

开发环境下，所有 `/api` 开头的请求会被代理到 `http://localhost:8080`，配置在 `vite.config.ts` 中。

## 环境变量

可以在项目根目录创建 `.env.local` 文件配置环境变量：

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 浏览器支持

- Chrome >= 87
- Firefox >= 78
- Safari >= 14
- Edge >= 88

## 常见问题

### 1. 安装依赖失败

尝试清除 npm 缓存：
```bash
npm cache clean --force
npm install
```

### 2. 开发服务器启动失败

检查端口 5173 是否被占用：
```bash
lsof -i :5173
```

### 3. API 请求失败

确保后端服务已启动在 `http://localhost:8080`

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 创建 Pull Request

## 许可证

MIT License
