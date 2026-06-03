<template>
  <div class="users-page">
    <div class="toolbar">
      <el-input v-model="filters.mobile" clearable placeholder="手机号" class="filter-input" />
      <el-select v-model="filters.status" clearable placeholder="状态" class="filter-select">
        <el-option label="待激活" value="pending" />
        <el-option label="启用" value="active" />
        <el-option label="停用" value="disabled" />
      </el-select>
      <el-button :icon="Search" @click="loadUsers">查询</el-button>
      <el-button type="primary" :icon="Plus" @click="openInviteDialog">新增账号</el-button>
    </div>

    <el-table :data="users" border v-loading="loading" empty-text="暂无账号">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="operator_no" label="运营编号" min-width="170" />
      <el-table-column prop="nickname" label="账户名" min-width="120" />
      <el-table-column prop="mobile" label="手机号" width="140" />
      <el-table-column prop="role" label="角色" width="110">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">{{ row.role === 'admin' ? '管理员' : '运营' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="level" label="权限等级" width="100" />
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusMeta[row.status]?.type || 'info'">{{ statusMeta[row.status]?.label || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="210" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status !== 'active'" size="small" @click="changeStatus(row, 'active')">启用</el-button>
          <el-button v-if="row.status !== 'disabled'" size="small" type="danger" @click="changeStatus(row, 'disabled')">停用</el-button>
          <el-button v-if="row.status !== 'pending'" size="small" @click="changeStatus(row, 'pending')">待激活</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="filters.page"
        v-model:page-size="filters.page_size"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadUsers"
        @size-change="loadUsers"
      />
    </div>

    <el-dialog v-model="inviteDialogVisible" title="新增后台账号" width="420px">
      <el-form ref="inviteFormRef" :model="inviteForm" :rules="inviteRules" label-width="84px">
        <el-form-item label="账户名" prop="nickname">
          <el-input v-model="inviteForm.nickname" placeholder="员工或运营名称" />
        </el-form-item>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="inviteForm.mobile" placeholder="登录手机号" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="inviteForm.role" class="full-width">
            <el-option label="运营" value="operation" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限等级" prop="level">
          <el-input-number v-model="inviteForm.level" :min="1" :max="9" class="full-width" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="inviteForm.remarks" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="inviteDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitInvite">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, FormInstance } from 'element-plus'
import type { FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { inviteBackendUser, queryBackendUsers, updateBackendUserStatus } from '@/api'
import type { BackendUserSession } from '@/api'

const loading = ref(false)
const saving = ref(false)
const users = ref<BackendUserSession[]>([])
const total = ref(0)
const inviteDialogVisible = ref(false)
const inviteFormRef = ref<FormInstance>()

const filters = reactive({
  mobile: '',
  status: '',
  page: 1,
  page_size: 20
})

const inviteForm = reactive({
  nickname: '',
  mobile: '',
  role: 'operation',
  level: 1,
  remarks: ''
})

const statusMeta: Record<string, { label: string; type: 'success' | 'warning' | 'info' | 'danger' }> = {
  pending: { label: '待激活', type: 'warning' },
  active: { label: '启用', type: 'success' },
  disabled: { label: '停用', type: 'danger' }
}

const inviteRules: FormRules = {
  nickname: [{ required: true, message: '请输入账户名', trigger: 'blur' }],
  mobile: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
  level: [{ required: true, message: '请输入权限等级', trigger: 'change' }]
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await queryBackendUsers(filters)
    users.value = res.data?.items || []
    total.value = res.data?.total || 0
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '账号列表加载失败')
  } finally {
    loading.value = false
  }
}

const openInviteDialog = () => {
  inviteForm.nickname = ''
  inviteForm.mobile = ''
  inviteForm.role = 'operation'
  inviteForm.level = 1
  inviteForm.remarks = ''
  inviteDialogVisible.value = true
}

const submitInvite = async () => {
  if (!inviteFormRef.value) return
  const valid = await inviteFormRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    await inviteBackendUser(inviteForm)
    ElMessage.success('账号已添加，等待首次注册激活')
    inviteDialogVisible.value = false
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '账号添加失败')
  } finally {
    saving.value = false
  }
}

const changeStatus = async (row: BackendUserSession, status: 'pending' | 'active' | 'disabled') => {
  try {
    await updateBackendUserStatus({ id: row.id, status })
    ElMessage.success('状态已更新')
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '状态更新失败')
  }
}

onMounted(loadUsers)
</script>

<style scoped>
.users-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  padding: 14px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.filter-input {
  width: 180px;
}

.filter-select {
  width: 130px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  padding: 12px 0;
}

.full-width {
  width: 100%;
}
</style>
