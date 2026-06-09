<template>
  <div class="after-sales-page">
    <section class="summary-band">
      <div class="summary-item">
        <span class="summary-label">申请数</span>
        <strong>{{ statistics.total_count }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">待审核</span>
        <strong>{{ statistics.pending_count }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">已完成</span>
        <strong>{{ statistics.completed_count }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">售后率</span>
        <strong>{{ formatPercent(statistics.after_sale_rate) }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">售后金额</span>
        <strong>¥{{ formatMoney(statistics.after_sale_amount) }}</strong>
      </div>
    </section>

    <section class="panel">
      <div class="toolbar">
        <el-input v-model="filters.order_id" placeholder="订单号" clearable class="field" @keyup.enter="handleSearch" />
        <el-input v-model="filters.return_order_id" placeholder="售后单号" clearable class="field" @keyup.enter="handleSearch" />
        <el-select v-model="filters.type" placeholder="售后类型" clearable class="field">
          <el-option label="退货" value="return" />
          <el-option label="换货" value="exchange" />
          <el-option label="仅退款" value="refund" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable class="field">
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
          <el-option label="买家已寄回" value="buyer_shipped" />
          <el-option label="已完成" value="completed" />
          <el-option label="已取消" value="canceled" />
        </el-select>
        <CompactDateRangePicker v-model="dateRange" />
        <el-button type="primary" :icon="Search" :loading="loading" @click="handleSearch">查询</el-button>
        <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        <el-button :icon="Download" @click="handleExportTask">导出</el-button>
      </div>

      <el-table :data="displayRows" border height="560" row-key="return_id" empty-text="暂无售后记录">
        <el-table-column prop="return_id" label="售后单号" min-width="170" />
        <el-table-column prop="order_id" label="订单号" min-width="150" />
        <el-table-column prop="type" label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.type)" size="small">{{ typeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="160" show-overflow-tooltip />
        <el-table-column prop="specific_reasons" label="说明" min-width="200" show-overflow-tooltip />
        <el-table-column prop="express_company" label="物流公司" width="120" />
        <el-table-column prop="express_number" label="物流单号" width="150" />
        <el-table-column prop="request_time" label="申请时间" width="180" />
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'pending'" link type="success" @click="approve(row)">通过</el-button>
            <el-button v-if="row.status === 'pending'" link type="danger" @click="reject(row)">拒绝</el-button>
            <el-button v-if="canReceive(row)" link type="warning" @click="receive(row)">确认收货</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadReturnOrders"
          @current-change="loadReturnOrders"
        />
      </div>
    </section>

    <section class="panel reason-panel">
      <div class="section-title">售后原因排行</div>
      <el-table :data="statistics.reason_rank" size="small" border empty-text="暂无原因统计">
        <el-table-column type="index" label="排名" width="70" />
        <el-table-column prop="reason" label="原因" />
        <el-table-column prop="count" label="数量" width="100" align="right" />
      </el-table>
    </section>

    <el-drawer v-model="detailVisible" title="售后详情" size="520px">
      <div v-if="selectedRow" class="detail-body">
        <div class="detail-row"><span>售后单号</span><strong>{{ selectedRow.return_id }}</strong></div>
        <div class="detail-row"><span>订单号</span><strong>{{ selectedRow.order_id }}</strong></div>
        <div class="detail-row"><span>类型</span><strong>{{ typeText(selectedRow.type) }}</strong></div>
        <div class="detail-row"><span>状态</span><strong>{{ statusText(selectedRow.status) }}</strong></div>
        <div class="detail-row"><span>原因</span><strong>{{ selectedRow.reason || '-' }}</strong></div>
        <div class="detail-row"><span>说明</span><strong>{{ selectedRow.specific_reasons || '-' }}</strong></div>
        <div class="detail-row"><span>买家电话</span><strong>{{ selectedRow.buyer_phone || '-' }}</strong></div>
        <div class="detail-row"><span>买家地址</span><strong>{{ buyerAddress(selectedRow) }}</strong></div>
        <div class="detail-row"><span>寄回物流</span><strong>{{ returnLogistics(selectedRow) }}</strong></div>
        <div class="detail-row"><span>申请时间</span><strong>{{ selectedRow.request_time || '-' }}</strong></div>
        <div class="detail-row"><span>完成时间</span><strong>{{ selectedRow.completed_time || '-' }}</strong></div>
        <div class="product-block">
          <div class="section-title">售后商品</div>
          <pre>{{ formatProductList(selectedRow.product_list) }}</pre>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Refresh, Search } from '@element-plus/icons-vue'
import CompactDateRangePicker from '@/components/CompactDateRangePicker.vue'
import {
  approveReturnOrder,
  createDownloadTask,
  queryReturnOrders,
  queryReturnOrderStatistics,
  receiveReturnOrder,
  type ReturnOrderItem,
  type ReturnOrderStatisticsData
} from '@/api'

const loading = ref(false)
const dateRange = ref<[string, string] | null>(null)
const rows = ref<ReturnOrderItem[]>([])
const total = ref(0)
const detailVisible = ref(false)
const selectedRow = ref<ReturnOrderItem | null>(null)

const emptyStatistics: ReturnOrderStatisticsData = {
  total_count: 0,
  pending_count: 0,
  completed_count: 0,
  after_sale_rate: 0,
  after_sale_amount: 0,
  reason_rank: [],
  completed_orders: 0,
  after_sale_orders: 0
}
const statistics = ref<ReturnOrderStatisticsData>({ ...emptyStatistics })

const filters = reactive({
  order_id: '',
  return_order_id: '',
  type: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  page_size: 10
})

const displayRows = computed(() => {
  if (!filters.type) return rows.value
  return rows.value.filter(row => row.type === filters.type)
})

const typeText = (type: string) => {
  const map: Record<string, string> = {
    return: '退货',
    exchange: '换货',
    refund: '仅退款',
    return_refund: '退货'
  }
  return map[type] || type || '-'
}

const typeTag = (type: string) => {
  const map: Record<string, 'danger' | 'warning' | 'info' | ''> = {
    return: 'danger',
    exchange: 'warning',
    refund: 'info',
    return_refund: 'danger'
  }
  return map[type] || ''
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝',
    buyer_shipped: '买家已寄回',
    shipped: '买家已寄回',
    completed: '已完成',
    returned: '已完成',
    canceled: '已取消'
  }
  return map[status] || status || '-'
}

const statusTag = (status: string) => {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'primary' | ''> = {
    pending: 'warning',
    approved: 'primary',
    rejected: 'danger',
    buyer_shipped: 'info',
    shipped: 'info',
    completed: 'success',
    returned: 'success',
    canceled: 'info'
  }
  return map[status] || ''
}

const formatMoney = (value: number) => Number(value || 0).toFixed(2)
const formatPercent = (value: number) => `${(Number(value || 0) * 100).toFixed(2)}%`

const queryDateParams = () => ({
  begin_time: dateRange.value?.[0] || '',
  end_time: dateRange.value?.[1] || ''
})

const loadReturnOrders = async () => {
  loading.value = true
  try {
    const res = await queryReturnOrders({
      order_id: filters.order_id,
      return_order_id: filters.return_order_id,
      status: filters.status,
      page: pagination.page,
      page_size: pagination.page_size
    })
    rows.value = res.data?.return_orders || []
    total.value = res.data?.total || 0
  } catch (error) {
    console.error('query return orders failed:', error)
    ElMessage.error('售后列表查询失败')
  } finally {
    loading.value = false
  }
}

const loadStatistics = async () => {
  try {
    const res = await queryReturnOrderStatistics(queryDateParams())
    statistics.value = res.data?.statistics || { ...emptyStatistics }
  } catch (error) {
    console.error('query return statistics failed:', error)
    ElMessage.error('售后统计查询失败')
  }
}

const reload = async () => {
  await Promise.all([loadReturnOrders(), loadStatistics()])
}

const handleSearch = () => {
  pagination.page = 1
  reload()
}

const handleReset = () => {
  filters.order_id = ''
  filters.return_order_id = ''
  filters.type = ''
  filters.status = ''
  dateRange.value = null
  pagination.page = 1
  pagination.page_size = 10
  reload()
}

const handleExportTask = async () => {
  try {
    await createDownloadTask({
      template_code: 'after_sale_export',
      file_format: 'xlsx',
      filters: {
        begin_time: dateRange.value?.[0] || undefined,
        end_time: dateRange.value?.[1] || undefined,
        status: filters.status || undefined,
        type: filters.type || undefined,
        order_id: filters.order_id || undefined,
        return_id: filters.return_order_id || undefined
      }
    })
    ElMessage.success('售后下载任务已创建，请到下载中心查看')
  } catch (error) {
    console.error('create after-sale download task failed:', error)
    ElMessage.error('售后下载任务创建失败')
  }
}

const approve = async (row: ReturnOrderItem) => {
  await submitApproval(row, 'approved')
}

const reject = async (row: ReturnOrderItem) => {
  const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝售后', {
    confirmButtonText: '确认拒绝',
    cancelButtonText: '取消',
    inputValidator: value => Boolean(value && value.trim()),
    inputErrorMessage: '拒绝原因不能为空'
  })
  await submitApproval(row, 'rejected', value)
}

