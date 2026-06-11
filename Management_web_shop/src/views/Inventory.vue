<template>
  <div class="inventory-page">
    <el-tabs v-model="activeTab" class="inventory-tabs">
      <el-tab-pane label="库存查询" name="query">
        <section class="panel">
          <div class="toolbar">
            <el-input
              v-model="queryForm.commodity_id"
              placeholder="SKU"
              clearable
              class="field"
              @keyup.enter="handleInventoryQuery"
            />
            <el-input
              v-model="queryForm.style_code"
              placeholder="款号"
              clearable
              class="field"
              @keyup.enter="handleInventoryQuery"
            />
            <el-button type="primary" :icon="Search" :loading="queryLoading" @click="handleInventoryQuery">查询</el-button>
            <el-button :icon="Refresh" @click="resetInventoryQuery">重置</el-button>
            <el-button :icon="Download" @click="handleInventoryExport('query')">导出</el-button>
            <span v-if="queryTotalInventory !== null" class="summary">总库存 {{ queryTotalInventory }}</span>
          </div>

          <el-table :data="queryRows" border height="560" empty-text="暂无库存数据">
            <el-table-column prop="commodity_id" label="SKU" min-width="150" />
            <el-table-column prop="name" label="商品名称" min-width="220" />
            <el-table-column prop="style_code" label="款号" width="140" />
            <el-table-column prop="category" label="分类" width="120" />
            <el-table-column prop="color" label="颜色" width="100" />
            <el-table-column prop="size" label="尺码" width="100" />
            <el-table-column prop="inventory" label="库存" width="100" align="right" />
            <el-table-column prop="price" label="价格" width="100" align="right">
              <template #default="{ row }">¥{{ formatMoney(row.price) }}</template>
            </el-table-column>
          </el-table>

          <div v-if="openInventoryRows.length" class="sub-section">
            <div class="sub-title">
              <span>开放库存余额</span>
              <span class="summary">
                实物 {{ openInventorySummary?.total_on_hand_qty ?? 0 }} /
                锁定 {{ openInventorySummary?.total_locked_qty ?? 0 }} /
                可用 {{ openInventorySummary?.total_available_qty ?? 0 }}
              </span>
            </div>
            <el-table :data="openInventoryRows" border height="260" empty-text="暂无开放库存余额">
              <el-table-column prop="commodity_id" label="SKU" min-width="150" />
              <el-table-column prop="style_code" label="款号" width="130" />
              <el-table-column prop="warehouse_code" label="仓库" width="120" />
              <el-table-column prop="on_hand_qty" label="实物库存" width="110" align="right" />
              <el-table-column prop="locked_qty" label="锁定库存" width="110" align="right" />
              <el-table-column prop="available_qty" label="可用库存" width="110" align="right" />
              <el-table-column prop="version" label="版本" width="90" align="right" />
              <el-table-column prop="updated_at" label="更新时间" width="180" />
            </el-table>
          </div>
        </section>
      </el-tab-pane>

      <el-tab-pane label="库存预警" name="warnings">
        <section class="panel">
          <div class="toolbar">
            <el-input-number v-model="warningParams.threshold" :min="1" :max="9999" controls-position="right" />
            <el-button type="primary" :icon="Search" :loading="warningLoading" @click="loadWarnings">查询</el-button>
            <el-button :icon="Refresh" @click="resetWarnings">重置</el-button>
            <el-button :icon="Download" @click="handleInventoryExport('warnings')">导出</el-button>
            <span class="summary">阈值 {{ warningThreshold }}</span>
          </div>

          <el-table :data="warningRows" border height="520" empty-text="暂无预警商品">
            <el-table-column prop="commodity_id" label="SKU" min-width="150" />
            <el-table-column prop="name" label="商品名称" min-width="220" />
            <el-table-column prop="style_code" label="款号" width="140" />
            <el-table-column prop="category" label="分类" width="120" />
            <el-table-column prop="inventory" label="当前库存" width="120" align="right">
              <template #default="{ row }">
                <el-tag :type="row.inventory <= warningThreshold ? 'danger' : 'warning'" size="small">
                  {{ row.inventory }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="warningParams.page"
              v-model:page-size="warningParams.page_size"
              :page-sizes="[10, 20, 50, 100]"
              :total="warningTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadWarnings"
              @current-change="loadWarnings"
            />
          </div>
        </section>
      </el-tab-pane>

      <el-tab-pane label="库存日志" name="logs">
        <section class="panel">
          <div class="toolbar">
            <el-input v-model="logParams.commodity_id" placeholder="SKU" clearable class="field" />
            <el-input v-model="logParams.style_code" placeholder="款号" clearable class="field" />
            <el-select v-model="logParams.change_type" placeholder="类型" clearable class="field">
              <el-option label="下单扣减" value="order_create_deduct" />
              <el-option label="取消回滚" value="order_cancel_restore" />
              <el-option label="售后回滚" value="return_completed_restore" />
              <el-option label="手动调整" value="manual_adjust" />
              <el-option label="聚水潭同步" value="jushuitan_sync" />
              <el-option label="库存调拨" value="stock_transfer" />
              <el-option label="库存盘点" value="stock_check" />
            </el-select>
            <el-button type="primary" :icon="Search" :loading="logLoading" @click="loadLogs">查询</el-button>
            <el-button :icon="Refresh" @click="resetLogs">重置</el-button>
            <el-button :icon="Download" @click="handleInventoryExport('logs')">导出</el-button>
          </div>

          <el-table :data="logRows" border height="520" empty-text="暂无库存日志">
            <el-table-column prop="created_at" label="时间" width="180" />
            <el-table-column prop="source" label="来源" width="90">
              <template #default="{ row }">
                <el-tag :type="row.source === 'open' ? 'success' : 'info'" size="small">
                  {{ logSourceText(row.source) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="movement_no" label="开放流水号" min-width="190" show-overflow-tooltip />
            <el-table-column prop="commodity_id" label="SKU" min-width="150" />
            <el-table-column prop="style_code" label="款号" width="130" />
            <el-table-column prop="warehouse_code" label="仓库" width="120" />
            <el-table-column prop="change_type" label="类型" width="140">
              <template #default="{ row }">{{ changeTypeText(row.change_type) }}</template>
            </el-table-column>
            <el-table-column prop="before_qty" label="前库存" width="90" align="right" />
            <el-table-column prop="change_qty" label="变动" width="90" align="right" />
            <el-table-column prop="after_qty" label="后库存" width="90" align="right" />
            <el-table-column prop="related_order_id" label="订单" width="150" />
            <el-table-column label="业务号" width="160">
              <template #default="{ row }">{{ inventoryBizNo(row) }}</template>
            </el-table-column>
            <el-table-column prop="operator_id" label="操作人" width="110" />
            <el-table-column prop="remark" label="备注" min-width="220" show-overflow-tooltip />
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="logParams.page"
              v-model:page-size="logParams.page_size"
              :page-sizes="[10, 20, 50, 100]"
              :total="logTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadLogs"
              @current-change="loadLogs"
            />
          </div>
        </section>
      </el-tab-pane>

      <el-tab-pane label="库存操作" name="operations">
        <section class="operation-grid">
          <div class="panel operation-panel">
            <div class="panel-title">
              <el-icon><EditPen /></el-icon>
              <span>手动调整</span>
            </div>
            <el-form :model="adjustForm" label-width="88px">
              <el-form-item label="SKU">
                <el-input v-model="adjustForm.commodity_id" />
              </el-form-item>
              <el-form-item label="变动数量">
                <el-input-number v-model="adjustForm.change_qty" :min="-99999" :max="99999" controls-position="right" />
              </el-form-item>
              <el-form-item label="仓库">
                <el-input v-model="adjustForm.warehouse_code" />
              </el-form-item>
              <el-form-item label="操作人">
                <el-input v-model="adjustForm.operator_id" />
              </el-form-item>
              <el-form-item label="备注">
                <el-input v-model="adjustForm.remark" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="EditPen" :loading="operationLoading" @click="submitAdjust">提交调整</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div class="panel operation-panel">
            <div class="panel-title">
              <el-icon><Switch /></el-icon>
              <span>库存调拨</span>
            </div>
            <el-form :model="transferForm" label-width="88px">
              <el-form-item label="SKU">
                <el-input v-model="transferForm.commodity_id" />
              </el-form-item>
              <el-form-item label="数量">
                <el-input-number v-model="transferForm.qty" :min="1" :max="99999" controls-position="right" />
              </el-form-item>
              <el-form-item label="源仓库">
                <el-input v-model="transferForm.source_warehouse_code" />
              </el-form-item>
              <el-form-item label="目标仓库">
                <el-input v-model="transferForm.target_warehouse_code" />
              </el-form-item>
              <el-form-item label="操作人">
                <el-input v-model="transferForm.operator_id" />
              </el-form-item>
              <el-form-item label="备注">
                <el-input v-model="transferForm.remark" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="Switch" :loading="operationLoading" @click="submitTransfer">提交调拨</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div class="panel operation-panel">
            <div class="panel-title">
              <el-icon><Tickets /></el-icon>
              <span>库存盘点</span>
            </div>
            <el-form :model="stockCheckForm" label-width="88px">
              <el-form-item label="SKU">
                <el-input v-model="stockCheckForm.commodity_id" />
              </el-form-item>
              <el-form-item label="实盘数量">
                <el-input-number v-model="stockCheckForm.actual_qty" :min="0" :max="99999" controls-position="right" />
              </el-form-item>
              <el-form-item label="仓库">
                <el-input v-model="stockCheckForm.warehouse_code" />
              </el-form-item>
              <el-form-item label="操作人">
                <el-input v-model="stockCheckForm.operator_id" />
              </el-form-item>
              <el-form-item label="备注">
                <el-input v-model="stockCheckForm.remark" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="Tickets" :loading="operationLoading" @click="submitStockCheck">提交盘点</el-button>
              </el-form-item>
            </el-form>
          </div>
        </section>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, EditPen, Refresh, Search, Switch, Tickets } from '@element-plus/icons-vue'
import {
  adjustInventory,
  queryInventory,
  queryInventoryLogs,
  queryInventoryWarnings,
  stockCheckInventory,
  transferInventory,
  createDownloadTask,
  type InventoryCommodity,
  type InventoryLogItem,
  type OpenInventoryBalanceItem,
  type OpenInventorySummary
} from '@/api'

const activeTab = ref('query')

const queryLoading = ref(false)
const warningLoading = ref(false)
const logLoading = ref(false)
const operationLoading = ref(false)

const queryForm = reactive({
  commodity_id: '',
  style_code: ''
})
const queryRows = ref<InventoryCommodity[]>([])
const queryTotalInventory = ref<number | null>(null)
const openInventoryRows = ref<OpenInventoryBalanceItem[]>([])
const openInventorySummary = ref<OpenInventorySummary | null>(null)

const warningParams = reactive({
  threshold: 5,
  page: 1,
  page_size: 10
})
const warningRows = ref<InventoryCommodity[]>([])
const warningTotal = ref(0)
const warningThreshold = ref(5)

const logParams = reactive({
  commodity_id: '',
  style_code: '',
  change_type: '',
  page: 1,
  page_size: 10
})
const logRows = ref<InventoryLogItem[]>([])
const logTotal = ref(0)

const adjustForm = reactive({
  commodity_id: '',
  change_qty: 0,
  warehouse_code: 'default',
  operator_id: '',
  remark: ''
})

const transferForm = reactive({
  commodity_id: '',
  qty: 1,
  source_warehouse_code: 'default',
  target_warehouse_code: '',
  operator_id: '',
  remark: ''
})

const stockCheckForm = reactive({
  commodity_id: '',
  actual_qty: 0,
  warehouse_code: 'default',
  operator_id: '',
  remark: ''
})

const formatMoney = (value: number) => Number(value || 0).toFixed(2)

const logSourceText = (source?: string) => {
  if (source === 'open') return '开放'
  if (source === 'legacy') return '旧'
  return source || '-'
}

const inventoryBizNo = (row: InventoryLogItem) => {
  return row.biz_id || row.related_sub_order_id || row.related_return_id || row.related_order_id || '-'
}

const changeTypeText = (type: string) => {
  const map: Record<string, string> = {
    order_create_deduct: '下单扣减',
    order_deduct: '下单扣减',
    order_cancel_restore: '取消回滚',
    return_completed_restore: '售后回滚',
    return_restore: '售后回滚',
    manual_adjust: '手动调整',
    jushuitan_sync: '聚水潭同步',
    sync_jushuitan: '聚水潭同步',
    stock_transfer: '库存调拨',
    stock_check: '库存盘点'
  }
  return map[type] || type || '-'
}

const handleInventoryQuery = async () => {
  if (!queryForm.commodity_id && !queryForm.style_code) {
    ElMessage.warning('请填写 SKU 或款号')
    return
  }
  queryLoading.value = true
  try {
    const res = await queryInventory({
      commodity_id: queryForm.commodity_id,
      style_code: queryForm.style_code
    })
    const data = res.data || {}
    if (data.commodity) {
      queryRows.value = [data.commodity]
      queryTotalInventory.value = data.commodity.inventory
    } else {
      queryRows.value = data.commodities || []
      queryTotalInventory.value = data.total_inventory ?? null
    }
    openInventoryRows.value = data.open_inventory?.items || []
    openInventorySummary.value = data.open_inventory?.summary || null
  } catch (error) {
    console.error('query inventory failed:', error)
    ElMessage.error('库存查询失败')
  } finally {
    queryLoading.value = false
  }
}

const resetInventoryQuery = () => {
  queryForm.commodity_id = ''
  queryForm.style_code = ''
  queryRows.value = []
  queryTotalInventory.value = null
  openInventoryRows.value = []
  openInventorySummary.value = null
}

const loadWarnings = async () => {
  warningLoading.value = true
  try {
    const res = await queryInventoryWarnings(warningParams)
    warningRows.value = res.data?.data || []
    warningTotal.value = res.data?.total || 0
    warningThreshold.value = res.data?.threshold || warningParams.threshold
  } catch (error) {
    console.error('query inventory warnings failed:', error)
    ElMessage.error('库存预警查询失败')
  } finally {
    warningLoading.value = false
  }
}

const resetWarnings = () => {
  warningParams.threshold = 5
  warningParams.page = 1
  warningParams.page_size = 10
  loadWarnings()
}

const loadLogs = async () => {
  logLoading.value = true
  try {
    const res = await queryInventoryLogs(logParams)
    logRows.value = res.data?.data || []
    logTotal.value = res.data?.total || 0
  } catch (error) {
    console.error('query inventory logs failed:', error)
    ElMessage.error('库存日志查询失败')
  } finally {
    logLoading.value = false
  }
}

const resetLogs = () => {
  logParams.commodity_id = ''
  logParams.style_code = ''
  logParams.change_type = ''
  logParams.page = 1
  logParams.page_size = 10
  loadLogs()
}

const handleInventoryExport = async (scope: 'query' | 'warnings' | 'logs') => {
  const filters: Record<string, any> = {}
  if (scope === 'query') {
    filters.commodity_id = queryForm.commodity_id || undefined
    filters.style_code = queryForm.style_code || undefined
  }
  if (scope === 'warnings') {
    filters.low_inventory_threshold = warningParams.threshold
  }
  if (scope === 'logs') {
    filters.commodity_id = logParams.commodity_id || undefined
    filters.style_code = logParams.style_code || undefined
  }
  try {
    await createDownloadTask({
      template_code: 'inventory_export',
      file_format: 'xlsx',
      filters
    })
    ElMessage.success('库存下载任务已创建，请到下载中心查看')
  } catch (error) {
    console.error('create inventory download task failed:', error)
    ElMessage.error('库存下载任务创建失败')
  }
}

const ensureCommodity = (commodityID: string) => {
  if (!commodityID.trim()) {
    ElMessage.warning('请填写 SKU')
    return false
  }
  return true
}

const afterOperation = async (message: string) => {
  ElMessage.success(message)
  await Promise.all([loadLogs(), loadWarnings()])
}

const submitAdjust = async () => {
  if (!ensureCommodity(adjustForm.commodity_id) || adjustForm.change_qty === 0) {
    if (adjustForm.change_qty === 0) ElMessage.warning('变动数量不能为 0')
    return
  }
  operationLoading.value = true
  try {
    await adjustInventory(adjustForm)
    await afterOperation('库存调整成功')
  } catch (error) {
    console.error('adjust inventory failed:', error)
    ElMessage.error('库存调整失败')
  } finally {
    operationLoading.value = false
  }
}

const submitTransfer = async () => {
  if (!ensureCommodity(transferForm.commodity_id)) return
  operationLoading.value = true
  try {
    await transferInventory(transferForm)
    await afterOperation('库存调拨成功')
  } catch (error) {
    console.error('transfer inventory failed:', error)
    ElMessage.error('库存调拨失败')
  } finally {
    operationLoading.value = false
  }
}

const submitStockCheck = async () => {
  if (!ensureCommodity(stockCheckForm.commodity_id)) return
  operationLoading.value = true
  try {
    await stockCheckInventory(stockCheckForm)
    await afterOperation('库存盘点成功')
  } catch (error) {
    console.error('stock check failed:', error)
    ElMessage.error('库存盘点失败')
  } finally {
    operationLoading.value = false
  }
}

onMounted(() => {
  loadWarnings()
  loadLogs()
})
</script>

<style scoped>
.inventory-page {
  padding: 20px;
}

.inventory-tabs {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 0 16px 16px;
}

.panel {
  background: #ffffff;
}

.toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding: 12px 0 16px;
}

.field {
  width: 180px;
}

.summary {
  color: #606266;
  font-size: 14px;
}

.sub-section {
  margin-top: 16px;
}

.sub-title {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #303133;
  font-weight: 600;
  margin-bottom: 10px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

.operation-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(260px, 1fr));
  gap: 16px;
  padding-top: 12px;
}

.operation-panel {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 16px;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #303133;
  font-weight: 600;
  margin-bottom: 16px;
}

@media (max-width: 1180px) {
  .operation-grid {
    grid-template-columns: 1fr;
  }
}
</style>
