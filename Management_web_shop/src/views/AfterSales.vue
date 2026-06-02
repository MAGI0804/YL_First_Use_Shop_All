<template>
  <div class="after-sales-page">
    <div class="top-bar">
      <div class="search-bar">
        <el-input
          v-model="searchOrderNo"
          placeholder="请输入订单号"
          style="width: 140px;"
          clearable
        />
        <span class="time-label">售后时间</span>
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          style="width: 320px; margin-left: 8px;"
          value-format="YYYY-MM-DD HH:mm:ss"
          :default-time="[new Date(2000, 1, 1, 0, 0, 0), new Date(2000, 1, 1, 23, 59, 59)]"
        />
        <el-select v-model="typeFilter" placeholder="售后类型" style="width: 100px; margin-left: 8px;">
          <el-option label="全部" value="" />
          <el-option label="退货" value="return" />
          <el-option label="换货" value="exchange" />
          <el-option label="退款" value="refund" />
        </el-select>
        <el-select v-model="statusFilter" placeholder="状态" style="width: 100px; margin-left: 8px;">
          <el-option label="全部" value="" />
          <el-option label="待处理" value="pending" />
          <el-option label="处理中" value="processing" />
          <el-option label="已完成" value="completed" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" style="margin-left: 8px;" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <el-table :data="filteredList" style="width: 100%; margin-top: 20px;" row-key="id">
      <el-table-column prop="orderNo" label="订单号" width="180" />
      <el-table-column label="商品信息" min-width="200">
        <template #default="{ row }">
          <div class="product-item">
            <div class="product-image"></div>
            <div class="product-detail">
              <div class="product-name">{{ row.productName }}</div>
              <div class="product-count">x{{ row.quantity }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="售后类型" width="100">
        <template #default="{ row }">
          <el-tag :type="getTypeTag(row.type)" size="small">
            {{ getTypeText(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="reason" label="售后原因" min-width="150" />
      <el-table-column prop="amount" label="退款金额" width="100">
        <template #default="{ row }">
          ¥{{ row.amount }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusTag(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="applyTime" label="申请时间" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="viewDetail(row.id)">查看</el-button>
          <el-button v-if="row.status === 'pending'" type="success" link @click="handleProcess(row)">处理</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="10"
        :total="filteredList.length"
        layout="total, prev, pager, next"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const searchOrderNo = ref('')
const dateRange = ref<[string, string] | null>(null)
const typeFilter = ref('')
const statusFilter = ref('')
const currentPage = ref(1)

const allData = ref([
  { id: 1, orderNo: 'ORDER20240427001', productName: '儿童连衣裙', quantity: 1, type: 'return', reason: '质量问题', amount: '299.00', status: 'pending', applyTime: '2024-04-27 15:30:00' },
  { id: 2, orderNo: 'ORDER20240426002', productName: '男童T恤', quantity: 1, type: 'refund', reason: '不想要了', amount: '89.00', status: 'processing', applyTime: '2024-04-26 10:20:00' },
  { id: 3, orderNo: 'ORDER40425003', productName: '女童外套', quantity: 1, type: 'exchange', reason: '尺码不合适', amount: '0.00', status: 'completed', applyTime: '2024-04-25 14:15:00' },
  { id: 4, orderNo: 'ORDER20240424004', productName: '男童裤子', quantity: 2, type: 'return', reason: '发货错误', amount: '318.00', status: 'rejected', applyTime: '2024-04-24 09:20:00' },
])

const filteredList = computed(() => {
  let list = [...allData.value]
  
  if (searchOrderNo.value) {
    list = list.filter(item => item.orderNo.includes(searchOrderNo.value))
  }
  
  if (typeFilter.value) {
    list = list.filter(item => item.type === typeFilter.value)
  }
  
  if (statusFilter.value) {
    list = list.filter(item => item.status === statusFilter.value)
  }
  
  return list
})

const getTypeTag = (type: string) => {
  const map: Record<string, string> = {
    return: 'danger',
    exchange: 'warning',
    refund: ''
  }
  return map[type] || ''
}

const getTypeText = (type: string) => {
  const map: Record<string, string> = {
    return: '退货',
    exchange: '换货',
    refund: '退款'
  }
  return map[type] || type
}

const getStatusTag = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    processing: 'primary',
    completed: 'success',
    rejected: 'info'
  }
  return map[status] || ''
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待处理',
    processing: '处理中',
    completed: '已完成',
    rejected: '已拒绝'
  }
  return map[status] || status
}

const handleSearch = () => {
  ElMessage.success(`找到 ${filteredList.value.length} 条售后记录`)
}

const handleReset = () => {
  searchOrderNo.value = ''
  dateRange.value = null
  typeFilter.value = ''
  statusFilter.value = ''
  ElMessage.success('已重置')
}

const viewDetail = (id: number) => {
  ElMessage.info('查看售后详情')
}

const handleProcess = (row: any) => {
  ElMessageBox.confirm('确认处理该售后申请？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    row.status = 'processing'
    ElMessage.success('已开始处理')
  }).catch(() => {})
}
</script>

<script lang="ts">
import { ElMessageBox } from 'element-plus'
export default {}
</script>

<style scoped>
.after-sales-page {
  padding: 20px;
}

.top-bar {
  margin-bottom: 0;
}

.search-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.time-label {
  margin-left: 8px;
  font-size: 14px;
  color: #666;
  white-space: nowrap;
}

.product-item {
  display: flex;
  align-items: center;
}

.product-image {
  width: 40px;
  height: 40px;
  background: #f5f5f5;
  margin-right: 8px;
  flex-shrink: 0;
}

.product-detail {
  display: flex;
  flex-direction: column;
}

.product-name {
  color: #1a1a1a;
}

.product-count {
  font-size: 12px;
  color: #999;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