const submitApproval = async (row: ReturnOrderItem, status: 'approved' | 'rejected', remark = '') => {
  try {
    await approveReturnOrder({
      return_order_id: row.return_id,
      approve_status: status,
      user_id: row.user_id || 1,
      remark
    })
    ElMessage.success(status === 'approved' ? '售后已通过' : '售后已拒绝')
    await reload()
  } catch (error) {
    console.error('approve return order failed:', error)
    ElMessage.error('售后审核失败')
  }
}

const canReceive = (row: ReturnOrderItem) => {
  return row.status === 'buyer_shipped' || row.status === 'shipped' || (row.type === 'refund' && row.status === 'approved')
}

const receive = async (row: ReturnOrderItem) => {
  await ElMessageBox.confirm('确认已收货并完成售后？完成后会触发库存回滚。', '确认收货', {
    confirmButtonText: '确认完成',
    cancelButtonText: '取消',
    type: 'warning'
  })
  try {
    await receiveReturnOrder({
      return_order_id: row.return_id,
      user_id: row.user_id || 1
    })
    ElMessage.success('售后已完成')
    await reload()
  } catch (error) {
    console.error('receive return order failed:', error)
    ElMessage.error('确认收货失败')
  }
}

const openDetail = (row: ReturnOrderItem) => {
  selectedRow.value = row
  detailVisible.value = true
}

