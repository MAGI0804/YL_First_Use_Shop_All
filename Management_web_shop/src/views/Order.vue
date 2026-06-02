<template>
  <div class="order-page">
    <div class="search-bar">
      <el-input
        v-model="searchOrderNo"
        placeholder="请输入订单号"
        prefix-icon="Search"
        style="width: 250px;"
        clearable
        @keyup.enter="handleSearch"
      />
      <span class="time-label">下单时间</span>
      <el-date-picker
        v-model="dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        style="width: 320px; margin-left: 8px;"
        value-format="YYYY-MM-DD HH:mm:ss"
        :default-time="[new Date(2000, 1, 1, 0, 0, 0), new Date(2000, 1, 1, 23, 59, 59)]"
        :shortcuts="shortcuts"
      />
      <div class="shortcuts">
        <el-button link @click="setToday">今</el-button>
        <el-button link @click="setYesterday">昨</el-button>
        <el-button link @click="setLast7Days">近7天</el-button>
        <el-button link @click="setLast30Days">近30天</el-button>
      </div>
      <el-select v-model="statusFilter" placeholder="状态" style="width: 120px; margin-left: 8px;">
        <el-option label="全部" value="" />
        <el-option label="未发货" value="pending" />
        <el-option label="已发货" value="shipped" />
        <el-option label="已送达" value="delivered" />
        <el-option label="已取消" value="canceled" />
        <el-option label="售后中" value="processing" />
      </el-select>
      <el-button type="primary" style="margin-left: 8px;" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <el-table :data="orderList" style="width: 100%; margin-top: 20px;" row-key="id">
      <el-table-column prop="orderNo" label="订单号" width="180" />
      <el-table-column label="商品信息" min-width="200">
        <template #default="{ row }">
          <div class="product-list">
            <div v-for="productId in row.productList" :key="productId" class="product-item">
              <div v-if="productMap[productId]?.image" class="product-image-wrapper">
                <img :src="productMap[productId].image" class="product-image" @click="handleImagePreview(productMap[productId].image)" />
                <div class="product-image-preview">
                  <img :src="productMap[productId].image" />
                </div>
              </div>
              <div v-else class="product-image"></div>
              <div class="product-detail">
                <el-tooltip 
                  v-if="productMap[productId]" 
                  placement="right" 
                  :show-after="300" 
                  :enterable="true"
                  popper-class="white-tooltip"
                >
                  <template #content>
                    <div class="product-tooltip">
                      <img v-if="productMap[productId].image" :src="productMap[productId].image" class="tooltip-image" />
                      <div class="tooltip-title">{{ productMap[productId].name }}</div>
                      <div class="tooltip-info">价格：¥{{ productMap[productId].price }}</div>
                      <div class="tooltip-info">尺码：{{ productMap[productId].size }}</div>
                      <div class="tooltip-info">类目：{{ productMap[productId].category }}</div>
                    </div>
                  </template>
                  <div class="product-name">{{ productMap[productId].name || productId }}</div>
                </el-tooltip>
                <div v-else class="product-name">{{ productId }}</div>
                <div v-if="productMap[productId]" class="product-info">
                  {{ productMap[productId].color }} / {{ productMap[productId].size }}
                </div>
              </div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="收货信息" min-width="180">
        <template #default="{ row }">
          <div class="receiver-info">
            <div>{{ row.receiver }} {{ row.phone }}</div>
            <div class="receiver-address">{{ row.address }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="buyer" label="买家" width="100" />
      <el-table-column prop="amount" label="订单金额" width="120">
        <template #default="{ row }">
          ¥{{ row.amount }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="订单状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="下单时间" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="viewDetail(row.orderNo)">查看</el-button>
          <el-button v-if="row.status === 'pending'" type="success" link @click="shipOrder(row.id)">发货</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="10"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>
  </div>
  
  <el-dialog v-model="showPreview" title="商品图片" width="60%" destroy-on-close>
    <div class="preview-image-container">
      <img :src="previewImageUrl" class="preview-image" />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { queryOrders, getToken, batchGetProducts } from '@/api'

const router = useRouter()
const searchOrderNo = ref('')
const dateRange = ref<[string, string] | null>(null)
const statusFilter = ref('')
const currentPage = ref(1)
const loading = ref(false)
const orderList = ref<any[]>([])
const total = ref(0)
const productMap = ref<Record<string, any>>({})
const previewImageUrl = ref('')
const showPreview = ref(false)

const CACHE_KEY = 'order_first_page_cache'
const CACHE_EXPIRE = 5 * 60 * 1000 // 5分钟缓存

const loadFromCache = () => {
  try {
    const cached = localStorage.getItem(CACHE_KEY)
    if (cached) {
      const { data, timestamp } = JSON.parse(cached)
      if (Date.now() - timestamp < CACHE_EXPIRE) {
        orderList.value = data.orderList
        productMap.value = data.productMap
        total.value = data.total
        return true
      }
    }
  } catch (e) {
    console.error('读取缓存失败:', e)
  }
  return false
}

const saveToCache = (orders: any[], products: any, totalCount: number) => {
  try {
    const data = {
      orderList: orders,
      productMap: products,
      total: totalCount
    }
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      data,
      timestamp: Date.now()
    }))
  } catch (e) {
    console.error('保存缓存失败:', e)
  }
}

