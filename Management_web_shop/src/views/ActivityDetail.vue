<template>
  <div class="activity-detail">
    <div class="page-header">
      <el-button link @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <span class="page-title">活动图详情</span>
    </div>

    <div v-if="loading" class="loading-container">
      <el-loading-spinner style="width: 32px; height: 32px;" />
      <span style="margin-left: 10px;">加载中...</span>
    </div>

    <div v-else-if="activityData" class="detail-content">
      <div class="left-panel">
        <el-card class="main-image-card">
          <template #header>
            <div class="card-header">
              <span>活动主图</span>
              <el-switch
                v-model="hasDetailSwitch"
                active-text="可跳转"
                inactive-text="不可跳转"
                @change="handleSetDetail"
              />
            </div>
          </template>
          <div class="main-image-wrapper">
            <img :src="cleanUrl(activityData.image)" class="main-image" alt="活动主图" />
          </div>
        </el-card>

        <el-card class="promotional-pics-card">
          <template #header>
            <div class="card-header">
              <span>宣传图</span>
              <div class="header-actions">
                <el-button v-if="isPromoOrderChanged" type="warning" size="small" @click="cancelPromoOrder">
                  取消
                </el-button>
                <el-button v-if="isPromoOrderChanged" type="success" size="small" @click="savePromoOrder">
                  确认调整
                </el-button>
                <el-button type="primary" size="small" @click="triggerPromoUpload">
                  <el-icon><Plus /></el-icon>
                  上传图片
                </el-button>
              </div>
            </div>
          </template>
          <input
            ref="promoFileInput"
            type="file"
            accept="image/*"
            style="display: none"
            @change="handlePromoUpload"
          />
          <div v-if="currentPromoPics.length > 0" class="promotional-pics-grid">
            <div
              v-for="(pic, index) in currentPromoPics"
              :key="(pic.order || index) + '-' + (isPromoOrderChanged ? 'temp' : 'original')"
              class="promotional-pic-item"
              :class="{ dragging: promoDragIndex === index }"
              draggable="true"
              @dragstart="handlePromoDragStart($event, index)"
              @dragover.prevent="handlePromoDragOver($event, index)"
              @dragend="handlePromoDragEnd"
              @drop="handlePromoDrop($event, index)"
            >
              <img :src="cleanUrl(pic.image_url || pic)" class="promotional-image" alt="宣传图" />
              <div class="promo-actions">
                <el-button type="danger" size="small" link @click.stop="removePromoPic(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <div class="promo-order">
                {{ index + 1 }}
              </div>
            </div>
          </div>
          <div v-else class="empty-tip">暂无宣传图</div>
        </el-card>

        <el-card class="style-codes-card">
          <template #header>
            <div class="card-header">
              <span>涉及款式</span>
              <el-button type="primary" size="small" @click="openEditDialog">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
            </div>
          </template>
          <div v-if="activityData.style_codes && activityData.style_codes.length > 0" class="style-codes-list">
            <el-tag v-for="(code, index) in activityData.style_codes" :key="index" class="style-code-tag">
              {{ code }}
            </el-tag>
          </div>
          <div v-else class="empty-tip">暂无款式</div>
        </el-card>

        <el-card class="info-card">
          <template #header>
            <span>基本信息</span>
          </template>
          <div class="info-list">
            <div class="info-item">
              <span class="label">ID：</span>
              <span>{{ activityData.id }}</span>
            </div>
            <div class="info-item">
              <span class="label">分类：</span>
              <span>{{ activityData.category || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">排序：</span>
              <span>{{ activityData.order || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">状态：</span>
              <el-tag :type="activityData.status === 'online' ? 'success' : 'info'">
                {{ activityData.status === 'online' ? '上线' : activityData.status === 'offline' ? '下线' : '待处理' }}
              </el-tag>
            </div>
            <div class="info-item">
              <span class="label">上线时间：</span>
              <span>{{ activityData.online_time || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">下线时间：</span>
              <span>{{ activityData.offline_time || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">创建时间：</span>
              <span>{{ activityData.created_at }}</span>
            </div>
            <div class="info-item">
              <span class="label">更新时间：</span>
              <span>{{ activityData.updated_at }}</span>
            </div>
            <div v-if="activityData.notes" class="info-item">
              <span class="label">备注：</span>
              <span>{{ activityData.notes }}</span>
            </div>
          </div>
        </el-card>
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
                    <div v-for="(pic, index) in currentPromoPics" :key="index" class="image-item">
                      <img :src="cleanUrl(pic.image_url || pic)" alt="promo-image" class="item-image" />
                    </div>
                  </div>
                </div>
              </div>
              <div class="tab-bar">
                <div class="tab-item">
                  <div class="tab-icon active"></div>
                  <div class="tab-text active">首页</div>
                </div>
                <div class="tab-item">
                  <div class="tab-icon"></div>
                  <div class="tab-text">童装</div>
                </div>
                <div class="tab-item">
                  <div class="tab-icon"></div>
                  <div class="tab-text">成人装</div>
                </div>
                <div class="tab-item">
                  <div class="tab-icon"></div>
                  <div class="tab-text">感兴趣</div>
                </div>
                <div class="tab-item">
                  <div class="tab-icon"></div>
                  <div class="tab-text">我的</div>
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
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Plus, Delete, Edit } from '@element-plus/icons-vue'
import { getActivityImageDetail, addPromotionalPic, updateActivityImageRelations, getAllCategories, updatePromotionalPicOrder, setActivityDetail } from '@/api'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const activityData = ref<any>(null)
const promoFileInput = ref<HTMLInputElement>()
const editDialogVisible = ref(false)
const savingEdit = ref(false)
const categories = ref<string[]>([])
const editForm = ref({
  category: '',
  style_codes: [] as string[]
})
const styleCodesInput = ref('')
const promoDragIndex = ref<number | null>(null)
const isPromoOrderChanged = ref(false)
const originalPromoPics = ref<any[]>([])
const tempPromoPics = ref<any[]>([])
const hasDetailSwitch = ref(false)

const cleanUrl = (url: string) => {
  if (!url) return ''
  return url.replace(/[`\s]/g, '')
}

const styleCodesList = computed(() => {
  return styleCodesInput.value
    .split(/[\n,]/)
    .map(s => s.trim())
    .filter(s => s.length > 0)
})

const currentPromoPics = computed(() => {
  return isPromoOrderChanged.value ? tempPromoPics.value : (activityData.value?.promotional_pics || [])
})

const loadCategories = async () => {
  try {
    const res = await getAllCategories({ shopname: 'youlan_kids' })
    if (res.code === 200 && res.data) {
      categories.value = res.data.categories || []
    }
  } catch (error) {
    console.error('获取分类失败:', error)
  }
}

const openEditDialog = () => {
  editForm.value = {
    category: activityData.value.category || '',
    style_codes: activityData.value.style_codes ? [...activityData.value.style_codes] : []
  }
  styleCodesInput.value = editForm.value.style_codes.join('\n')
  editDialogVisible.value = true
}

const saveEdit = async () => {
  try {
    savingEdit.value = true
    const params: any = {
      activity_id: activityData.value.id
    }
    if (editForm.value.category !== undefined && editForm.value.category !== null) {
      params.category = editForm.value.category
    }
    if (styleCodesList.value.length > 0) {
      params.style_codes = styleCodesList.value
    }
    
    const res = await updateActivityImageRelations(params)
    if (res.code === 200) {
      ElMessage.success('保存成功')
      editDialogVisible.value = false
      // 刷新数据
      loadActivityDetail()
    } else {
      ElMessage.error(res.msg || '保存失败')
    }
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败')
  } finally {
    savingEdit.value = false
  }
}

const processPromotionalPics = (pics: any) => {
  if (!pics) return []
  if (Array.isArray(pics)) return pics
  // 如果是对象格式，按order排序后保持对象数组
  const picArray = Object.values(pics) as any[]
  // 按order排序
  picArray.sort((a: any, b: any) => a.order - b.order)
  return picArray
}

const loadActivityDetail = async () => {
  const id = route.params.id
  if (!id) {
    ElMessage.error('活动ID不能为空')
    return
  }

  loading.value = true
  try {
    const res = await getActivityImageDetail({ activity_id: Number(id) })
    if (res.code === 200 && res.data) {
      const data = res.data
      // 处理宣传图数据结构
      data.promotional_pics = processPromotionalPics(data.promotional_pics)
      activityData.value = data
      hasDetailSwitch.value = !!data.has_activity_detail
    } else {
      ElMessage.error(res.msg || '获取详情失败')
    }
  } catch (error) {
    console.error('获取活动详情失败:', error)
    ElMessage.error('获取活动详情失败')
  } finally {
    loading.value = false
  }
}

const handleSetDetail = async (value: boolean) => {
  try {
    await ElMessageBox.confirm(`确定要设置为${value ? '可跳转' : '不可跳转'}吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await setActivityDetail({
      activity_id: activityData.value.id,
      has_activity_detail: value
    })
    if (res.code === 200) {
      activityData.value.has_activity_detail = value
      ElMessage.success('设置成功')
    } else {
      ElMessage.error(res.msg || '设置失败')
      hasDetailSwitch.value = !value
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('设置失败:', error)
      ElMessage.error('设置失败')
      hasDetailSwitch.value = !value
    }
  }
}

const triggerPromoUpload = () => {
  promoFileInput.value?.click()
}

const handlePromoUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  try {
    const formData = new FormData()
    formData.append('activity_id', String(activityData.value.id))
    formData.append('image', file)
    const res = await addPromotionalPic(formData)
    if (res.code === 200 && res.data) {
      if (isPromoOrderChanged.value) {
        tempPromoPics.value.push(res.data)
      } else {
        if (!activityData.value.promotional_pics) {
          activityData.value.promotional_pics = []
        }
        activityData.value.promotional_pics.push(res.data)
      }
      ElMessage.success('宣传图上传成功')
    } else {
      ElMessage.error(res.msg || '上传失败')
    }
  } catch (error) {
    console.error('上传宣传图失败:', error)
    ElMessage.error('上传宣传图失败')
  } finally {
    if (target) {
      target.value = ''
    }
  }
}

const removePromoPic = (index: number) => {
  if (isPromoOrderChanged.value) {
    tempPromoPics.value.splice(index, 1)
  } else {
    activityData.value.promotional_pics.splice(index, 1)
  }
  ElMessage.success('宣传图已移除')
}

const handlePromoDragStart = (event: DragEvent, index: number) => {
  promoDragIndex.value = index
  // 开始拖拽时，如果还没有进入编辑模式，先保存原始数据
  if (!isPromoOrderChanged.value && activityData.value?.promotional_pics) {
    originalPromoPics.value = [...activityData.value.promotional_pics]
    tempPromoPics.value = [...activityData.value.promotional_pics]
  }
}

const handlePromoDragOver = (event: DragEvent, index: number) => {
  // 防止默认行为以允许拖拽
}

const handlePromoDragEnd = () => {
  promoDragIndex.value = null
}

const handlePromoDrop = (event: DragEvent, targetIndex: number) => {
  if (promoDragIndex.value === null || promoDragIndex.value === targetIndex) {
    return
  }
  
  // 如果还没有标记为已修改，标记一下
  if (!isPromoOrderChanged.value) {
    isPromoOrderChanged.value = true
  }
  
  // 只更新临时状态，不调用API
  const pics = [...tempPromoPics.value]
  const [removed] = pics.splice(promoDragIndex.value, 1)
  pics.splice(targetIndex, 0, removed)
  tempPromoPics.value = pics
}

const cancelPromoOrder = () => {
  isPromoOrderChanged.value = false
  tempPromoPics.value = []
  originalPromoPics.value = []
  ElMessage.info('已取消调整')
}

const savePromoOrder = async () => {
  try {
    // 找到被移动的图片及其原始order
    // 创建一个map记录原始order
    const originalOrderMap = new Map()
    originalPromoPics.value.forEach((pic, idx) => {
      const url = typeof pic === 'object' ? cleanUrl(pic.image_url) : cleanUrl(pic)
      let order = idx + 1
      if (typeof pic === 'object' && pic.order !== null && pic.order !== undefined) {
        order = pic.order
      }
      originalOrderMap.set(url, order)
      console.log(`[原始位置] ${url} -> ${order}`)
    })
    
    // 遍历临时列表，找到位置变化的图片，逐个调用API
    for (let i = 0; i < tempPromoPics.value.length; i++) {
      const pic = tempPromoPics.value[i]
      const url = typeof pic === 'object' ? cleanUrl(pic.image_url) : cleanUrl(pic)
      const originalOrder = originalOrderMap.get(url)
      const newOrder = i + 1
      
      console.log(`[处理] ${url}: original=${originalOrder}, new=${newOrder}`)
      
      // 如果order没有变化，跳过
      if (originalOrder === newOrder) continue
      
      if (originalOrder === undefined || originalOrder === null) {
        console.warn(`未找到 ${url} 的原始order，跳过`)
        continue
      }
      
      // 调用API调整这个图片的位置
      const res = await updatePromotionalPicOrder({
        activity_id: activityData.value.id,
        old_order: originalOrder,
        new_order: newOrder
      })
      
      if (res.code !== 200) {
        ElMessage.error(res.msg || '保存失败')
        return
      }
    }
    
    // 更新最终数据并重置状态
    tempPromoPics.value.forEach((pic, idx) => {
      if (typeof pic === 'object') {
        pic.order = idx + 1
      }
    })
    activityData.value.promotional_pics = tempPromoPics.value
    isPromoOrderChanged.value = false
    tempPromoPics.value = []
    originalPromoPics.value = []
    ElMessage.success('调整位置成功')
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败')
  }
}

const goBack = () => {
  router.back()
}

onMounted(() => {
  loadActivityDetail()
  loadCategories()
})
</script>

<style scoped>
.activity-detail {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 400px;
}

.detail-content {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.left-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.right-panel {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
}

.main-image-card,
.promotional-pics-card,
.style-codes-card,
.info-card {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.main-image-wrapper {
  display: flex;
  justify-content: flex-start;
}

.main-image {
  max-width: 200px;
  width: 100%;
  height: auto;
  border-radius: 8px;
}

.promotional-pics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
}

.promotional-pic-item {
  width: 100%;
  aspect-ratio: 4/3;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
  cursor: grab;
  transition: transform 0.2s, box-shadow 0.2s;
}

.promotional-pic-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.promotional-pic-item.dragging {
  opacity: 0.5;
  transform: scale(1.05);
  cursor: grabbing;
}

.promotional-pic-item:hover .promo-actions {
  opacity: 1;
}

.promo-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.3s;
  z-index: 10;
}

.promo-order {
  position: absolute;
  bottom: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
}

.promotional-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.main-image-card .card-header,
.promotional-pics-card .card-header,
.style-codes-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.style-codes-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.style-code-tag {
  font-size: 13px;
  padding: 6px 12px;
}

.empty-tip {
  color: #999;
  padding: 30px 0;
  text-align: center;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item {
  display: flex;
  align-items: center;
}

.info-item .label {
  width: 90px;
  color: #666;
  flex-shrink: 0;
}

/* 手机预览样式 - 与首页管理一致 */
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
</style>
