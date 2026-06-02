# 儿童商店管理后台 - 前端页面规划

## 1. 项目概述

- **项目名称**: 悠兰儿童商店管理后台 (Management Web Shop)
- **技术栈**: Vue 3 + Vite + TypeScript
- **UI 框架**: Element Plus
- **风格**: 简约现代
- **主色调**: 
  - 黑色 (#1a1a1a)
  - 白色 (#ffffff)
  - 浅蓝色 (#87CEEB / #B0E0E6)

## 2. 目录结构

```
Management_web_shop/
├── src/
│   ├── assets/          # 静态资源
│   │   └── styles/      # 全局样式
│   ├── components/      # 公共组件
│   │   ├── Header.vue
│   │   ├── Sidebar.vue
│   │   └── Footer.vue
│   ├── layout/         # 布局组件
│   │   └── MainLayout.vue
│   ├── views/          # 页面组件
│   │   ├── Login.vue          # 登录页
│   │   ├── Dashboard.vue      # 数据总览
│   │   ├── Order.vue          # 订单管理
│   │   ├── OrderDetail.vue    # 订单详情
│   │   ├── Product.vue        # 商品管理
│   │   ├── ProductDetail.vue   # 商品详情
│   │   └── Report.vue         # 报表管理
│   ├── router/         # 路由配置
│   │   └── index.ts
│   ├── stores/        # 状态管理
│   │   └── user.ts
│   ├── api/           # API 接口
│   │   └── index.ts
│   ├── utils/         # 工具函数
│   │   └── index.ts
│   ├── App.vue
│   └── main.ts
├── index.html
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 3. 页面详细规划

### 3.1 登录页面 (Login.vue)

**布局**:
- 简洁的登录卡片居中显示
- 左侧可添加简约装饰图案（可选）

**功能**:
- 用户名输入框
- 密码输入框
- 记住我复选框
- 登录按钮
- 错误提示信息

**视觉设计**:
- 背景: 白色 (#ffffff) 或极浅灰色 (#f5f5f5)
- 登录卡片: 白色，带轻微阴影
- 按钮: 浅蓝色 (#87CEEB) 主按钮
- 文字: 黑色 (#1a1a1a) 主文字

---

### 3.2 数据总览页面 (Dashboard.vue)

**布局**:
- 顶部统计卡片行（4个指标）
- 中间图表区域（销售额趋势图）
- 底部快捷操作区域

**功能**:
- 今日订单数
- 今日销售额
- 待处理订单
- 商品总数
- 销售额趋势图（折线图）
- 最近订单列表（5条）

**视觉设计**:
- 卡片: 白色背景，浅蓝色边框或阴影
- 数字: 黑色加粗
- 图表: 浅蓝色线条

---

### 3.3 订单管理页面 (Order.vue)

**布局**:
- 顶部搜索筛选区域
- 中间订单表格
- 底部分页组件

**功能**:
- 订单搜索（订单号、买家姓名）
- 订单状态筛选（全部、待付款、待发货、已完成、已取消）
- 订单列表表格（订单号、商品信息、买家、金额、状态、时间）
- 订单详情查看
- 订单状态更新（发货、取消）

**表格列**:
| 列名 | 说明 |
|------|------|
| 订单号 | 唯一标识 |
| 商品信息 | 商品名称和数量 |
| 买家 | 买家姓名 |
| 订单金额 | 金额 |
| 订单状态 | 状态标签 |
| 下单时间 | 时间 |
| 操作 | 查看/发货/取消 |

---

### 3.4 订单详情页面 (OrderDetail.vue)

**布局**:
- 页面顶部返回按钮
- 左侧订单信息卡片
- 右侧商品信息列表

**功能**:
- 查看订单基本信息（订单号、下单时间、支付方式）
- 查看买家信息（姓名、电话、地址）
- 查看商品列表（商品图片、名称、规格、数量、单价、小计）
- 查看订单金额明细（商品金额、运费、优惠、实付）
- 查看订单状态流转记录
- 订单操作（发货、取消、备注）

**信息卡片**:
| 模块 | 内容 |
|------|------|
| 订单信息 | 订单号、下单时间、支付方式、支付时间 |
| 买家信息 | 买家姓名、联系电话、收货地址 |
| 物流信息 | 物流公司、物流单号（发货后显示） |

**视觉设计**:
- 卡片: 白色背景，轻微阴影
- 状态标签: 待付款(橙色)、待发货(蓝色)、已完成(绿色)、已取消(灰色)

---

### 3.5 商品管理页面 (Product.vue)

**布局**:
- 顶部搜索筛选区域
- 中间商品表格
- 底部分页组件

**功能**:
- 商品搜索（商品名称、货号）
- 商品分类筛选
- 商品列表展示
- 商品图片、名称、货号、库存、价格、状态
- 商品上下架操作
- 商品编辑

**表格列**:
| 列名 | 说明 |
|------|------|
| 商品图片 | 缩略图 |
| 商品名称 | 商品标题 |
| 货号 | SKU |
| 库存 | 数量 |
| 价格 | 销售价 |
| 状态 | 上架/下架 |
| 操作 | 编辑/上下架/删除 |

---

### 3.6 商品详情页面 (ProductDetail.vue)

**布局**:
- 页面顶部返回按钮和操作按钮
- 商品基本信息区域
- 商品图片区域
- 商品库存/价格区域

**功能**:
- 查看/编辑商品基本信息（名称、货号、分类、描述）
- 查看/编辑商品图片（主图、详情图）
- 查看/编辑商品规格（颜色、尺码）
- 查看/编辑商品库存（各规格库存数量）
- 查看/编辑商品价格（销售价、原价）
- 查看商品状态（上架/下架/售罄）
- 商品上下架操作
- 保存修改

**信息区域**:
| 模块 | 内容 |
|------|------|
| 基本信息 | 商品名称、货号、分类、品牌、描述 |
| 价格设置 | 销售价、原价、会员价 |
| 库存管理 | 总库存、规格库存 |
| 图片管理 | 主图、详情图列表 |
| 状态设置 | 上架/下架/售罄 |

**视觉设计**:
- 表单: 白色背景，整洁的输入框
- 图片: 缩略图列表，支持点击预览
- 状态标签: 上架(绿色)、下架(灰色)、售罄(红色)

---

### 3.7 报表管理页面 (Report.vue)

**布局**:
- 顶部日期筛选区域
- 多个报表卡片

**功能**:
- 日期范围选择
- 销售额报表（按日/按月）
- 订单统计报表
- 商品销售排行
- 导出报表功能

**报表卡片**:
1. 销售趋势图
2. 订单统计
3. 商品销售排行
4. 客户统计

---

## 4. 公共布局设计

### 4.1 侧边栏 (Sidebar.vue)

- 宽度: 200px（可折叠）
- 背景: 黑色 (#1a1a1a)
- 文字: 白色 / 浅灰色
- 菜单项: 图标 + 文字
- 选中状态: 浅蓝色高亮

**菜单项**:
1. 数据总览 (Dashboard)
2. 订单管理 (Order)
3. 商品管理 (Product)
4. 报表管理 (Report)

### 4.2 顶部栏 (Header.vue)

- 高度: 60px
- 背景: 白色 (#ffffff)
- 左侧: 系统标题
- 右侧: 用户信息、退出登录

---

## 5. 颜色规范

| 用途 | 颜色代码 | 说明 |
|------|----------|------|
| 主色 | #87CEEB | 浅蓝色 |
| 主色深 | #5BC0DE | 深一点的蓝色 |
| 背景色 | #FFFFFF | 白色 |
| 背景灰 | #F5F7FA | 浅灰色背景 |
| 文字主色 | #1A1A1A | 黑色 |
| 文字次色 | #666666 | 灰色 |
| 边框色 | #E4E7ED | 浅边框 |
| 成功色 | #67C23A | 绿色 |
| 警告色 | #E6A23C | 橙色 |
| 危险色 | #F56C6C | 红色 |

---

## 6. 响应式设计

- 桌面端: > 1200px（完整布局）
- 平板端: 768px - 1200px（侧边栏折叠）
- 移动端: < 768px（侧边栏隐藏，顶部菜单）

---

## 7. 组件命名规范

- 页面组件: Login.vue, Dashboard.vue, Order.vue, OrderDetail.vue, Product.vue, ProductDetail.vue, Report.vue
- 公共组件: AppHeader.vue, AppSidebar.vue, AppFooter.vue
- 业务组件: OrderTable.vue, ProductCard.vue, StatsCard.vue

---

## 8. 路由配置

```typescript
// router/index.ts
const routes = [
  { path: '/login', name: 'Login', component: () => import('@/views/Login.vue') },
  { 
    path: '/', 
    component: () => import('@/layout/MainLayout.vue'),
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/Dashboard.vue') },
      { path: 'order', name: 'Order', component: () => import('@/views/Order.vue') },
      { path: 'order/:id', name: 'OrderDetail', component: () => import('@/views/OrderDetail.vue') },
      { path: 'product', name: 'Product', component: () => import('@/views/Product.vue') },
      { path: 'product/:id', name: 'ProductDetail', component: () => import('@/views/ProductDetail.vue') },
      { path: 'report', name: 'Report', component: () => import('@/views/Report.vue') },
    ]
  }
]
```

---

## 9. 下一步

1. 初始化 Vue 项目 (npm create vite@latest)
2. 安装依赖 (element-plus, vue-router, pinia, axios, echarts)
3. 创建项目结构
4. 实现各个页面组件
5. 对接后端 API
