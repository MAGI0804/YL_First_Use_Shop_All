<template>
  <div class="home-manage">
    <div class="content-wrapper">
      <div class="left-panel">
        <div class="panel-header">
          <span>首页展示图</span>
          <div class="header-actions">
            <el-button v-if="hasOrderChanged" type="warning" size="small" @click="cancelSort">取消</el-button>
            <el-button v-if="hasOrderChanged" type="success" size="small" @click="saveOrder">保存顺序</el-button>
          </div>
        </div>
        <input
          ref="fileInput"
          type="file"
          accept="image/*"
          style="display: none"
          @change="handleFileChange"
        />
        <div class="top-images">
          <div class="top-images-title">小程序首页展示（拖拽调整顺序）</div>
          <div class="top-images-list">
            <div
              v-for="(item, index) in displayOnlineImages"
              :key="item.id"
              class="image-item"
              :class="{ active: selectedIndex === index, dragging: dragIndex === index }"
              draggable="true"
              @dragstart="handleDragStart($event, index)"
              @dragover.prevent="handleDragOver($event, index)"
              @dragend="handleDragEnd"
              @drop="handleDrop($event, index)"
              @click="goToDetail(item)"
            >
              <img :src="item.previewUrl" alt="image" />
              <div class="image-order">Order: {{ hasOrderChanged ? index + 1 : item.order }}</div>
              <div class="drag-handle">
                <el-icon><Rank /></el-icon>
              </div>
            </div>
            <div v-if="displayOnlineImages.length === 0" class="empty-text">暂无上线图片</div>
          </div>
        </div>

        <div class="bottom-section">
          <div class="section-header">
            <div class="section-title">所有图片列表</div>
            <el-button type="primary" size="small" @click="addImage">
              <el-icon><Plus /></el-icon>
              添加图片
            </el-button>
          </div>
          <div class="filter-bar">
            <el-select v-model="filterParams.status" placeholder="全部状态" clearable size="small" style="width: 120px;">
              <el-option label="上线" value="online" />
              <el-option label="下线" value="offline" />
              <el-option label="待处理" value="pending" />
            </el-select>
            <el-select v-model="filterParams.has_activity_detail" placeholder="全部跳转设置" clearable size="small" style="width: 140px;">
              <el-option label="可跳转" :value="true" />
              <el-option label="不可跳转" :value="false" />
            </el-select>
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              size="small"
              style="width: 380px;"
              value-format="YYYY-MM-DD HH:mm:ss"
            />
            <el-button type="primary" size="small" @click="handleSearch">搜索</el-button>
            <el-button size="small" @click="handleReset">重置</el-button>
          </div>
          <div v-if="loading" class="loading-container">
            <el-loading-spinner style="width: 32px; height: 32px;" />
            <span style="margin-left: 10px;">加载中...</span>
          </div>
          <div v-else class="image-list">
            <div
              v-for="(item, index) in paginatedImages"
              :key="item.id"
              class="image-item"
              :class="{ active: selectedIndex === index }"
            >
              <img :src="item.previewUrl" alt="image" />
              <div :class="['detail-badge', item.has_activity_detail ? 'can-jump' : 'cannot-jump']">
                {{ item.has_activity_detail ? '可跳转' : '不可跳转' }}
              </div>
              <div class="image-status" :class="item.status">
                {{ getStatusText(item.status) }}
              </div>
              <div v-if="item.order !== null && item.order !== undefined" class="image-order">
                顺序: {{ item.order }}
              </div>
              <div class="image-actions">
                <el-button v-if="item.status === 'offline' || item.status === 'pending'" type="success" link size="small" @click.stop="onlineImage(index)">
                  上线
                </el-button>
                <el-button v-if="item.status === 'online'" type="warning" link size="small" @click.stop="offlineImage(index)">
                  下线
                </el-button>
                <el-button type="primary" link size="small" @click.stop="goToDetail(item)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button type="danger" link size="small" @click.stop="deleteImage(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
          <div class="pagination">
            <div class="pagination-info">
              共 {{ filteredImages.length }} 条，第 {{ currentPage }} / {{ totalPages }} 页
            </div>
            <el-pagination
              v-model:current-page="currentPage"
              :page-size="pageSize"
              :total="filteredImages.length"
              layout="prev, pager, next"
              small
            />
          </div>
        </div>
      </div>

      <div class="right-panel">
        <div class="phone-wrapper">
          <div class="phone-14pm">
            <div class="phone-notch"></div>
            <div class="phone-content">
              <div class="search-container">
                <div class="search-box">
                  <el-icon class="search-icon"><Search /></el-icon>
                  <input type="text" placeholder="搜索商品" class="search-input" />
                  <div class="search-btn">搜索</div>
                </div>
              </div>
              <div class="scrollarea">
                <div class="container">
                  <div class="image-list">
                    <div v-for="(item, index) in displayOnlineImages" :key="item.id" class="image-item">
                      <img :src="item.previewUrl" alt="banner" class="item-image" />
                    </div>
                  </div>
                </div>
              </div>
              <div class="tab-bar">
                <div class="tab-item" v-for="(tab, index) in tabList" :key="index">
                  <div class="tab-icon" :class="{ active: index === 0 }"></div>
                  <div class="tab-text" :class="{ active: index === 0 }">{{ tab.text }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <el-dialog v-model="editDialogVisible" title="编辑图片信息" width="500px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="分类">
          <el-select v-model="editForm.category" placeholder="请选择分类" clearable style="width: 100%;">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品款号">
          <el-input 
            v-model="styleCodesInput" 
            type="textarea" 
            :rows="3"
            placeholder="多个款号用换行或逗号分隔"
          />
          <div style="margin-top: 8px; color: #999; font-size: 12px;">
            已输入: {{ styleCodesList.length }} 个款号
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEdit" :loading="savingEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Search, Rank, Link } from '@element-plus/icons-vue'
import { 
  batchQueryActivityImages, 
  updateActivityImageRelations,
  batchUpdateActivityImageOrder,
  getAllCategories,
  type ActivityImageItem,
  addActivityImg,
  activityImageOnline,
  activityImageOffline
} from '@/api'

const router = useRouter()

const tabList = [
  { text: '首页' },
  { text: '童装' },
  { text: '成人装' },
  { text: '感兴趣' },
  { text: '我的' }
]

const fileInput = ref<HTMLInputElement>()
const selectedIndex = ref(0)
const images = ref<Array<ActivityImageItem & { previewUrl?: string }>>([])
const onlineImages = ref<Array<ActivityImageItem & { previewUrl?: string }>>([])
const displayOnlineImages = ref<Array<ActivityImageItem & { previewUrl?: string }>>([])
const loading = ref(false)
const uploading = ref(false)
const saving = ref(false)
const savingEdit = ref(false)
const dragIndex = ref<number | null>(null)
const dateRange = ref<[string, string] | null>(null)
const editDialogVisible = ref(false)
const categories = ref<string[]>([])
const editingImageId = ref<number | null>(null)
const editForm = ref({
  category: '',
  style_codes: [] as string[]
})
const styleCodesInput = ref('')

const filterParams = ref({
  status: '',
  has_activity_detail: undefined as boolean | undefined,
  start_time: '',
  end_time: ''
})
const currentPage = ref(1)
const pageSize = 10

const styleCodesList = computed(() => {
  if (!styleCodesInput.value) return []
  return styleCodesInput.value
    .split(/[\n,，]/)
    .map(s => s.trim())
    .filter(s => s.length > 0)
})

const hasOrderChanged = computed(() => {
  if (displayOnlineImages.value.length !== onlineImages.value.length) return true
  return displayOnlineImages.value.some((item, index) => item.id !== onlineImages.value[index].id)
})

const topOnlineImages = computed(() => {
  return onlineImages.value
})

const filteredImages = computed(() => {
  let result = images.value
  if (filterParams.value.status) {
    result = result.filter(item => item.status === filterParams.value.status)
  }
  if (filterParams.value.has_activity_detail !== undefined && filterParams.value.has_activity_detail !== null) {
    result = result.filter(item => item.has_activity_detail === filterParams.value.has_activity_detail)
  }
  return result
})

const paginatedImages = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return filteredImages.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(filteredImages.value.length / pageSize)
})

