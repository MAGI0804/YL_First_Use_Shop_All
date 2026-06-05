<template>
  <div class="member-page">
    <div class="toolbar">
      <div class="filters">
        <el-input v-model="filters.mobile" placeholder="手机号" clearable class="filter-item" @keyup.enter="fetchMembers" />
        <el-input v-model="filters.member_no" placeholder="会员号" clearable class="filter-item" @keyup.enter="fetchMembers" />
        <el-input v-model="filters.manual_unique_code" placeholder="唯一字段" clearable class="filter-item" @keyup.enter="fetchMembers" />
        <el-select v-model="filters.status" placeholder="状态" clearable class="filter-item">
          <el-option label="正常" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
        <el-select v-model="filters.tag_id" placeholder="标签" clearable class="filter-item">
          <el-option v-for="tag in tags" :key="tag.id" :label="tag.name" :value="tag.id" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
      <el-button type="primary" @click="openAddDialog">新增会员</el-button>
    </div>

    <el-table v-loading="loading" :data="members" row-key="id" class="member-table">
      <el-table-column label="会员" min-width="190">
        <template #default="{ row }">
          <div class="member-main">
            <span class="member-name">{{ row.nickname || '-' }}</span>
            <span class="member-sub">{{ row.mobile }}</span>
            <span class="member-sub">{{ row.member_no }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="manual_unique_code" label="唯一字段" min-width="130" show-overflow-tooltip />
      <el-table-column label="金额" min-width="150">
        <template #default="{ row }">
          <div class="amount-lines">
            <span>下单 ¥{{ formatMoney(row.total_order_amount) }}</span>
            <span>已付 ¥{{ formatMoney(row.total_paid_amount) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="平台信息" min-width="190">
        <template #default="{ row }">
          <div class="amount-lines">
            <span>天猫 {{ row.tmall_id || '-' }} / ¥{{ formatMoney(row.tmall_amount) }}</span>
            <span>有赞 {{ row.youzan_id || '-' }} / ¥{{ formatMoney(row.youzan_amount) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
            {{ row.status === 'active' ? '正常' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" min-width="160" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="viewDetail(row.id)">详情</el-button>
          <el-button type="warning" link @click="openTagDialog(row)">标签</el-button>
          <el-button link @click="toggleStatus(row)">
            {{ row.status === 'active' ? '停用' : '启用' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchMembers"
      />
    </div>

    <el-dialog v-model="addDialogVisible" title="新增会员" width="560px" destroy-on-close>
      <el-form :model="memberForm" label-width="96px">
        <el-form-item label="手机号" required>
          <el-input v-model="memberForm.mobile" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="唯一字段">
          <el-input v-model="memberForm.manual_unique_code" placeholder="手动录入，系统校验唯一" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="memberForm.nickname" placeholder="会员昵称" />
        </el-form-item>
        <el-form-item label="天猫ID">
          <el-input v-model="memberForm.tmall_id" />
        </el-form-item>
        <el-form-item label="天猫金额">
          <el-input-number v-model="memberForm.tmall_amount" :min="0" :precision="2" class="full-input" />
        </el-form-item>
        <el-form-item label="有赞ID">
          <el-input v-model="memberForm.youzan_id" />
        </el-form-item>
        <el-form-item label="有赞金额">
          <el-input-number v-model="memberForm.youzan_amount" :min="0" :precision="2" class="full-input" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="memberForm.remarks" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitMember">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="tagDialogVisible" title="会员标签" width="480px" destroy-on-close>
      <el-select v-model="selectedTagIds" multiple placeholder="选择标签" class="full-input">
        <el-option v-for="tag in tags" :key="tag.id" :label="tag.name" :value="tag.id" />
      </el-select>
      <div class="new-tag-line">
        <el-input v-model="newTagName" placeholder="新增标签名" />
        <el-button @click="createTag">新增</el-button>
      </div>
      <template #footer>
        <el-button @click="tagDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="saveTags">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createMember,
  createMemberTag,
  queryMemberDetail,
  queryMembers,
  queryMemberTags,
  setMemberTags,
  updateMember,
  type MemberItem,
  type MemberTagItem
} from '@/api'

const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const members = ref<MemberItem[]>([])
const tags = ref<MemberTagItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const addDialogVisible = ref(false)
const tagDialogVisible = ref(false)
const currentMember = ref<MemberItem | null>(null)
const selectedTagIds = ref<number[]>([])
const newTagName = ref('')

const filters = reactive({
  mobile: '',
  member_no: '',
  manual_unique_code: '',
  status: '',
  tag_id: undefined as number | undefined
})

const memberForm = reactive({
  mobile: '',
  manual_unique_code: '',
  nickname: '',
  tmall_id: '',
  tmall_amount: 0,
  youzan_id: '',
  youzan_amount: 0,
  remarks: ''
})

const fetchTags = async () => {
  const res = await queryMemberTags({ page: 1, page_size: 100 })
  if (res.code === 200) {
    tags.value = res.data.items || []
  }
}

const fetchMembers = async () => {
  loading.value = true
  try {
    const res = await queryMembers({
      page: page.value,
      page_size: pageSize,
      mobile: filters.mobile || undefined,
      member_no: filters.member_no || undefined,
      manual_unique_code: filters.manual_unique_code || undefined,
      status: filters.status || undefined,
      tag_id: filters.tag_id
    })
    if (res.code === 200) {
      members.value = res.data.items || []
      total.value = res.data.total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '会员列表加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchMembers()
}

const handleReset = () => {
  filters.mobile = ''
  filters.member_no = ''
  filters.manual_unique_code = ''
  filters.status = ''
  filters.tag_id = undefined
  page.value = 1
  fetchMembers()
}

const openAddDialog = () => {
  Object.assign(memberForm, {
    mobile: '',
    manual_unique_code: '',
    nickname: '',
    tmall_id: '',
    tmall_amount: 0,
    youzan_id: '',
    youzan_amount: 0,
    remarks: ''
  })
  addDialogVisible.value = true
}

const submitMember = async () => {
  if (!memberForm.mobile.trim()) {
    ElMessage.warning('手机号不能为空')
    return
  }
  submitting.value = true
  try {
    await createMember({ ...memberForm })
    ElMessage.success('会员已新增')
    addDialogVisible.value = false
    fetchMembers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '新增会员失败')
  } finally {
    submitting.value = false
  }
}

const openTagDialog = async (row: MemberItem) => {
  currentMember.value = row
  selectedTagIds.value = []
  tagDialogVisible.value = true
  try {
    const res = await queryMemberDetail({ id: row.id })
    if (res.code === 200) {
      selectedTagIds.value = (res.data.detail.tags || []).map(tag => tag.id)
    }
  } catch {
    selectedTagIds.value = []
  }
}

const createTag = async () => {
  const name = newTagName.value.trim()
  if (!name) return
  try {
    await createMemberTag({ name })
    newTagName.value = ''
    await fetchTags()
    ElMessage.success('标签已新增')
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '新增标签失败')
  }
}

const saveTags = async () => {
  if (!currentMember.value) return
  submitting.value = true
  try {
    await setMemberTags({ member_id: currentMember.value.id, tag_ids: selectedTagIds.value })
    ElMessage.success('标签已保存')
    tagDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '保存标签失败')
  } finally {
    submitting.value = false
  }
}

const toggleStatus = async (row: MemberItem) => {
  const nextStatus = row.status === 'active' ? 'disabled' : 'active'
  const action = nextStatus === 'active' ? '启用' : '停用'
  try {
    await ElMessageBox.confirm(`确认${action}会员 ${row.mobile}？`, '提示', { type: 'warning' })
    await updateMember({ ...row, status: nextStatus })
    ElMessage.success(`会员已${action}`)
    fetchMembers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || `${action}失败`)
    }
  }
}

const viewDetail = (id: number) => {
  router.push(`/member/${id}`)
}

const formatMoney = (value: number | string | undefined | null) => Number(value || 0).toFixed(2)

onMounted(async () => {
  await fetchTags()
  fetchMembers()
})
</script>

<style scoped>
.member-page {
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.filter-item {
  width: 150px;
}

.member-table {
  width: 100%;
  margin-top: 18px;
}

.member-main,
.amount-lines {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.member-name {
  font-weight: 600;
  color: #1f2937;
}

.member-sub,
.amount-lines span {
  font-size: 12px;
  color: #6b7280;
}

.pagination {
  margin-top: 18px;
  display: flex;
  justify-content: flex-end;
}

.full-input {
  width: 100%;
}

.new-tag-line {
  display: flex;
  gap: 8px;
  margin-top: 14px;
}
</style>