const fetchOrders = async () => {
  const isFirstPage = currentPage.value === 1 && !searchOrderNo.value && !dateRange.value && !statusFilter.value
  
  if (isFirstPage && loadFromCache()) {
    return
  }
  
  loading.value = true
  try {
    await getToken()
    const params: any = {
      page: currentPage.value,
      page_size: 10,
      shopname: 'youlan_kids'
    }
    if (statusFilter.value) params.status = statusFilter.value
    if (dateRange.value?.[0]) params.begin_time = dateRange.value[0]
    if (dateRange.value?.[1]) params.end_time = dateRange.value[1]
    if (searchOrderNo.value) params.tid = searchOrderNo.value
    const res = await queryOrders(params)
    if (res.code === 200 && res.data?.code === 200) {
      const orders = res.data.data
      
      const allProductIds: string[] = []
      orders.forEach((order: any) => {
        if (order.product_list) {
          allProductIds.push(...order.product_list)
        }
      })
      
      if (allProductIds.length > 0) {
        const productRes = await batchGetProducts({ commodity_ids: allProductIds })
        if (productRes.code === 200 && productRes.data?.data) {
          productMap.value = {}
          productRes.data.data.forEach((product: any) => {
            productMap.value[product.commodity_id] = product
          })
        }
      }
      
      const newOrderList = orders.map((item: any) => ({
        id: item.user_id,
        orderNo: item.order_id,
        productList: item.product_list || [],
        buyer: item.receiver_name,
        receiver: item.receiver_name,
        phone: item.receiver_phone,
        address: `${item.province}${item.city}${item.county}${item.detailed_address}`,
        amount: item.order_amount,
        status: item.status,
        createTime: item.order_time
      }))
      
      orderList.value = newOrderList
      total.value = res.data.total
      
      if (isFirstPage) {
        saveToCache(newOrderList, productMap.value, res.data.total)
      }
    }
  } catch (error) {
    console.error('获取订单列表失败:', error)
    ElMessage.error('获取订单列表失败')
  } finally {
    loading.value = false
  }
}

const handleImagePreview = (url: string) => {
  previewImageUrl.value = url
  showPreview.value = true
}

onMounted(() => {
  fetchOrders()
})

const handleSearch = () => {
  currentPage.value = 1
  fetchOrders()
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchOrders()
}

const shortcuts = [
  { text: '今天', value: () => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const end = new Date()
    end.setHours(23, 59, 59, 999)
    return [today, end]
  }},
  { text: '昨天', value: () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    yesterday.setHours(0, 0, 0, 0)
    const end = new Date()
    end.setDate(end.getDate() - 1)
    end.setHours(23, 59, 59, 999)
    return [yesterday, end]
  }},
  { text: '近7天', value: () => {
    const start = new Date()
    start.setDate(start.getDate() - 7)
    start.setHours(0, 0, 0, 0)
    const end = new Date()
    end.setHours(23, 59, 59, 999)
    return [start, end]
  }},
  { text: '近30天', value: () => {
    const start = new Date()
    start.setDate(start.getDate() - 30)
    start.setHours(0, 0, 0, 0)
    const end = new Date()
    end.setHours(23, 59, 59, 999)
    return [start, end]
  }}
]

