<template>
  <div class="users-page">
    <div class="toolbar">
      <el-input v-model="filters.mobile" clearable placeholder="手机号" class="filter-input" />
      <el-select v-model="filters.role" clearable placeholder="身份" class="filter-select">
        <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-select v-model="filters.status" clearable placeholder="状态" class="filter-select">
        <el-option label="待激活" value="pending" />
        <el-option label="启用" value="active" />
        <el-option label="停用" value="disabled" />
      </el-select>
      <el-button :icon="Search" @click="loadUsers">查询</el-button>
      <el-button type="primary" :icon="Plus" @click="openCreateDialog">新增账号</el-button>
    </div>

    <el-table :data="users" border v-loading="loading" empty-text="暂无账号">
      <el-table-column prop="operator_no" label="运营编号" min-width="170" />
      <el-table-column prop="nickname" label="账户名" min-width="120" />
      <el-table-column prop="mobile" label="手机号" width="140" />
      <el-table-column prop="role" label="身份" width="110">
        <template #default="{ row }">
          <el-tag :type="roleMeta[row.role]?.type || 'info'">{{ roleMeta[row.role]?.label || row.role }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusMeta[row.status]?.type || 'info'">{{ statusMeta[row.status]?.label || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="可访问页面" min-width="260">
        <template #default="{ row }">
          <div class="permission-tags">
            <el-tag v-for="permission in row.permissions" :key="permission" size="small" effect="plain">
              {{ permissionLabel(permission) }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
          <el-button
            v-if="row.status !== 'disabled'"
            size="small"
            type="danger"
            @click="quickDisable(row)"
          >
            停用
          </el-button>
          <el-button v-else size="small" type="success" @click="quickEnable(row)">启用</el-button>
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

    <el-dialog v-model="dialogVisible" :title="editingUser ? '编辑账号' : '新增账号'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="84px">
        <el-form-item label="账户名" prop="nickname">
          <el-input v-model="form.nickname" placeholder="员工或运营名称" />
        </el-form-item>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="form.mobile" :disabled="!!editingUser" placeholder="登录手机号" />
        </el-form-item>
        <el-form-item label="身份" prop="role">
          <el-segmented v-model="form.role" :options="roleOptions" @change="applyRoleDefaultPermissions" />
        </el-form-item>
        <el-form-item v-if="editingUser" label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio-button label="pending">待激活</el-radio-button>
            <el-radio-button label="active">启用</el-radio-button>
            <el-radio-button label="disabled">停用</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="页面权限" prop="permissions">
          <el-checkbox-group v-model="form.permissions" class="permission-grid">
            <el-checkbox
              v-for="item in pagePermissions"
              :key="item.value"
              :label="item.value"
              :disabled="item.value === 'users' && form.role !== 'admin'"
            >
              {{ item.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remarks" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { inviteBackendUser, queryBackendUsers, updateBackendUser, updateBackendUserStatus } from '@/api'
import type { BackendUserSession } from '@/api'

type BackendRole = 'operation' | 'customer_service' | 'admin'
type BackendStatus = 'pending' | 'active' | 'disabled'

const pagePermissions = [
  { label: '数据总览', value: 'dashboard' },
  { label: '主页管理', value: 'home-manage' },
  { label: '商品管理', value: 'product' },
  { label: '库存管理', value: 'inventory' },
  { label: '订单管理', value: 'order' },
  { label: '售后中心', value: 'after-sales' },
  { label: '评价管理', value: 'reviews' },
  { label: '会员管理', value: 'member' },
  { label: '报表管理', value: 'report' },
  { label: '下载中心', value: 'download-center' },
  { label: '账号管理', value: 'users' }
]

const roleOptions = [
  { label: '运营', value: 'operation' },
  { label: '客服', value: 'customer_service' },
  { label: '管理员', value: 'admin' }
]

const roleMeta: Record<string, { label: string; type: 'success' | 'warning' | 'info' | 'danger' }> = {
  operation: { label: '运营', type: 'info' },
  customer_service: { label: '客服', type: 'warning' },
  admin: { label: '管理员', type: 'danger' }
}

const statusMeta: Record<string, { label: string; type: 'success' | 'warning' | 'info' | 'danger' }> = {
  pending: { label: '待激活', type: 'warning' },
  active: { label: '启用', type: 'success' },
  disabled: { label: '停用', type: 'danger' }
}

const defaultPermissions: Record<BackendRole, string[]> = {
  operation: ['dashboard', 'home-manage', 'product', 'inventory', 'order', 'after-sales', 'reviews', 'member', 'report', 'download-center'],
  customer_service: ['dashboard', 'order', 'after-sales', 'reviews', 'member'],
  admin: pagePermissions.map((item) => item.value)
}

const loading = ref(false)
const saving = ref(false)
const users = ref<BackendUserSession[]>([])
const total = ref(0)
const dialogVisible = ref(false)
const editingUser = ref<BackendUserSession | null>(null)
const formRef = ref<FormInstance>()

const filters = reactive({
  mobile: '',
  role: '',
  status: '',
  page: 1,
  page_size: 20
})

const form = reactive({
  nickname: '',
  mobile: '',
  role: 'operation' as BackendRole,
  status: 'pending' as BackendStatus,
  permissions: [...defaultPermissions.operation],
  remarks: ''
})

const rules: FormRules = {
  nickname: [{ required: true, message: '请输入账户名', trigger: 'blur' }],
  mobile: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
  ],
  role: [{ required: true, message: '请选择身份', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
  permissions: [{ type: 'array', required: true, min: 1, message: '请至少选择一个页面', trigger: 'change' }]
}

const normalizedFormPermissions = computed(() => {
  if (form.role !== 'admin') {
    return form.permissions.filter((permission) => permission !== 'users')
  }
  return form.permissions
})

const permissionLabel = (value: string) => {
  return pagePermissions.find((item) => item.value === value)?.label || value
}

const applyRoleDefaultPermissions = () => {
  form.permissions = [...defaultPermissions[form.role]]
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

const resetForm = () => {
  form.nickname = ''
  form.mobile = ''
  form.role = 'operation'
  form.status = 'pending'
  form.permissions = [...defaultPermissions.operation]
  form.remarks = ''
}

const openCreateDialog = () => {
  editingUser.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row: BackendUserSession) => {
  editingUser.value = row
  form.nickname = row.nickname
  form.mobile = row.mobile
  form.role = row.role
  form.status = row.status
  form.permissions = row.permissions?.length ? [...row.permissions] : [...defaultPermissions[row.role]]
  form.remarks = row.remarks || ''
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const payload = {
      nickname: form.nickname,
      role: form.role,
      permissions: normalizedFormPermissions.value,
      remarks: form.remarks
    }
    if (editingUser.value) {
      await updateBackendUser({ id: editingUser.value.id, status: form.status, ...payload })
      ElMessage.success('账号已更新')
    } else {
      await inviteBackendUser({ mobile: form.mobile, ...payload })
      ElMessage.success('账号已添加，等待首次注册激活')
    }
    dialogVisible.value = false
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '保存失败')
  } finally {
    saving.value = false
  }
}

const quickDisable = async (row: BackendUserSession) => {
  try {
    await updateBackendUserStatus({ id: row.id, status: 'disabled' })
    ElMessage.success('账号已停用')
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '停用失败')
  }
}

const quickEnable = async (row: BackendUserSession) => {
  try {
    await updateBackendUserStatus({ id: row.id, status: 'active' })
    ElMessage.success('账号已启用')
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '启用失败')
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

.permission-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.permission-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(130px, 1fr));
  gap: 4px 12px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  padding: 12px 0;
}

@media (max-width: 640px) {
  .permission-grid {
    grid-template-columns: 1fr;
  }
}
</style>
