<template>
  <div class="order-detail">
    <div class="page-header">
      <el-button link @click="router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <span class="page-title">订单详情</span>
    </div>

    <div v-if="loading" class="loading-container">
      <el-loading-spinner style="width: 32px; height: 32px;" />
      <span style="margin-left: 10px;">加载中...</span>
    </div>

    <div v-else class="detail-content">
      <el-card class="main-card">
        <!-- 顶部行：订单号+复制、下单时间 -->
        <div class="top-row">
          <div class="order-id-section">
            <span class="label-text">订单号</span>
            <span class="order-id">{{ orderData.order_id }}</span>
            <el-button type="primary" link size="small" @click="copyOrderId">
              <el-icon><DocumentCopy /></el-icon>
              复制
            </el-button>
          </div>
          <div class="order-time-section">
            <span class="label-text">下单时间</span>
            <span class="time-text">{{ orderData.order_time }}</span>
          </div>
        </div>

        <div class="divider-line"></div>

        <!-- 订单状态和收货人信息 -->
        <div class="info-section">
          <div class="info-item">
            <span class="label-text">订单状态</span>
            <el-tag :type="getStatusType(orderData.status)" size="small">{{ getStatusText(orderData.status) }}</el-tag>
          </div>
          <div class="info-item">
            <span class="label-text">收货人</span>
            <span class="value-text">{{ orderData.receiver_name }}</span>
          </div>
          <div class="info-item">
            <span class="label-text">联系电话</span>
            <span class="value-text">{{ orderData.receiver_phone }}</span>
          </div>
          <div class="info-item full-width">
            <span class="label-text">收货地址</span>
            <span class="value-text">{{ orderData.province }}{{ orderData.city }}{{ orderData.county }}{{ orderData.detailed_address }}</span>
          </div>
          <div v-if="orderData.express_company || orderData.express_number" class="info-item">
            <span class="label-text">快递信息</span>
            <span class="value-text">
              {{ orderData.express_company }} 
              <span v-if="orderData.express_number">{{ orderData.express_number }}</span>
            </span>
          </div>
        </div>

        <div class="divider-line"></div>

        <!-- 商品信息 -->
        <div class="section-title">商品信息</div>
        <el-table :data="productList" size="small" style="width: 100%;">
          <el-table-column label="商品图片" width="80">
            <template #default="{ row }">
              <div v-if="productMap[row]?.image" class="product-image-wrapper">
                <img :src="productMap[row].image" class="product-image" />
              </div>
              <div v-else class="product-image"></div>
            </template>
          </el-table-column>
          <el-table-column label="商品ID" width="180">
            <template #default="{ row }">
              {{ row }}
            </template>
          </el-table-column>
          <el-table-column label="商品名称" min-width="200">
            <template #default="{ row }">
              <div class="product-name-text">{{ productMap[row]?.name || '-' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="规格" width="140">
            <template #default="{ row }">
              <span v-if="productMap[row]">
                {{ productMap[row].color }} / {{ productMap[row].size }}
              </span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="单价" width="100" align="right">
            <template #default="{ row }">
              ¥{{ productMap[row]?.price || '-' }}
            </template>
          </el-table-column>
        </el-table>

        <div class="divider-line"></div>

        <!-- 金额信息 -->
        <div class="amount-section">
          <div class="amount-item">
            <span>商品金额</span>
            <span>¥{{ orderData.order_amount?.toFixed(2) || '0.00' }}</span>
          </div>
          <div class="amount-item">
            <span>运费</span>
            <span>¥0.00</span>
          </div>
          <div class="amount-item total">
            <span>实付金额</span>
            <span>¥{{ orderData.order_amount?.toFixed(2) || '0.00' }}</span>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, DocumentCopy } from '@element-plus/icons-vue'
import { queryOrderDetail, getToken, batchGetProducts } from '@/api'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const orderData = ref<any>({})
const productList = ref<string[]>([])
const productMap = ref<Record<string, any>>({})

const copyOrderId = () => {
  if (orderData.value.order_id) {
    navigator.clipboard.writeText(orderData.value.order_id)
    ElMessage.success('订单号已复制')
  }
}

const fetchOrderDetail = async () => {
  loading.value = true
  try {
    await getToken()
    const orderId = route.params.id as string
    console.log('订单号:', orderId)
    
    const requestParams = {
      order_id: orderId,
      inquired_list: ['order_id', 'order_amount', 'product_list', 'province', 'city', 'county', 'detailed_address', 'status', 'remarks', 'order_time', 'receiver_phone', 'express_company', 'express_number'],
      shopname: 'youlan_kids'
    }
    console.log('请求参数:', requestParams)
    
    const res = await queryOrderDetail(requestParams)
    console.log('响应数据:', res)
    
    // 检查响应格式
    let orderDetailData = null
    if (res.code === 200 && res.data?.data) {
      orderDetailData = res.data.data
    }
    
    if (orderDetailData) {
      orderData.value = orderDetailData
      productList.value = orderDetailData.product_list || []
      
      if (productList.value.length > 0) {
        const productRes = await batchGetProducts({ commodity_ids: productList.value })
        console.log('商品数据:', productRes)
        if (productRes.code === 200 && productRes.data?.data) {
          productMap.value = {}
          productRes.data.data.forEach((product: any) => {
            productMap.value[product.commodity_id] = product
          })
        }
      }
    } else {
      console.error('业务错误:', res)
      ElMessage.error('获取订单详情失败: ' + (res.msg || res.data?.message || '未知错误'))
    }
  } catch (error: any) {
    console.error('获取订单详情失败:', error)
    console.error('错误详情:', error.response?.data)
    ElMessage.error('获取订单详情失败: ' + (error.response?.data?.msg || error.message))
  } finally {
    loading.value = false
  }
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

onMounted(() => {
  fetchOrderDetail()
})
</script>

<style scoped>
.order-detail {
  padding: 16px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.page-title {
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
}

.main-card :deep(.el-card__body) {
  padding: 12px 16px;
}

.top-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-id-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-text {
  font-size: 12px;
  color: #999;
  min-width: 56px;
}

.order-id {
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.order-time-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-text {
  font-size: 13px;
  color: #333;
}

.divider-line {
  height: 1px;
  background: #f0f0f0;
  margin: 12px 0;
}

.info-section {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 24px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-item.full-width {
  width: 100%;
}

.value-text {
  font-size: 13px;
  color: #333;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.product-image-wrapper {
  width: 44px;
  height: 44px;
}

.product-image {
  width: 44px;
  height: 44px;
  background: #f5f5f5;
  border-radius: 4px;
  object-fit: cover;
}

.amount-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.amount-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
}

.amount-item.total {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin-top: 4px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}
</style>
