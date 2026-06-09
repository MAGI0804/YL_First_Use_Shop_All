<template>
  <div class="report-page">
    <section class="panel">
      <div class="toolbar">
        <CompactDateRangePicker v-model="dateRange" />
        <el-input v-model="filters.category" placeholder="商品分类" clearable class="field" @keyup.enter="handleSearch" />
        <el-input v-model="filters.style_code" placeholder="款号" clearable class="field" @keyup.enter="handleSearch" />
        <el-input-number v-model="filters.low_inventory_threshold" :min="0" :max="9999" controls-position="right" class="number-field" />
        <el-button type="primary" :icon="Search" :loading="loading" @click="handleSearch">查询</el-button>
        <el-button :icon="Download" @click="handleExport">导出</el-button>
      </div>
    </section>

    <section class="summary-band">
      <div class="summary-item">
        <span>销售额</span>
        <strong>¥{{ formatMoney(sales.sales_amount) }}</strong>
      </div>
      <div class="summary-item">
        <span>已支付订单</span>
        <strong>{{ sales.paid_order_count }}</strong>
      </div>
      <div class="summary-item">
        <span>优惠金额</span>
        <strong>¥{{ formatMoney(sales.discount_amount) }}</strong>
      </div>
      <div class="summary-item">
        <span>退款金额</span>
        <strong>¥{{ formatMoney(sales.refund_amount) }}</strong>
      </div>
      <div class="summary-item">
        <span>客单价</span>
        <strong>¥{{ formatMoney(sales.average_order_value) }}</strong>
      </div>
    </section>

    <section class="report-grid">
      <div class="panel wide">
        <div class="section-title">销售日明细</div>
        <el-table :data="sales.daily" border height="360" empty-text="暂无销售数据">
          <el-table-column prop="date" label="日期" width="120" />
          <el-table-column prop="order_count" label="订单数" width="90" align="right" />
          <el-table-column prop="paid_order_count" label="已支付" width="90" align="right" />
          <el-table-column prop="sales_amount" label="销售额" min-width="120" align="right">
            <template #default="{ row }">¥{{ formatMoney(row.sales_amount) }}</template>
          </el-table-column>
          <el-table-column prop="discount_amount" label="优惠" min-width="100" align="right">
            <template #default="{ row }">¥{{ formatMoney(row.discount_amount) }}</template>
          </el-table-column>
          <el-table-column prop="refund_amount" label="退款" min-width="100" align="right">
            <template #default="{ row }">¥{{ formatMoney(row.refund_amount) }}</template>
          </el-table-column>
          <el-table-column prop="average_order_value" label="客单价" min-width="110" align="right">
            <template #default="{ row }">¥{{ formatMoney(row.average_order_value) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="panel">
        <div class="section-title">用户分析</div>
        <div class="metric-list">
          <div><span>新增用户</span><strong>{{ users.new_user_count }}</strong></div>
          <div><span>新增会员</span><strong>{{ users.new_member_count }}</strong></div>
          <div><span>下单用户</span><strong>{{ users.order_user_count }}</strong></div>
          <div><span>支付用户</span><strong>{{ users.paid_user_count }}</strong></div>
          <div><span>复购用户</span><strong>{{ users.repurchase_user_count }}</strong></div>
        </div>
      </div>

      <div class="panel">
        <div class="section-title">商品评价与库存</div>
        <div class="metric-list">
          <div><span>库存周转率</span><strong>{{ formatNumber(products.inventory_turnover_rate) }}</strong></div>
          <div><span>低库存商品</span><strong>{{ products.low_inventory_count }}</strong></div>
          <div><span>平均评分</span><strong>{{ formatNumber(products.average_rating) }}</strong></div>
          <div><span>好评率</span><strong>{{ formatPercent(products.good_rate) }}</strong></div>
        </div>
      </div>

      <div class="panel">
        <div class="section-title">热销 SKU</div>
        <el-table :data="products.hot_skus" size="small" border empty-text="暂无热销 SKU">
          <el-table-column type="index" label="排名" width="60" />
          <el-table-column prop="name" label="商品" min-width="150" show-overflow-tooltip />
          <el-table-column prop="sales_qty" label="销量" width="80" align="right" />
          <el-table-column prop="sales_amount" label="销售额" width="110" align="right">
            <template #default="{ row }">¥{{ formatMoney(row.sales_amount) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="panel">
        <div class="section-title">热销款号</div>
        <el-table :data="products.hot_style_codes" size="small" border empty-text="暂无款号数据">
          <el-table-column type="index" label="排名" width="60" />
          <el-table-column prop="style_code" label="款号" />
          <el-table-column prop="sales_qty" label="销量" width="80" align="right" />
          <el-table-column prop="inventory" label="库存" width="80" align="right" />
        </el-table>
      </div>

      <div class="panel">
        <div class="section-title">分类偏好</div>
        <el-table :data="users.category_preferences" size="small" border empty-text="暂无分类偏好">
          <el-table-column prop="name" label="分类" />
          <el-table-column prop="user_count" label="用户数" width="90" align="right" />
          <el-table-column prop="sales_qty" label="购买件数" width="100" align="right" />
        </el-table>
      </div>

      <div class="panel">
        <div class="section-title">滞销商品</div>
        <el-table :data="products.slow_moving_products" size="small" border empty-text="暂无滞销商品">
          <el-table-column prop="name" label="商品" min-width="150" show-overflow-tooltip />
          <el-table-column prop="style_code" label="款号" width="110" />
          <el-table-column prop="sales_qty" label="销量" width="80" align="right" />
          <el-table-column prop="inventory" label="库存" width="80" align="right" />
        </el-table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, Search } from '@element-plus/icons-vue'
