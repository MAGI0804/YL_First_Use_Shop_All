<template>
  <div class="main-layout">
    <el-container>
      <el-aside width="200px">
        <div class="logo">
          <img src="/title.jpg" alt="logo" class="logo-img" />
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          background-color="#1a1a1a"
          text-color="#ffffff"
          active-text-color="#87ceeb"
        >
          <el-menu-item v-for="item in visibleMenus" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-container>
        <el-header>
          <div class="header-left">
            <span class="page-title">{{ pageTitle }}</span>
          </div>
          <div class="header-right">
            <el-dropdown @command="handleCommand">
              <span class="user-info">
                <el-icon><User /></el-icon>
                <span>{{ backendUser?.nickname || '管理员' }}</span>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                  <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Box, ChatDotRound, DataAnalysis, Document, Download, Goods, PieChart, Service, User, UserFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { clearBackendSession, getStoredBackendUser } from '@/api'

const route = useRoute()
const router = useRouter()
const backendUser = ref(getStoredBackendUser())

const menus = [
  { path: '/dashboard', label: '数据总览', permission: 'dashboard', icon: DataAnalysis },
  { path: '/home-manage', label: '主页管理', permission: 'home-manage', icon: Document },
  { path: '/product', label: '商品管理', permission: 'product', icon: Goods },
  { path: '/inventory', label: '库存管理', permission: 'inventory', icon: Box },
  { path: '/order', label: '订单管理', permission: 'order', icon: Document },
  { path: '/after-sales', label: '售后中心', permission: 'after-sales', icon: Service },
  { path: '/reviews', label: '评价管理', permission: 'reviews', icon: ChatDotRound },
  { path: '/member', label: '会员管理', permission: 'member', icon: UserFilled },
  { path: '/report', label: '报表管理', permission: 'report', icon: PieChart },
  { path: '/download-center', label: '下载中心', permission: 'download-center', icon: Download },
  { path: '/users', label: '账号管理', permission: 'users', icon: UserFilled }
]

const activeMenu = computed(() => route.path)
const pageTitle = computed(() => (route.meta.title as string) || '数据总览')
const visibleMenus = computed(() => {
  const permissions = backendUser.value?.permissions || []
  return menus.filter((item) => permissions.includes(item.permission))
})

const handleCommand = (command: string) => {
  if (command === 'logout') {
    clearBackendSession()
    ElMessage.success('已退出登录')
    router.push('/login')
  } else if (command === 'profile') {
    ElMessage.info('个人中心功能开发中')
  }
}
</script>

<style scoped>
.main-layout {
  height: 100vh;
}

.el-container {
  height: 100%;
}

.el-aside {
  background-color: #1a1a1a;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #1a1a1a;
  border-bottom: 1px solid #333333;
}

.logo-img {
  height: 40px;
  width: auto;
  object-fit: contain;
}

.el-menu {
  border-right: none;
}

.el-menu-item {
  height: 50px;
  line-height: 50px;
}

.el-menu-item:hover {
  background-color: #2a2a2a !important;
}

.el-menu-item.is-active {
  background-color: #2a2a2a !important;
}

.el-menu-item span {
  margin-left: 8px;
}

.el-header {
  background-color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
}

.page-title {
  font-size: 16px;
  font-weight: 500;
  color: #1a1a1a;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  color: #666666;
  font-size: 14px;
}

.user-info:hover {
  color: #87ceeb;
}

.el-main {
  background-color: #f5f7fa;
  padding: 20px;
}

:deep(.el-dropdown-menu__item) {
  padding: 8px 20px;
}
</style>
