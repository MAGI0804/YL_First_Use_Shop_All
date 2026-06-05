<template>
  <div class="member-detail">
    <div class="page-header">
      <el-button link @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <span class="page-title">会员详情</span>
    </div>

    <el-skeleton v-if="loading" :rows="8" animated />

    <div v-else class="detail-content">
      <el-row :gutter="16">
        <el-col :span="8">
          <section class="section-block">
            <div class="section-title">基础信息</div>
            <div class="info-row"><span>会员号</span><strong>{{ member.member_no }}</strong></div>
            <div class="info-row"><span>唯一字段</span><strong>{{ member.manual_unique_code || '-' }}</strong></div>
            <div class="info-row"><span>昵称</span><strong>{{ member.nickname || '-' }}</strong></div>
            <div class="info-row"><span>手机号</span><strong>{{ member.mobile }}</strong></div>
            <div class="info-row"><span>关联用户ID</span><strong>{{ member.user_id || '-' }}</strong></div>
            <div class="info-row">
              <span>状态</span>
              <el-tag :type="member.status === 'active' ? 'success' : 'info'" size="small">
                {{ member.status === 'active' ? '正常' : '停用' }}
              </el-tag>
            </div>
          </section>
        </el-col>
        <el-col :span="8">
          <section class="section-block">
            <div class="section-title">金额统计</div>
            <div class="info-row"><span>总下单金额</span><strong>¥{{ formatMoney(member.total_order_amount) }}</strong></div>
            <div class="info-row"><span>已支付金额</span><strong>¥{{ formatMoney(member.total_paid_amount) }}</strong></div>
            <div class="info-row"><span>天猫</span><strong>{{ member.tmall_id || '-' }} / ¥{{ formatMoney(member.tmall_amount) }}</strong></div>
            <div class="info-row"><span>有赞</span><strong>{{ member.youzan_id || '-' }} / ¥{{ formatMoney(member.youzan_amount) }}</strong></div>
          </section>
        </el-col>
        <el-col :span="8">
          <section class="section-block">
            <div class="section-title with-action">
              <span>标签</span>
              <el-button size="small" @click="tagDialogVisible = true">编辑</el-button>
            </div>
            <div class="tag-list">
              <el-tag v-for="tag in memberTags" :key="tag.id" size="small">{{ tag.name }}</el-tag>
              <span v-if="memberTags.length === 0" class="empty-text">暂无标签</span>
            </div>
          </section>
        </el-col>
      </el-row>

      <section class="section-block wide">
        <div class="section-title with-action">
          <span>购物车</span>
          <div>
            <el-button size="small" @click="cartDialogVisible = true">新增商品</el-button>
            <el-button size="small" type="primary" :disabled="cartItems.length === 0" @click="openOrderDialog">从购物车下单</el-button>
          </div>
        </div>
        <el-table :data="cartItems" size="small">
          <el-table-column prop="commodity_code" label="商品编码" min-width="180" />
          <el-table-column prop="quantity" label="数量" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.quantity" :min="0" :precision="0" size="small" @change="updateCart(row)" />
            </template>
          </el-table-column>
          <el-table-column prop="added_time" label="加入时间" min-width="160" />
          <el-table-column label="操作" width="90">
            <template #default="{ row }">
              <el-button type="danger" link @click="deleteCart(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="section-block wide">
        <div class="section-title">操作记录</div>
        <el-table :data="operationLogs" size="small">
          <el-table-column prop="created_at" label="时间" min-width="160" />
          <el-table-column prop="operator_mobile" label="操作人" width="130" />
          <el-table-column prop="action" label="动作" min-width="190" />
          <el-table-column prop="target_id" label="对象" min-width="120" />
          <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        </el-table>
      </section>
    </div>

    <el-dialog v-model="tagDialogVisible" title="编辑标签" width="480px">
      <el-select v-model="selectedTagIds" multiple class="full-input" placeholder="选择标签">
        <el-option v-for="tag in allTags" :key="tag.id" :label="tag.name" :value="tag.id" />
      </el-select>
      <template #footer>
        <el-button @click="tagDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveTags">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="cartDialogVisible" title="新增购物车商品" width="420px">
      <el-form label-width="90px">
        <el-form-item label="商品编码">
          <el-input v-model="cartForm.commodity_code" />
        </el-form-item>
        <el-form-item label="数量">
          <el-input-number v-model="cartForm.quantity" :min="1" :precision="0" class="full-input" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cartDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addCart">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="orderDialogVisible" title="后台代会员下单" width="560px">
      <el-form :model="orderForm" label-width="96px">
        <el-form-item label="收货人" required><el-input v-model="orderForm.receiver_name" /></el-form-item>
        <el-form-item label="手机号" required><el-input v-model="orderForm.receiver_phone" /></el-form-item>
        <el-form-item label="省份" required><el-input v-model="orderForm.province" /></el-form-item>
        <el-form-item label="城市" required><el-input v-model="orderForm.city" /></el-form-item>
        <el-form-item label="区县" required><el-input v-model="orderForm.county" /></el-form-item>
        <el-form-item label="详细地址" required><el-input v-model="orderForm.detailed_address" /></el-form-item>
        <el-form-item label="后台备注"><el-input v-model="orderForm.backend_remark" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orderDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="createOrder">下单</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import {
  addMemberCartItem,
  backendCreateOrder,
  deleteMemberCartItems,
  queryMemberCart,
  queryMemberDetail,
  queryMemberTags,
  queryOperationLogs,
  setMemberTags,
  updateMemberCartQuantity,
  type MemberItem,
  type MemberTagItem,
  type OperationLogItem
} from '@/api'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const submitting = ref(false)
const tagDialogVisible = ref(false)
const cartDialogVisible = ref(false)
const orderDialogVisible = ref(false)
const member = ref<MemberItem>({} as MemberItem)
const memberTags = ref<MemberTagItem[]>([])
const allTags = ref<MemberTagItem[]>([])
const selectedTagIds = ref<number[]>([])
const cartItems = ref<any[]>([])
const operationLogs = ref<OperationLogItem[]>([])

