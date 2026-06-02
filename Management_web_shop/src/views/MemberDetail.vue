<template>
  <div class="member-detail">
    <div class="page-header">
      <el-button link @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <span class="page-title">会员详情</span>
    </div>

    <div class="detail-content">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card class="info-card">
            <template #header>
              <span>基本信息</span>
            </template>
            <div class="info-row">
              <span class="label">会员昵称：</span>
              <span class="value">{{ member.username }}</span>
            </div>
            <div class="info-row">
              <span class="label">手机号：</span>
              <span class="value">{{ member.phone }}</span>
            </div>
            <div class="info-row">
              <span class="label">会员等级：</span>
              <el-tag :type="getLevelType(member.level)" size="small">
                {{ getLevelText(member.level) }}
              </el-tag>
            </div>
            <div class="info-row">
              <span class="label">积分：</span>
              <span class="value">{{ member.points }}</span>
            </div>
            <div class="info-row">
              <span class="label">注册时间：</span>
              <span class="value">{{ member.createTime }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="info-card">
            <template #header>
              <span>标签</span>
            </template>
            <div class="tags-container">
              <el-tag v-for="tag in member.tags" :key="tag" style="margin-right: 8px; margin-bottom: 8px;">
                {{ getTagText(tag) }}
              </el-tag>
              <el-button size="small" @click="tagDialogVisible = true">编辑标签</el-button>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="info-card">
            <template #header>
              <span>账户统计</span>
            </template>
            <div class="info-row">
              <span class="label">订单总数：</span>
              <span class="value">{{ member.orderCount }}</span>
            </div>
            <div class="info-row">
              <span class="label">消费总额：</span>
              <span class="value">¥{{ member.totalSpent }}</span>
            </div>
            <div class="info-row">
              <span class="label">平均客单价：</span>
              <span class="value">¥{{ member.avgOrder }}</span>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-card class="order-card" style="margin-top: 20px;">
        <template #header>
          <span>最近订单</span>
        </template>
        <el-table :data="member.orders">
          <el-table-column prop="orderNo" label="订单号" width="180" />
          <el-table-column prop="amount" label="金额" width="100">
            <template #default="{ row }">
              ¥{{ row.amount }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'completed' ? 'success' : 'info'" size="small">
                {{ row.status === 'completed' ? '已完成' : '进行中' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createTime" label="下单时间" width="180" />
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="tagDialogVisible" title="编辑标签" width="400px">
      <el-select v-model="selectedTags" multiple placeholder="请选择标签" style="width: 100%;">
        <el-option label="活跃用户" value="active" />
        <el-option label="高消费" value="high消费" />
        <el-option label="新用户" value="new" />
        <el-option label="沉睡用户" value="sleep" />
        <el-option label="VIP" value="vip" />
      </el-select>
      <template #footer>
        <el-button @click="tagDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSaveTag">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'

const router = useRouter()
const tagDialogVisible = ref(false)
const selectedTags = ref<string[]>([])

const member = reactive({
  id: 1,
  username: '张三',
  phone: '138****8888',
  level: 'gold',
  points: 5000,
  tags: ['active', 'vip'],
  createTime: '2024-03-15 10:30:00',
  orderCount: 25,
  totalSpent: '8,680',
  avgOrder: '347',
  orders: [
    { orderNo: 'ORDER20240427001', amount: '299', status: 'completed', createTime: '2024-04-27 10:30:00' },
    { orderNo: 'ORDER20240426002', amount: '159', status: 'completed', createTime: '2024-04-26 14:20:00' },
    { orderNo: 'ORDER20240425003', amount: '499', status: 'completed', createTime: '2024-04-25 09:15:00' },
  ]
})

selectedTags.value = [...member.tags]

const getLevelType = (level: string) => {
  const map: Record<string, string> = {
    normal: 'info',
    silver: '',
    gold: 'warning',
    black: 'danger'
  }
  return map[level] || 'info'
}

const getLevelText = (level: string) => {
  const map: Record<string, string> = {
    normal: 'lv1',
    silver: 'lv2',
    gold: 'lv3',
    black: 'lv4'
  }
  return map[level] || level
}

const getTagText = (tag: string) => {
  const map: Record<string, string> = {
    active: '活跃用户',
    high消费: '高消费',
    new: '新用户',
    sleep: '沉睡用户',
    vip: 'VIP'
  }
  return map[tag] || tag
}

const goBack = () => {
  router.back()
}

const confirmSaveTag = () => {
  member.tags = [...selectedTags.value]
  ElMessage.success('标签保存成功')
  tagDialogVisible.value = false
}
</script>

<style scoped>
.member-detail {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.page-title {
  font-size: 18px;
  font-weight: 500;
  color: #1a1a1a;
}

.info-card :deep(.el-card__header) {
  font-weight: 500;
}

.info-row {
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f5f5f5;
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  width: 100px;
  color: #999;
}

.value {
  flex: 1;
  color: #1a1a1a;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}
</style>