const cleanUrl = (url: string) => {
  if (!url) return ''
  return url.replace(/[`\s]/g, '')
}

const loadCategories = async () => {
  try {
    const res = await getAllCategories({ shopname: 'youlan_kids' })
    if (res.code === 200 && res.data && res.data.categories) {
      categories.value = res.data.categories
    }
  } catch (error) {
    console.error('获取分类失败:', error)
  }
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    'online': '上线',
    'offline': '下线',
    'pending': '待处理'
  }
  return statusMap[status] || status
}

const loadOnlineImages = async () => {
  try {
    const params = {
      shopname: 'youlan_kids',
      page: 1,
      pageSize: 100
    }
    const res = await batchQueryActivityImages(params)
    if (res.code === 200 && res.data) {
      const sorted = [...res.data.items]
        .filter(item => item.status === 'online')
        .sort((a, b) => {
          if (a.order !== null && a.order !== undefined && b.order !== null && b.order !== undefined) {
            return a.order - b.order
          }
          if (a.order !== null && a.order !== undefined) return -1
          if (b.order !== null && b.order !== undefined) return 1
          return (a.id || 0) - (b.id || 0)
        })
      
      onlineImages.value = sorted.map(item => ({
        ...item,
        previewUrl: cleanUrl(item.image)
      }))
      displayOnlineImages.value = [...onlineImages.value]
    }
  } catch (error) {
    console.error('获取上线图片失败:', error)
  }
}

const loadActivityImages = async () => {
  loading.value = true
  try {
    const params: any = {
      shopname: 'youlan_kids',
      page: 1,
      pageSize: 100
    }
    if (filterParams.value.status) {
      params.status = filterParams.value.status
    }
    if (filterParams.value.has_activity_detail !== undefined && filterParams.value.has_activity_detail !== null) {
      params.has_activity_detail = filterParams.value.has_activity_detail
    }
    if (dateRange.value && dateRange.value[0]) {
      params.start_time = dateRange.value[0]
    }
    if (dateRange.value && dateRange.value[1]) {
      params.end_time = dateRange.value[1]
    }
    const res = await batchQueryActivityImages(params)
    if (res.code === 200 && res.data) {
      images.value = res.data.items.map(item => ({
        ...item,
        previewUrl: cleanUrl(item.image)
      }))
    }
    await loadOnlineImages()
  } catch (error) {
    console.error('获取活动图失败:', error)
    ElMessage.error('获取活动图失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadActivityImages()
}

const handleReset = () => {
  filterParams.value = {
    status: '',
    has_activity_detail: undefined,
    start_time: '',
    end_time: ''
  }
  dateRange.value = null
  loadActivityImages()
}

onMounted(() => {
  loadActivityImages()
  loadCategories()
})

const selectImage = (index: number) => {
  selectedIndex.value = index
}

const addImage = () => {
  fileInput.value?.click()
}

const handleFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('image', file)
    
    const res = await addActivityImg(formData)
    if (res.code === 200) {
      ElMessage.success('添加图片成功')
      loadActivityImages()
    } else {
      ElMessage.error(res.msg || '添加图片失败')
    }
  } catch (error) {
    console.error('添加图片失败:', error)
    ElMessage.error('添加图片失败')
  } finally {
    uploading.value = false
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  }
}

const onlineImage = async (index: number) => {
  const image = paginatedImages.value[index]
  try {
    const res = await activityImageOnline({ activity_id: image.id })
    if (res.code === 200) {
      ElMessage.success('上线成功')
      loadActivityImages()
    } else {
      ElMessage.error(res.msg || '上线失败')
    }
  } catch (error) {
    console.error('上线失败:', error)
    ElMessage.error('上线失败')
  }
}

const offlineImage = async (index: number) => {
  const image = paginatedImages.value[index]
  try {
    const res = await activityImageOffline({ activity_id: image.id })
    if (res.code === 200) {
      ElMessage.success('下线成功')
      loadActivityImages()
    } else {
      ElMessage.error(res.msg || '下线失败')
    }
  } catch (error) {
    console.error('下线失败:', error)
    ElMessage.error('下线失败')
  }
}

const uploadImage = (index: number) => {
  const image = paginatedImages.value[index]
  if (!image) return
  
  editingImageId.value = image.id
  editForm.value = {
    category: image.category || '',
    style_codes: Array.isArray(image.style_codes) ? image.style_codes : []
  }
  styleCodesInput.value = editForm.value.style_codes.join('\n')
  editDialogVisible.value = true
}

const saveEdit = async () => {
  if (!editingImageId.value) return
  
  savingEdit.value = true
  try {
    const params: any = {
      activity_id: editingImageId.value
    }
    
    if (editForm.value.category) {
      params.category = editForm.value.category
    }
    
    if (styleCodesList.value.length > 0) {
      params.style_codes = styleCodesList.value
    }
    
    const res = await updateActivityImageRelations(params)
    if (res.code === 200) {
      ElMessage.success('更新成功')
      editDialogVisible.value = false
      loadActivityImages()
    } else {
      ElMessage.error(res.msg || '更新失败')
    }
  } catch (error) {
    console.error('更新失败:', error)
    ElMessage.error('更新失败')
  } finally {
    savingEdit.value = false
  }
}

const goToDetail = (item: ActivityImageItem) => {
  ElMessageBox.confirm('确定要查看此活动图详情吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info'
  }).then(() => {
    router.push(`/activity/${item.id}`)
  }).catch(() => { })
}

const deleteImage = (index: number) => {
  ElMessageBox.confirm('确定要删除这张图片吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.info('删除图片功能开发中')
  }).catch(() => { })
}

const handleDragStart = (event: DragEvent, index: number) => {
  dragIndex.value = index
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

const handleDragOver = (event: DragEvent, index: number) => {
  if (dragIndex.value === null || dragIndex.value === index) return
  const item = displayOnlineImages.value[dragIndex.value]
  displayOnlineImages.value.splice(dragIndex.value, 1)
  displayOnlineImages.value.splice(index, 0, item)
  dragIndex.value = index
}

const handleDragEnd = () => {
  dragIndex.value = null
}

const handleDrop = (event: DragEvent, index: number) => {
  event.preventDefault()
}

const cancelSort = () => {
  displayOnlineImages.value = [...onlineImages.value]
}

const saveOrder = async () => {
  try {
    await ElMessageBox.confirm('确定保存新的图片顺序吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    saving.value = true
    const orderData = displayOnlineImages.value.map((item, index) => ({
      id: item.id,
      order: index + 1
    }))
    
    const res = await batchUpdateActivityImageOrder({ images: orderData })
    if (res.code === 200) {
      ElMessage.success('批量更新顺序成功')
      loadActivityImages()
    } else {
      ElMessage.error(res.msg || '批量更新顺序失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('保存顺序失败:', error)
      ElMessage.error('保存顺序失败')
    }
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.home-manage {
  min-height: calc(100vh - 120px);
  display: flex;
  background-color: #ffffff;
}

.content-wrapper {
  display: flex;
  width: 100%;
  gap: 24px;
}

.left-panel {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  overflow: visible;
  min-height: 100%;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 500;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.top-images {
  flex-shrink: 0;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.top-images-title {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
}

.top-images-list {
  display: grid;
  grid-template-columns: repeat(5, minmax(160px, 1fr));
  gap: 16px;
}

.top-images-list .image-item {
  width: 100%;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  border: 2px solid transparent;
  cursor: pointer;
  user-select: none;
  transition: transform 0.2s, box-shadow 0.2s;
}

.top-images-list .image-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.top-images-list .image-item.dragging {
  opacity: 0.5;
  transform: scale(1.05);
  cursor: grabbing;
}

.top-images-list .image-item.active {
  border-color: #409eff;
}

.top-images-list .image-item img {
  width: 100%;
  height: auto;
  object-fit: contain;
  display: block;
}

.image-order {
  position: absolute;
  bottom: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.drag-handle {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px;
  border-radius: 4px;
  font-size: 14px;
  cursor: grab;
}

.empty-text {
  color: #999;
  font-size: 14px;
  padding: 20px 0;
}

.bottom-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: visible;
  min-height: 600px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 14px;
  color: #666;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.image-list {
  display: grid;
  grid-template-columns: repeat(5, minmax(160px, 1fr));
  gap: 16px;
  overflow-y: visible;
  flex: none;
}

.left-panel .image-list .image-item {
  width: 100%;
  position: relative;
  border: 2px solid transparent;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
}

.left-panel .image-list .image-item.active {
  border-color: #409eff;
}

.left-panel .image-list .image-item img {
  width: 100%;
  height: auto;
  object-fit: contain;
  display: block;
}

.image-status {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: white;
  z-index: 1;
}

.image-status.online {
  background: #67c23a;
}

.image-status.offline {
  background: #909399;
}

.image-status.pending {
  background: #e6a23c;
}

.left-panel .image-list .image-order {
  position: absolute;
  bottom: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.image-actions {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  opacity: 0;
  transition: opacity 0.3s;
}

.left-panel .image-list .image-item:hover .image-actions {
  opacity: 1;
}

.detail-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  padding: 4px 10px;
  border-radius: 0;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  color: white;
}

.detail-badge.can-jump {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.detail-badge.cannot-jump {
  background: linear-gradient(135deg, #9e9e9e 0%, #757575 100%);
}

.right-panel {
  width: 480px;
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.phone-wrapper {
  display: flex;
  justify-content: center;
  flex: 1;
}

.phone-14pm {
  width: 393px;
  height: 852px;
  background: #1a1a1a;
  border-radius: 55px;
  padding: 12px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.3);
  position: relative;
}

.phone-notch {
  position: absolute;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  width: 150px;
  height: 37px;
  background: #1a1a1a;
  border-radius: 0 0 20px 20px;
  z-index: 10;
}

.phone-content {
  width: 100%;
  height: 100%;
  background-color: #e6f2ff;
  border-radius: 45px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 搜索框样式 - 与小程序一致 */
.search-container {
  padding: 10px;
  padding-top: 50px;
  background-color: #fff;
  border-bottom: 0.5px solid #eee;
  position: relative;
  z-index: 100;
  flex-shrink: 0;
}

.search-box {
  display: flex;
  align-items: center;
  background-color: #fff;
  border-radius: 20px;
  height: 35px;
  padding: 0 10px;
  border: 0.5px solid #d1e9ff;
}

.search-icon {
  color: #999;
  margin-right: 5px;
  font-size: 14px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 14px;
  background: transparent;
}

.search-btn {
  font-size: 14px;
  color: #1989fa;
  margin-left: 10px;
  cursor: pointer;
}

/* 滚动区域样式 - 与小程序一致 */
.scrollarea {
  margin-top: 0;
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch;
}

.container {
  padding: 0;
}

.scrollarea .image-list {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: auto;
  grid-template-columns: none;
  gap: 0;
}

.scrollarea .image-item {
  width: 100%;
  height: auto;
  margin-bottom: 0;
  overflow: hidden;
  border-radius: 0;
  flex-shrink: 0;
}

.item-image {
  width: 100%;
  height: auto;
  display: block;
  object-fit: contain;
}

/* TabBar样式 - 与小程序app.json一致 */
.tab-bar {
  height: 50px;
  background-color: #fff;
  border-top: 0.5px solid #eee;
  flex-shrink: 0;
  display: flex;
  flex-direction: row;
  padding-bottom: 20px;
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 5px;
}

.tab-icon {
  width: 24px;
  height: 24px;
  background-color: #ccc;
  margin-bottom: 2px;
}

.tab-icon.active {
  background-color: #121212;
}

.tab-text {
  font-size: 10px;
  color: #121212;
}

.tab-text.active {
  color: #121212;
}

.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.pagination-info {
  font-size: 14px;
  color: #666;
}
</style>