const buyerAddress = (row: ReturnOrderItem) => {
  const address = `${row.buyer_province || ''}${row.buyer_city || ''}${row.buyer_county || ''}${row.buyer_address || ''}`
  return address || '-'
}

const returnLogistics = (row: ReturnOrderItem) => {
  if (!row.express_company && !row.express_number) return '-'
  return `${row.express_company || ''} ${row.express_number || ''}`.trim()
}

const formatProductList = (productList: ReturnOrderItem['product_list']) => {
  if (!productList) return '-'
  if (Array.isArray(productList)) return JSON.stringify(productList, null, 2)
  try {
    return JSON.stringify(JSON.parse(productList), null, 2)
  } catch {
    return String(productList)
  }
}

onMounted(() => {
  reload()
})
</script>

<style scoped>
.after-sales-page {
  padding: 20px;
}

.summary-band {
  display: grid;
  grid-template-columns: repeat(5, minmax(140px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.summary-item {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.summary-label {
  color: #606266;
  font-size: 13px;
}

.summary-item strong {
  color: #1f2937;
  font-size: 22px;
}

.panel {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.field {
  width: 160px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

.reason-panel {
  margin-top: 16px;
}

.section-title {
  color: #303133;
  font-weight: 600;
  margin-bottom: 12px;
}

.detail-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: grid;
  grid-template-columns: 92px 1fr;
  gap: 12px;
  line-height: 1.5;
}

.detail-row span {
  color: #909399;
}

.detail-row strong {
  color: #303133;
  font-weight: 500;
  word-break: break-word;
}

.product-block {
  margin-top: 8px;
}

.product-block pre {
  margin: 0;
  padding: 12px;
  background: #f6f8fa;
  border-radius: 6px;
  color: #303133;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 1100px) {
  .summary-band {
    grid-template-columns: repeat(2, minmax(140px, 1fr));
  }
}
</style>