const setToday = () => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const end = new Date()
  end.setHours(23, 59, 59, 999)
  dateRange.value = [formatDate(today), formatDate(end)]
}

const setYesterday = () => {
  const yesterday = new Date()
  yesterday.setDate(yesterday.getDate() - 1)
  yesterday.setHours(0, 0, 0, 0)
  const end = new Date()
  end.setDate(end.getDate() - 1)
  end.setHours(23, 59, 59, 999)
  dateRange.value = [formatDate(yesterday), formatDate(end)]
}

const setLast7Days = () => {
  const start = new Date()
  start.setDate(start.getDate() - 7)
  start.setHours(0, 0, 0, 0)
  const end = new Date()
  end.setHours(23, 59, 59, 999)
  dateRange.value = [formatDate(start), formatDate(end)]
}

const setLast30Days = () => {
  const start = new Date()
  start.setDate(start.getDate() - 30)
  start.setHours(0, 0, 0, 0)
  const end = new Date()
  end.setHours(23, 59, 59, 999)
  dateRange.value = [formatDate(start), formatDate(end)]
}

const formatDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

const handleReset = () => {
  searchOrderNo.value = ''
  dateRange.value = null
  statusFilter.value = ''
  currentPage.value = 1
  fetchOrders()
  ElMessage.success('已重置')
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'primary',
    shipped: 'warning',
    delivered: 'success',
    canceled: 'info',
    processing: 'danger'
  }
  return map[status] || 'info'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '未发货',
    shipped: '已发货',
    delivered: '已送达',
    canceled: '已取消',
    processing: '售后中'
  }
  return map[status] || status
}

const viewDetail = (orderNo: string) => {
  router.push(`/order/${orderNo}`)
}

const shipOrder = (id: number) => {
  ElMessage.success('发货成功')
}
</script>

<style scoped>
.order-page {
  padding: 20px;
}

:deep(.el-table .el-table__row) {
  overflow: visible !important;
}

:deep(.el-table .cell) {
  overflow: visible !important;
}

.search-bar {
  display: flex;
  align-items: center;
}

.time-label {
  margin-left: 8px;
  font-size: 14px;
  color: #666;
  white-space: nowrap;
}

.shortcuts {
  display: flex;
  gap: 4px;
  margin-left: 8px;
}

.shortcuts .el-button {
  padding: 4px 8px;
  font-size: 12px;
}

.product-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.product-item {
  display: flex;
  align-items: center;
}

.product-image-wrapper {
  position: relative;
  margin-right: 8px;
  flex-shrink: 0;
  overflow: visible;
}

.product-image {
  width: 48px;
  height: 48px;
  background: #f5f5f5;
  flex-shrink: 0;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  transition: transform 0.2s;
}

.product-image:hover {
  transform: scale(1.1);
}

.product-image-preview {
  position: absolute;
  left: 60px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 9999;
  display: none;
  padding: 8px;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 0;
  box-shadow: none;
}

.product-image-wrapper:hover .product-image-preview {
  display: block;
}

.product-image-preview img {
  width: 100px;
  height: 100px;
  object-fit: cover;
}

.preview-image {
  max-width: 100%;
  max-height: 500px;
  object-fit: contain;
}

.product-detail {
  flex: 1;
  min-width: 0;
}

.product-name {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-info {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.receiver-info {
  display: flex;
  flex-direction: column;
}

.receiver-address {
  font-size: 12px;
  color: #999;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-image-container {
  display: flex;
  justify-content: center;
  align-items: center;
}

.preview-image {
  width: 400px;
  height: 400px;
  object-fit: contain;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

<style>
.white-tooltip {
  background: #fff !important;
  border: 1px solid #e5e7eb !important;
}

.white-tooltip .el-popper__arrow::before {
  background: #fff !important;
  border: 1px solid #e5e7eb !important;
}

.product-tooltip {
  padding: 12px;
  min-width: 200px;
}

.tooltip-image {
  width: 120px;
  height: 120px;
  object-fit: cover;
  margin-bottom: 12px;
  display: block;
}

.tooltip-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
  line-height: 1.4;
  word-break: break-all;
}

.tooltip-info {
  font-size: 13px;
  color: #666;
  line-height: 1.8;
}
</style>
