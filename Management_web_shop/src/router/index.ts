import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('@/views/ForgotPassword.vue'),
    meta: { title: '忘记密码' }
  },
  {
    path: '/',
    component: () => import('@/layout/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '数据总览' }
      },
      {
        path: 'home-manage',
        name: 'HomeManage',
        component: () => import('@/views/HomeManage.vue'),
        meta: { title: '主页管理' }
      },
      {
        path: 'activity/:id',
        name: 'ActivityDetail',
        component: () => import('@/views/ActivityDetail.vue'),
        meta: { title: '活动图详情' }
      },
      {
        path: 'order',
        name: 'Order',
        component: () => import('@/views/Order.vue'),
        meta: { title: '订单管理' }
      },
      {
        path: 'order/:id',
        name: 'OrderDetail',
        component: () => import('@/views/OrderDetail.vue'),
        meta: { title: '订单详情' }
      },
      {
        path: 'product',
        name: 'Product',
        component: () => import('@/views/Product.vue'),
        meta: { title: '商品管理' }
      },
      {
        path: 'product/:styleCode',
        name: 'ProductDetail',
        component: () => import('@/views/ProductDetail.vue'),
        meta: { title: '商品详情' }
      },
      {
        path: 'inventory',
        name: 'Inventory',
        component: () => import('@/views/Inventory.vue'),
        meta: { title: '库存管理' }
      },
      {
        path: 'report',
        name: 'Report',
        component: () => import('@/views/Report.vue'),
        meta: { title: '报表管理' }
      },
      {
        path: 'member',
        name: 'Member',
        component: () => import('@/views/Member.vue'),
        meta: { title: '会员管理' }
      },
      {
        path: 'member/:id',
        name: 'MemberDetail',
        component: () => import('@/views/MemberDetail.vue'),
        meta: { title: '会员详情' }
      },
      {
        path: 'after-sales',
        name: 'AfterSales',
        component: () => import('@/views/AfterSales.vue'),
        meta: { title: '售后中心' }
      },
      {
        path: 'reviews',
        name: 'Reviews',
        component: () => import('@/views/Reviews.vue'),
        meta: { title: '评价管理' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  if (to.meta.title) {
    document.title = to.meta.title as string
  }
  next()
})

export default router
