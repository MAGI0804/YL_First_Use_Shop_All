<template>
  <div class="reviews-page">
    <section class="summary-band">
      <div class="summary-item">
        <span>评价总数</span>
        <strong>{{ statistics.total }}</strong>
      </div>
      <div class="summary-item">
        <span>平均评分</span>
        <strong>{{ formatNumber(statistics.average_rating) }}</strong>
      </div>
      <div class="summary-item">
        <span>好评率</span>
        <strong>{{ formatPercent(statistics.good_rate) }}</strong>
      </div>
      <div class="summary-item">
        <span>待审核</span>
        <strong>{{ pendingCount }}</strong>
      </div>
    </section>

    <section class="panel">
      <div class="toolbar">
        <el-input v-model="filters.order_id" placeholder="订单号" clearable class="field" @keyup.enter="handleSearch" />
        <el-input v-model="filters.commodity_id" placeholder="商品 ID" clearable class="field" @keyup.enter="handleSearch" />
        <el-input v-model="filters.style_code" placeholder="款号" clearable class="field" @keyup.enter="handleSearch" />
        <el-select v-model="filters.status" placeholder="状态" clearable class="field">
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
          <el-option label="已隐藏" value="hidden" />
        </el-select>
        <el-button type="primary" :icon="Search" :loading="loading" @click="handleSearch">查询</el-button>
        <el-button :icon="Refresh" @click="handleReset">重置</el-button>
      </div>

      <el-table :data="rows" border height="560" row-key="id" empty-text="暂无评价">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_id" label="订单号" min-width="150" />
        <el-table-column prop="commodity_id" label="商品 ID" min-width="130" />
        <el-table-column prop="style_code" label="款号" width="120" />
        <el-table-column prop="rating" label="评分" width="150">
          <template #default="{ row }">
            <el-rate :model-value="row.rating" disabled />
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="评价内容" min-width="220" show-overflow-tooltip />
        <el-table-column label="标签" min-width="160">
          <template #default="{ row }">
            <el-tag v-for="tag in parseList(row.tags)" :key="tag" size="small" class="tag-item">{{ tag }}</el-tag>
            <span v-if="parseList(row.tags).length === 0">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'pending'" link type="success" @click="audit(row, 'approved')">通过</el-button>
            <el-button v-if="row.status === 'pending'" link type="danger" @click="audit(row, 'rejected')">拒绝</el-button>
            <el-button v-if="row.status === 'approved'" link type="warning" @click="audit(row, 'hidden')">隐藏</el-button>
            <el-button link type="info" @click="reply(row)">回复</el-button>
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
          @size-change="loadReviews"
          @current-change="loadReviews"
        />
      </div>
    </section>

    <el-drawer v-model="detailVisible" title="评价详情" size="520px">
      <div v-if="selectedRow" class="detail-body">
        <div class="detail-row"><span>评价 ID</span><strong>{{ selectedRow.id }}</strong></div>
        <div class="detail-row"><span>订单号</span><strong>{{ selectedRow.order_id }}</strong></div>
        <div class="detail-row"><span>子订单号</span><strong>{{ selectedRow.sub_order_id || '-' }}</strong></div>
        <div class="detail-row"><span>商品 ID</span><strong>{{ selectedRow.commodity_id }}</strong></div>
        <div class="detail-row"><span>款号</span><strong>{{ selectedRow.style_code || '-' }}</strong></div>
        <div class="detail-row"><span>评分</span><el-rate :model-value="selectedRow.rating" disabled /></div>
        <div class="detail-row"><span>状态</span><strong>{{ statusText(selectedRow.status) }}</strong></div>
        <div class="detail-row"><span>审核备注</span><strong>{{ selectedRow.audit_remark || '-' }}</strong></div>
        <div class="content-block">
          <div class="section-title">评价内容</div>
          <p>{{ selectedRow.content || '-' }}</p>
        </div>
        <div class="content-block">
          <div class="section-title">图片</div>
          <div class="image-list" v-if="parseList(selectedRow.images).length > 0">
            <img v-for="image in parseList(selectedRow.images)" :key="image" :src="image" />
          </div>
          <span v-else>-</span>
        </div>
        <div class="content-block">
          <div class="section-title">商家回复</div>
          <div v-if="selectedRow.replies && selectedRow.replies.length > 0" class="reply-list">
            <div v-for="item in selectedRow.replies" :key="item.id" class="reply-item">
              <strong>{{ item.operator_id }}</strong>
              <span>{{ item.created_at }}</span>
              <p>{{ item.content }}</p>
            </div>
          </div>
          <span v-else>-</span>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import {
  auditReview,
  queryBackendReviews,
  queryReviewStatistics,
  replyReview,
  type ReviewItem,
  type ReviewStatisticsData
} from '@/api'

