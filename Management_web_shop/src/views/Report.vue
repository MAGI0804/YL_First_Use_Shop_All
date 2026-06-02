<template>
  <div class="report-page">
    <div class="filter-bar">
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        style="width: 300px;"
      />
      <el-button type="primary" style="margin-left: 12px;">查询</el-button>
      <el-button>导出报表</el-button>
    </div>

    <div class="report-cards">
      <el-card class="report-card">
        <template #header>
          <span>销售趋势</span>
        </template>
        <div class="chart-placeholder">
          <el-icon :size="48" color="#ddd"><TrendCharts /></el-icon>
          <p>销售趋势图表</p>
        </div>
      </el-card>

      <el-card class="report-card">
        <template #header>
          <span>订单统计</span>
        </template>
        <div class="stat-grid">
          <div class="stat-item">
            <div class="stat-value">128</div>
            <div class="stat-label">订单总数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">¥28,680</div>
            <div class="stat-label">销售总额</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">¥224</div>
            <div class="stat-label">平均客单价</div>
          </div>
        </div>
      </el-card>

      <el-card class="report-card">
        <template #header>
          <span>商品销售排行</span>
        </template>
        <el-table :data="productRank" size="small">
          <el-table-column type="index" label="排名" width="60" />
          <el-table-column prop="name" label="商品名称" />
          <el-table-column prop="sales" label="销量" width="80" />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { TrendCharts } from '@element-plus/icons-vue'

const dateRange = ref([])

const productRank = ref([
  { name: '儿童公主裙', sales: 45 },
  { name: '男童运动T恤', sales: 38 },
  { name: '女童羽绒服', sales: 25 },
  { name: '男童牛仔裤', sales: 20 },
  { name: '女童连衣裙', sales: 18 },
])
</script>

<style scoped>
.report-page {
  padding: 20px;
}

.filter-bar {
  display: flex;
  align-items: center;
}

.report-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-top: 20px;
}

.report-card:first-child {
  grid-column: span 2;
}

.report-card :deep(.el-card__header) {
  font-weight: 500;
}

.chart-placeholder {
  height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
}

.chart-placeholder p {
  margin-top: 12px;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a1a;
}

.stat-label {
  font-size: 14px;
  color: #999;
  margin-top: 8px;
}
</style>
