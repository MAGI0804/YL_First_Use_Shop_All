<template>
  <div class="dashboard">
    <div class="filter-bar">
      <CompactDateRangePicker v-model="dateRange" />
      <el-button type="primary" @click="handleSearch">查询</el-button>
    </div>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: #e3f2fd;">
          <el-icon :size="24" color="#87CEEB"><Document /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.todayOrders }}</div>
          <div class="stat-label">今日订单</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background: #e8f5e9;">
          <el-icon :size="24" color="#67C23A"><Money /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">¥{{ stats.todaySales }}</div>
          <div class="stat-label">今日销售额</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background: #fff3e0;">
          <el-icon :size="24" color="#E6A23C"><Box /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.pendingOrders }}</div>
          <div class="stat-label">待处理订单</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background: #fce4ec;">
          <el-icon :size="24" color="#F56C6C"><Goods /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalProducts }}</div>
          <div class="stat-label">商品总数</div>
        </div>
      </div>
    </div>

    <div class="chart-section">
      <el-card class="chart-card">
        <template #header>
          <span class="chart-title">销售趋势</span>
        </template>
        <div class="chart-placeholder">
          <el-icon :size="48" color="#ddd"><TrendCharts /></el-icon>
          <p>销售趋势图表</p>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Money, Box, Goods, TrendCharts } from '@element-plus/icons-vue'
import CompactDateRangePicker from '@/components/CompactDateRangePicker.vue'

const dateRange = ref<[string, string] | null>(null)

const stats = reactive({
  todayOrders: 28,
  todaySales: '3,680',
  pendingOrders: 5,
  totalProducts: 156
})

const handleSearch = () => {
  if (!dateRange.value) {
    ElMessage.warning('请选择日期范围')
    return
  }
  const [start, end] = dateRange.value
  ElMessage.success(`查询时间范围: ${start} 至 ${end}`)
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  background: #ffffff;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid #e4e7ed;
}

.stat-icon {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a1a;
}

.stat-label {
  font-size: 14px;
  color: #999;
  margin-top: 4px;
}

.chart-section {
  margin-top: 20px;
}

.chart-card {
  border: 1px solid #e4e7ed;
}

.chart-title {
  font-size: 16px;
  font-weight: 500;
  color: #1a1a1a;
}

.chart-placeholder {
  height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
}

.chart-placeholder p {
  margin-top: 12px;
}
</style>