const loading = ref(false)
const rows = ref<ReviewItem[]>([])
const total = ref(0)
const detailVisible = ref(false)
const selectedRow = ref<ReviewItem | null>(null)

const filters = reactive({
  order_id: '',
  commodity_id: '',
  style_code: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  page_size: 10
})

const statistics = ref<ReviewStatisticsData>({
  total: 0,
  average_rating: 0,
  good_rate: 0,
  rating_distribution: {}
})

const pendingCount = computed(() => rows.value.filter(row => row.status === 'pending').length)

const formatNumber = (value: number) => Number(value || 0).toFixed(2)
const formatPercent = (value: number) => `${(Number(value || 0) * 100).toFixed(2)}%`

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝',
    hidden: '已隐藏'
  }
  return map[status] || status || '-'
}

const statusTag = (status: string) => {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info' | ''> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger',
    hidden: 'info'
  }
  return map[status] || ''
}

const parseList = (value: string) => {
  if (!value) return []
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) return parsed.filter(Boolean).map(item => String(item))
  } catch {
    // legacy values may be comma-separated.
  }
  return value.split(',').map(item => item.trim()).filter(Boolean)
}

const loadReviews = async () => {
  loading.value = true
  try {
    const res = await queryBackendReviews({
      order_id: filters.order_id,
      commodity_id: filters.commodity_id,
      style_code: filters.style_code,
      status: filters.status,
      page: pagination.page,
      page_size: pagination.page_size
    })
    rows.value = res.data?.data || []
    total.value = res.data?.total || 0
  } catch (error) {
    console.error('query reviews failed:', error)
    ElMessage.error('评价列表查询失败')
  } finally {
    loading.value = false
  }
}

const loadStatistics = async () => {
  try {
    const res = await queryReviewStatistics({
      commodity_id: filters.commodity_id,
      style_code: filters.style_code
    })
    statistics.value = res.data?.statistics || statistics.value
  } catch (error) {
    console.error('query review statistics failed:', error)
    ElMessage.error('评价统计查询失败')
  }
}

const reload = async () => {
  await Promise.all([loadReviews(), loadStatistics()])
}

const handleSearch = () => {
  pagination.page = 1
  reload()
}

const handleReset = () => {
  filters.order_id = ''
  filters.commodity_id = ''
  filters.style_code = ''
  filters.status = ''
  pagination.page = 1
  pagination.page_size = 10
  reload()
}

const audit = async (row: ReviewItem, status: 'approved' | 'rejected' | 'hidden') => {
  let remark = ''
  if (status === 'rejected' || status === 'hidden') {
    const result = await ElMessageBox.prompt(status === 'rejected' ? '请输入拒绝原因' : '请输入隐藏原因', statusText(status), {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      inputValidator: value => Boolean(value && value.trim()),
      inputErrorMessage: '原因不能为空'
    })
    remark = result.value
  } else {
    await ElMessageBox.confirm('确认通过该评价？通过后会展示在商品详情。', '审核评价', {
      confirmButtonText: '通过',
      cancelButtonText: '取消',
      type: 'warning'
    })
  }

  try {
    await auditReview({
      review_id: row.id,
      status,
      audit_remark: remark
    })
    ElMessage.success('评价状态已更新')
    await reload()
  } catch (error) {
    console.error('audit review failed:', error)
    ElMessage.error('评价审核失败')
  }
}

const reply = async (row: ReviewItem) => {
  const result = await ElMessageBox.prompt('请输入商家回复内容', '回复评价', {
    confirmButtonText: '发送回复',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputValidator: value => Boolean(value && value.trim()),
    inputErrorMessage: '回复内容不能为空'
  })

  try {
    await replyReview({
      review_id: row.id,
      operator_id: 'admin',
      content: result.value
    })
    ElMessage.success('已回复评价')
    await reload()
  } catch (error) {
    console.error('reply review failed:', error)
    ElMessage.error('回复评价失败')
  }
}

const openDetail = (row: ReviewItem) => {
  selectedRow.value = row
  detailVisible.value = true
}

onMounted(() => {
  reload()
})
</script>

<style scoped>
.reviews-page {
  padding: 20px;
}

.summary-band {
  display: grid;
  grid-template-columns: repeat(4, minmax(140px, 1fr));
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

.summary-item span {
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

.tag-item {
  margin-right: 6px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
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

.section-title {
  color: #303133;
  font-weight: 600;
  margin-bottom: 8px;
}

.content-block p {
  margin: 0;
  color: #303133;
  line-height: 1.6;
}

.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.image-list img {
  width: 88px;
  height: 88px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
}

.reply-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.reply-item {
  padding: 10px;
  background: #f6f8fa;
  border-radius: 6px;
}

.reply-item span {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

.reply-item p {
  margin: 6px 0 0;
}

@media (max-width: 1100px) {
  .summary-band {
    grid-template-columns: repeat(2, minmax(140px, 1fr));
  }
}
</style>
