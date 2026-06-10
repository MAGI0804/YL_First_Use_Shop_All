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
      <div class="toolbar-actions">
        <el-button @click="openImportDrawer">导入名单</el-button>
        <el-button type="primary" @click="openAddDialog">新增会员</el-button>
      </div>
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

    <el-drawer v-model="importDrawerVisible" title="导入会员名单" size="820px" destroy-on-close @closed="resetImportState">
      <div class="import-panel">
        <div class="import-actions">
          <el-upload
            accept=".xlsx"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleImportFileChange"
          >
            <el-button type="primary">导入Excel</el-button>
          </el-upload>
          <el-button @click="downloadImportTemplate">下载模板</el-button>
          <el-button :disabled="!importFile" :loading="matching" @click="matchImportFile">匹配名单</el-button>
        </div>

        <div v-if="importFile" class="import-file">
          当前文件：{{ importFile.name }}
        </div>

        <el-alert
          v-if="importResult"
          class="import-summary"
          :closable="false"
          type="info"
          show-icon
        >
          <template #title>
            共读取 {{ importResult.total_rows }} 行，匹配可导入 {{ importResult.matched_count }} 行，不符合 {{ importResult.invalid_count }} 行
          </template>
        </el-alert>

        <el-table v-if="importRows.length" :data="importRows" max-height="480" class="import-table">
          <el-table-column prop="row_index" label="Excel行" width="80" />
          <el-table-column prop="mobile" label="手机号" min-width="130" />
          <el-table-column prop="manual_unique_code" label="唯一字段" min-width="130" show-overflow-tooltip />
          <el-table-column prop="nickname" label="昵称" min-width="110" show-overflow-tooltip />
          <el-table-column label="平台金额" min-width="180">
            <template #default="{ row }">
              <div class="amount-lines">
                <span>天猫 {{ row.tmall_id || '-' }} / ¥{{ formatMoney(row.tmall_amount) }}</span>
                <span>有赞 {{ row.youzan_id || '-' }} / ¥{{ formatMoney(row.youzan_amount) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="匹配结果" min-width="190">
            <template #default="{ row }">
              <el-tag v-if="row.matched" type="success" size="small">可导入</el-tag>
              <div v-else class="import-errors">
                <el-tag type="danger" size="small">不符合</el-tag>
                <span>{{ row.errors.join('；') }}</span>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else class="import-empty" description="请先导入Excel并点击匹配名单" />
      </div>

      <template #footer>
        <el-button @click="importDrawerVisible = false">关闭</el-button>
        <el-button
          type="primary"
          :disabled="matchedImportRows.length === 0"
          :loading="confirmingImport"
          @click="confirmImportRows"
        >
          确认导入{{ matchedImportRows.length ? ` ${matchedImportRows.length} 条` : '' }}
        </el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import {
  confirmMemberImport,
  createMember,
  createMemberTag,
  downloadMemberImportTemplate,
  matchMemberImportFile,
  queryMemberDetail,
  queryMembers,
  queryMemberTags,
  setMemberTags,
  updateMember,
  type MemberItem,
  type MemberImportRow,
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
const importDrawerVisible = ref(false)
const currentMember = ref<MemberItem | null>(null)
const selectedTagIds = ref<number[]>([])
const newTagName = ref('')
const importFile = ref<File | null>(null)
const importResult = ref<{ total_rows: number; matched_count: number; invalid_count: number } | null>(null)
const importRows = ref<MemberImportRow[]>([])
const matching = ref(false)
const confirmingImport = ref(false)

const matchedImportRows = computed(() => importRows.value.filter(row => row.matched))

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

const openImportDrawer = () => {
  importDrawerVisible.value = true
}

const resetImportState = () => {
  importFile.value = null
  importResult.value = null
  importRows.value = []
  matching.value = false
  confirmingImport.value = false
}

const handleImportFileChange = (uploadFile: UploadFile) => {
  const file = uploadFile.raw
  if (!file) return
  if (!/\.xlsx$/i.test(file.name)) {
    ElMessage.warning('请上传 .xlsx 格式的Excel文件')
    return
  }
  importFile.value = file
  importResult.value = null
  importRows.value = []
}

const downloadImportTemplate = async () => {
  try {
    const blob = await downloadMemberImportTemplate()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = '会员导入模板.xlsx'
    link.click()
    URL.revokeObjectURL(url)
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '模板下载失败')
  }
}

const matchImportFile = async () => {
  if (!importFile.value) {
    ElMessage.warning('请先选择Excel文件')
    return
  }
  const formData = new FormData()
  formData.append('file', importFile.value)
  matching.value = true
  try {
    const res = await matchMemberImportFile(formData)
    if (res.code === 200) {
      importResult.value = {
        total_rows: res.data.result.total_rows,
        matched_count: res.data.result.matched_count,
        invalid_count: res.data.result.invalid_count
      }
      importRows.value = res.data.result.items || []
      if (res.data.result.matched_count === 0) {
        ElMessage.warning('没有匹配到可导入的数据')
      } else {
        ElMessage.success(`已匹配 ${res.data.result.matched_count} 条可导入数据`)
      }
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '名单匹配失败')
  } finally {
    matching.value = false
  }
}

const confirmImportRows = async () => {
  if (matchedImportRows.value.length === 0) {
    ElMessage.warning('没有可导入的数据')
    return
  }
  try {
    await ElMessageBox.confirm(`确认导入 ${matchedImportRows.value.length} 条会员？`, '确认导入', { type: 'warning' })
  } catch {
    return
  }
  confirmingImport.value = true
  try {
    const items = matchedImportRows.value.map(row => ({
      mobile: row.mobile,
      manual_unique_code: row.manual_unique_code,
      nickname: row.nickname,
      tmall_id: row.tmall_id,
      tmall_amount: row.tmall_amount,
      youzan_id: row.youzan_id,
      youzan_amount: row.youzan_amount,
      remarks: row.remarks
    }))
    const res = await confirmMemberImport({ items })
    ElMessage.success(`已导入 ${res.data.result.imported_count} 条会员`)
    importDrawerVisible.value = false
    page.value = 1
    fetchMembers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '确认导入失败，请重新匹配后再试')
  } finally {
    confirmingImport.value = false
  }
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

.toolbar-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
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

.import-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.import-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.import-file {
  font-size: 13px;
  color: #4b5563;
}

.import-summary {
  margin-top: 2px;
}

.import-table {
  width: 100%;
}

.import-errors {
  display: flex;
  flex-direction: column;
  gap: 4px;
  color: #b91c1c;
  font-size: 12px;
  line-height: 1.4;
}

.import-empty {
  padding: 44px 0;
}

@media (max-width: 900px) {
  .toolbar {
    flex-direction: column;
  }

  .toolbar-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