const cartForm = reactive({ commodity_code: '', quantity: 1 })
const orderForm = reactive({
  receiver_name: '',
  receiver_phone: '',
  province: '',
  city: '',
  county: '',
  detailed_address: '',
  backend_remark: ''
})

const memberId = Number(route.params.id)

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await queryMemberDetail({ id: memberId })
    if (res.code === 200) {
      member.value = res.data.detail.member
      memberTags.value = res.data.detail.tags || []
      selectedTagIds.value = memberTags.value.map(tag => tag.id)
      await Promise.all([fetchCart(), fetchLogs()])
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '会员详情加载失败')
  } finally {
    loading.value = false
  }
}

const fetchTags = async () => {
  const res = await queryMemberTags({ page: 1, page_size: 100 })
  if (res.code === 200) allTags.value = res.data.items || []
}

const fetchCart = async () => {
  if (!member.value.id) return
  const res: any = await queryMemberCart({ member_id: member.value.id })
  if (res.code === 200) {
    cartItems.value = res.data.cart?.cart_items || []
  }
}

const fetchLogs = async () => {
  const res = await queryOperationLogs({ page: 1, page_size: 20, member_id: memberId })
  if (res.code === 200) operationLogs.value = res.data.items || []
}

const saveTags = async () => {
  await setMemberTags({ member_id: memberId, tag_ids: selectedTagIds.value })
  tagDialogVisible.value = false
  ElMessage.success('标签已保存')
  fetchDetail()
}

const addCart = async () => {
  if (!cartForm.commodity_code.trim()) {
    ElMessage.warning('商品编码不能为空')
    return
  }
  await addMemberCartItem({ member_id: memberId, commodity_code: cartForm.commodity_code, quantity: cartForm.quantity })
  cartDialogVisible.value = false
  cartForm.commodity_code = ''
  cartForm.quantity = 1
  ElMessage.success('购物车已新增')
  await Promise.all([fetchCart(), fetchLogs()])
}

const updateCart = async (row: any) => {
  await updateMemberCartQuantity({ member_id: memberId, commodity_code: row.commodity_code, quantity: row.quantity })
  ElMessage.success('购物车数量已更新')
  await Promise.all([fetchCart(), fetchLogs()])
}

const deleteCart = async (row: any) => {
  await ElMessageBox.confirm(`确认删除商品 ${row.commodity_code}？`, '提示', { type: 'warning' })
  await deleteMemberCartItems({ member_id: memberId, commodity_codes: [row.commodity_code] })
  ElMessage.success('购物车商品已删除')
  await Promise.all([fetchCart(), fetchLogs()])
}

const openOrderDialog = () => {
  orderForm.receiver_name = member.value.nickname || ''
  orderForm.receiver_phone = member.value.mobile || ''
  orderForm.province = ''
  orderForm.city = ''
  orderForm.county = ''
  orderForm.detailed_address = ''
  orderForm.backend_remark = ''
  orderDialogVisible.value = true
}

const createOrder = async () => {
  submitting.value = true
  try {
    await backendCreateOrder({
      member_id: memberId,
      ...orderForm,
      items: cartItems.value.map(item => ({
        commodity_code: item.commodity_code,
        quantity: item.quantity
      }))
    })
    orderDialogVisible.value = false
    ElMessage.success('后台代下单成功')
    await Promise.all([fetchDetail(), fetchCart(), fetchLogs()])
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '后台代下单失败')
  } finally {
    submitting.value = false
  }
}

const formatMoney = (value: number | string | undefined | null) => Number(value || 0).toFixed(2)
const goBack = () => router.back()

onMounted(async () => {
  await fetchTags()
  fetchDetail()
})
</script>

<style scoped>
.member-detail {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.section-block {
  border: 1px solid #ebeef5;
  padding: 16px;
  min-height: 180px;
  background: #fff;
}

.section-block.wide {
  margin-top: 16px;
  min-height: auto;
}

.section-title {
  font-weight: 600;
  margin-bottom: 12px;
  color: #1f2937;
}

.section-title.with-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #f3f4f6;
}

.info-row span {
  color: #6b7280;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.empty-text {
  color: #9ca3af;
  font-size: 13px;
}

.full-input {
  width: 100%;
}
</style>
