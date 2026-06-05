<template>
  <div class="download-center-page">
    <section class="panel">
      <div class="toolbar">
        <el-select v-model="filters.business_type" placeholder="业务类型" clearable class="field" @change="loadTasks">
          <el-option v-for="item in businessOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.status" placeholder="任务状态" clearable class="field" @change="loadTasks">
          <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.template_code" placeholder="下载模板" clearable class="template-field" @change="loadTasks">
          <el-option v-for="item in templates" :key="item.template_code" :label="item.template_name" :value="item.template_code" />
        </el-select>
        <el-button type="primary" :icon="Refresh" :loading="loading" @click="loadTasks">刷新</el-button>
      </div>
    </section>

    <section class="panel">
      <el-table :data="tasks" border stripe v-loading="loading" empty-text="暂无下载任务">
        <el-table-column prop="task_name" label="任务名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="business_type" label="业务类型" width="110">
          <template #default="{ row }">{{ businessLabel(row.business_type) }}</template>
        </el-table-column>
        <el-table-column prop="template_code" label="模板编码" min-width="150" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" effect="plain">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress" label="进度" width="150">
          <template #default="{ row }">
            <el-progress :percentage="Number(row.progress || 0)" :stroke-width="8" />
          </template>
        </el-table-column>
        <el-table-column prop="row_count" label="行数" width="90" align="right" />
        <el-table-column prop="file_size" label="文件大小" width="110" align="right">
          <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column prop="download_count" label="下载次数" width="100" align="right" />
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="finished_at" label="完成时间" width="170">
          <template #default="{ row }">{{ formatTime(row.finished_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button :icon="Download" size="small" type="primary" plain :disabled="row.status !== 'success'" @click="handleDownload(row)">下载</el-button>
            <el-button :icon="RefreshLeft" size="small" plain :disabled="row.status !== 'failed'" @click="handleRetry(row)">重试</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next"
          @size-change="loadTasks"
          @current-change="loadTasks"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, Refresh, RefreshLeft } from '@element-plus/icons-vue'
import {
  downloadTaskFile,
  queryDownloadTasks,
  queryDownloadTemplates,
  retryDownloadTask,
  type DownloadTaskItem,
  type DownloadTemplateItem
} from '@/api'

const loading = ref(false)
const tasks = ref<DownloadTaskItem[]>([])
const templates = ref<DownloadTemplateItem[]>([])

const filters = reactive({
  business_type: '',
  status: '',
  template_code: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const businessOptions = [
  { label: '订单', value: 'order' },
  { label: '商品', value: 'product' },
  { label: '报表', value: 'report' },
  { label: '库存', value: 'inventory' },
  { label: '售后', value: 'after_sale' }
]

const statusOptions = [
  { label: '待生成', value: 'pending' },
  { label: '生成中', value: 'running' },
  { label: '已完成', value: 'success' },
  { label: '生成失败', value: 'failed' },
  { label: '已过期', value: 'expired' }
]

const businessLabel = (value: string) => businessOptions.find((item) => item.value === value)?.label || value
const statusLabel = (value: string) => statusOptions.find((item) => item.value === value)?.label || value
const statusType = (value: string) => {
  if (value === 'success') return 'success'
  if (value === 'failed') return 'danger'
  if (value === 'running') return 'warning'
  if (value === 'expired') return 'info'
  return ''
}

const formatTime = (value?: string) => {
  if (!value) return '-'
  const normalized = value.replace('T', ' ').replace(/\.\d+Z?$/, '')
  return normalized.slice(0, 19)
}

const formatFileSize = (value?: number) => {
  const size = Number(value || 0)
  if (!size) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

const loadTemplates = async () => {
  try {
    const res = await queryDownloadTemplates()
    templates.value = res.data?.list || []
  } catch (error) {
    console.error('query download templates failed:', error)
    ElMessage.error('下载模板查询失败')
  }
}

const loadTasks = async () => {
  loading.value = true
  try {
    const res = await queryDownloadTasks({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      business_type: filters.business_type || undefined,
      template_code: filters.template_code || undefined
    })
    tasks.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('query download tasks failed:', error)
    ElMessage.error('下载任务查询失败')
  } finally {
    loading.value = false
  }
}

const handleRetry = async (task: DownloadTaskItem) => {
  try {
    await retryDownloadTask(task.task_id)
    ElMessage.success('已重新加入生成队列')
    loadTasks()
  } catch (error) {
    console.error('retry download task failed:', error)
    ElMessage.error('重试失败')
  }
}

const handleDownload = async (task: DownloadTaskItem) => {
  try {
    const blob = await downloadTaskFile(task.task_id)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = task.file_name || `${task.task_name}.xlsx`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success('开始下载')
    loadTasks()
  } catch (error) {
    console.error('download task file failed:', error)
    ElMessage.error('文件下载失败')
  }
}

onMounted(async () => {
  await loadTemplates()
  await loadTasks()
})
</script>

<style scoped>
.download-center-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
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
  width: 140px;
}

.template-field {
  width: 220px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
