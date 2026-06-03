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
          <el-menu-item index="/dashboard">
            <el-icon><DataAnalysis /></el-icon>
            <span>数据总览</span>
          </el-menu-item>
          <el-menu-item index="/home-manage">
            <el-icon><Document /></el-icon>
            <span>主页管理</span>
          </el-menu-item>
          <el-menu-item index="/product">
            <el-icon><Goods /></el-icon>
            <span>商品管理</span>
          </el-menu-item>
          <el-menu-item index="/inventory">
            <el-icon><Box /></el-icon>
            <span>库存管理</span>
          </el-menu-item>
          <el-menu-item index="/order">
            <el-icon><Document /></el-icon>
            <span>订单管理</span>
          </el-menu-item>
          <el-menu-item index="/after-sales">
            <el-icon><Service /></el-icon>
            <span>售后中心</span>
          </el-menu-item>
          <el-menu-item index="/reviews">
            <el-icon><ChatDotRound /></el-icon>
            <span>评价管理</span>
          </el-menu-item>
          <el-menu-item index="/member">
            <el-icon><UserFilled /></el-icon>
            <span>会员管理</span>
          </el-menu-item>
          <el-menu-item index="/report">
            <el-icon><PieChart /></el-icon>
            <span>报表管理</span>
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
                <span>管理员</span>
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
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Box, ChatDotRound, DataAnalysis, Document, Goods, PieChart, User, UserFilled, Service } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()

const activeMenu = computed(() => route.path)

const pageTitle = computed(() => {
  return (route.meta.title as string) || '数据总览'
})

const handleCommand = (command: string) => {
  if (command === 'logout') {
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