import CompactDateRangePicker from '@/components/CompactDateRangePicker.vue'
import {
  createDownloadTask,
  queryProductSummary,
  querySalesSummary,
  queryUserSummary,
  type AnalyticsFilterParams,
  type ProductSummaryData,
  type SalesSummaryData,
  type UserSummaryData
} from '@/api'

const loading = ref(false)
const dateRange = ref<[string, string] | null>(null)

const filters = reactive({
  category: '',
  style_code: '',
  low_inventory_threshold: 5,
  slow_sales_threshold: 0,
  limit: 20
})

const sales = ref<SalesSummaryData>({
  order_count: 0,
  paid_order_count: 0,
  canceled_order_count: 0,
  sales_amount: 0,
  paid_amount: 0,
  original_order_amount: 0,
  discount_amount: 0,
  refund_amount: 0,
  average_order_value: 0,
  daily: []
})

const users = ref<UserSummaryData>({
  new_user_count: 0,
  new_member_count: 0,
  order_user_count: 0,
  paid_user_count: 0,
  repurchase_user_count: 0,
  category_preferences: [],
  style_preferences: []
})

const products = ref<ProductSummaryData>({
  hot_skus: [],
  hot_style_codes: [],
  slow_moving_products: [],
  inventory_turnover_rate: 0,
  low_inventory_count: 0,
  average_rating: 0,
  good_rate: 0
})

const formatMoney = (value: number) => Number(value || 0).toFixed(2)
const formatNumber = (value: number) => Number(value || 0).toFixed(2)
const formatPercent = (value: number) => `${(Number(value || 0) * 100).toFixed(2)}%`

const queryParams = (): AnalyticsFilterParams => ({
  begin_time: dateRange.value?.[0] || '',
  end_time: dateRange.value?.[1] || '',
  category: filters.category,
  style_code: filters.style_code,
  low_inventory_threshold: filters.low_inventory_threshold,
  slow_sales_threshold: filters.slow_sales_threshold,
  limit: filters.limit
})

const loadReports = async () => {
  loading.value = true
  try {
    const params = queryParams()
    const [salesRes, usersRes, productsRes] = await Promise.all([
      querySalesSummary(params),
      queryUserSummary(params),
      queryProductSummary(params)
    ])
    sales.value = salesRes.data?.summary || sales.value
    users.value = usersRes.data?.summary || users.value
    products.value = productsRes.data?.summary || products.value
  } catch (error) {
    console.error('query reports failed:', error)
    ElMessage.error('报表查询失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadReports()
}

const handleExport = async () => {
  try {
    await createDownloadTask({
      template_code: 'analytics_sales_export',
      filters: queryParams(),
      file_format: 'xlsx'
    })
    ElMessage.success('下载任务已创建，请到下载中心查看')
  } catch (error) {
    console.error('create report download task failed:', error)
    ElMessage.error('下载任务创建失败')
  }
}

onMounted(() => {
  loadReports()
})
</script>

<style scoped>
.report-page {
  padding: 20px;
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
}

.field {
  width: 150px;
}

.number-field {
  width: 140px;
}

.summary-band {
  display: grid;
  grid-template-columns: repeat(5, minmax(140px, 1fr));
  gap: 12px;
  margin: 16px 0;
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

.report-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.wide {
  grid-column: span 2;
}

.section-title {
  color: #303133;
  font-weight: 600;
  margin-bottom: 12px;
}

.metric-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(120px, 1fr));
  gap: 12px;
}

.metric-list div {
  border: 1px solid #eef0f3;
  border-radius: 6px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.metric-list span {
  color: #606266;
  font-size: 13px;
}

.metric-list strong {
  color: #1f2937;
  font-size: 20px;
}

@media (max-width: 1100px) {
  .summary-band,
  .report-grid {
    grid-template-columns: 1fr;
  }

  .wide {
    grid-column: span 1;
  }
}
</style>
