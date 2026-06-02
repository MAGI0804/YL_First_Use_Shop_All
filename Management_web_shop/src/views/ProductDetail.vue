<template>
  <div class="product-detail">
    <div class="page-header">
      <el-button link @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <span class="page-title">商品详情</span>
      <div class="header-edit-btn">
        <el-button type="primary" @click="toggleEdit" v-if="productData">
          <el-icon><Edit /></el-icon>
          {{ isEditing ? '取消编辑' : '编辑' }}
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-loading-spinner style="width: 32px; height: 32px;" />
      <span style="margin-left: 10px;">加载中...</span>
    </div>

    <div v-else-if="productData" class="detail-content">
      <el-card>
        <template #header>
          <span>基本信息</span>
        </template>
        <div class="basic-info">
          <div class="product-image-section">
            <img :src="cleanUrl(isEditing && editForm.mainImagePreview ? editForm.mainImagePreview : productData.main_image?.url)" class="main-image" />
            <div v-if="isEditing" class="edit-image-input">
              <el-upload
                v-model:file-list="mainImageFileList"
                :auto-upload="false"
                :limit="1"
                :on-change="handleMainImageChange"
                :on-exceed="handleMainImageExceed"
                accept="image/*"
                list-type="text"
              >
                <el-button type="primary">上传主图</el-button>
                <template #tip>
                  <div class="el-upload__tip">
                    只能上传一张图片
                  </div>
                </template>
              </el-upload>
            </div>
          </div>
          <div class="product-info">
            <div v-if="!isEditing">
              <h3 class="product-name">{{ productData.name }}</h3>
            </div>
            <div v-else>
              <el-input v-model="editForm.name" placeholder="商品名称" size="large" />
            </div>
            <div class="info-item">
              <span class="label">款式编码：</span>
              <span>{{ styleCode }}</span>
            </div>
            <div class="info-item">
              <span class="label">类别：</span>
              <span v-if="!isEditing">{{ productData.category || '-' }}</span>
              <el-select v-else v-model="editForm.category" placeholder="请选择类别" size="small" style="width: 200px;">
                <el-option v-for="cat in allCategories" :key="cat" :label="cat" :value="cat" />
              </el-select>
            </div>
            <div class="info-item">
              <span class="label">款式：</span>
              <span v-if="!isEditing">{{ productData.labels?.label_five || '-' }}</span>
              <el-select v-else v-model="editForm.labels.label_five" placeholder="请选择款式" size="small" style="width: 200px;">
                <el-option v-for="option in allLabels.label_five || []" :key="option" :label="option" :value="option" />
              </el-select>
            </div>
            <div class="info-item">
              <span class="label">价格：</span>
              <span v-if="!isEditing" class="price">¥{{ productData.price }}</span>
              <el-input-number v-else v-model="editForm.price" :min="0" :precision="2" size="small" />
            </div>
            <div class="info-item">
              <span class="label">总库存：</span>
              <span>{{ productData.inventory }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <el-card style="margin-top: 20px;">
        <template #header>
          <span>商品标签</span>
        </template>
        <div class="labels-container">
          <div v-if="!isEditing">
            <div class="label-item" v-for="(value, key) in safeLabels" :key="key">
              <span class="label-key">{{ getLabelName(String(key)) }}：</span>
              <el-tag type="info">{{ value || '-' }}</el-tag>
            </div>
            <div v-if="Object.keys(safeLabels).length === 0" style="color: #999;">暂无标签</div>
          </div>
          <div v-else class="edit-labels">
            <div class="label-edit-item" v-for="(options, key) in safeAllLabels" :key="key">
              <span class="label-key">{{ getLabelName(String(key)) }}：</span>
              <el-select v-model="editForm.labels[key]" placeholder="请选择" style="width: 200px;">
                <el-option v-for="option in options" :key="option" :label="option" :value="option" />
              </el-select>
            </div>
          </div>
        </div>
      </el-card>

      <el-card style="margin-top: 20px;">
        <template #header>
          <span>商品图片</span>
        </template>
        <div class="image-list">
          <!-- 非编辑状态显示原图片 -->
          <template v-if="!isEditing">
            <div
              v-for="(url, key) in getDisplayPictures()"
              :key="`display-${key}`"
              class="image-item"
            >
              <img :src="cleanUrl(url)" />
            </div>
          </template>
          
          <!-- 编辑状态显示图片位置 -->
          <template v-else>
            <div
              v-for="(slot, index) in editForm.displayPictureSlots"
              :key="`slot-${slot.position}`"
              class="image-item"
              :class="{ 'cleared': slot.cleared, 'empty': !slot.previewUrl && !slot.cleared }"
            >
              <div v-if="slot.previewUrl" class="image-wrapper">
                <img :src="slot.previewUrl" />
              </div>
              <div v-else class="image-placeholder">
                <span v-if="slot.cleared">已清空</span>
                <span v-else>位置 {{ slot.position }}</span>
              </div>
              
              <!-- 编辑按钮 -->
              <div class="image-overlay">
                <div class="action-buttons">
                  <el-button
                    type="primary"
                    size="small"
                    @click="handleSlotReplace(index)"
                    :disabled="slot.cleared"
                  >
                    替换
                  </el-button>
                  <el-button
                    v-if="!slot.cleared"
                    type="danger"
                    size="small"
                    @click="handleSlotClear(index)"
                  >
                    清空
                  </el-button>
                  <el-button
                    v-else
                    type="success"
                    size="small"
                    @click="handleSlotRestore(index)"
                  >
                    恢复
                  </el-button>
                </div>
              </div>
              
              <!-- 位置标签 -->
              <div class="position-badge">位置 {{ slot.position }}</div>
            </div>
            
            <!-- 添加新位置按钮 -->
            <div class="image-item add-slot" @click="handleAddSlot">
              <div class="add-icon">
                <el-icon size="40"><Plus /></el-icon>
              </div>
              <span>添加位置</span>
            </div>
          </template>
        </div>
        
        <!-- 隐藏文件输入 -->
        <input
          ref="hiddenFileInput"
          type="file"
          accept="image/*"
          style="display: none"
          @change="handleFileChange"
        />
      </el-card>

      <el-card style="margin-top: 20px;">
        <template #header>
          <span>库存规格</span>
        </template>
        <el-collapse v-model="activeColors">
          <el-collapse-item
            v-for="(item, index) in productData.items"
            :key="index"
            :name="index"
          >
            <template #title>
              <div class="color-title">
                <img :src="cleanUrl(item.color_image)" class="color-preview" />
                <span>{{ item.color }}</span>
                <el-tag type="info" size="small">{{ item.sizes.length }}个规格</el-tag>
              </div>
            </template>
            <el-table :data="item.sizes" style="width: 100%;">
              <el-table-column prop="size" label="尺码" />
              <el-table-column prop="commodity_id" label="商品ID" />
              <el-table-column prop="inventory" label="库存" />
            </el-table>
          </el-collapse-item>
        </el-collapse>
      </el-card>

      <!-- 固定保存和取消按钮 -->
      <div v-if="isEditing" class="fixed-save-buttons">
        <el-button class="square-btn" @click="cancelEdit">取消</el-button>
        <el-button type="primary" class="square-btn" @click="handleSave" :loading="saving">保存</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadFile, type UploadUserFile } from 'element-plus'
import { ArrowLeft, Edit, Plus } from '@element-plus/icons-vue'
import { getToken, getStyleCodeCommodity, getAllLabels, getAllCategories } from '@/api'
import type { StyleCodeCommodityData, StyleCodeCommodityLabels } from '@/api'
import http from '@/api/request'

const router = useRouter()
const route = useRoute()
const loading = ref(true)
const productData = ref<StyleCodeCommodityData | null>(null)
const styleCode = ref('')
const activeColors = ref<number[]>([0])

const isEditing = ref(false)
const saving = ref(false)
const allLabels = ref<Record<string, string[]>>({})
const allCategories = ref<string[]>([])
const mainImageFileList = ref<UploadUserFile[]>([])
const displayPicturesFileList = ref<UploadUserFile[]>([])
const editForm = ref<{
  name: string
  category: string
  price: number
  labels: StyleCodeCommodityLabels
  mainImageFile: File | null
  mainImagePreview: string
  // 展示图片位置
  displayPictureSlots: Array<{
    position: number
    originalUrl: string | null
    newFile: File | null
    previewUrl: string
    cleared: boolean
  }>
} | null>(null)

const safeLabels = computed(() => {
  if (!productData.value || !productData.value.labels) {
    return {}
  }
  const { label_five, ...otherLabels } = productData.value.labels
  return otherLabels
})

const safeAllLabels = computed(() => {
  const { label_five, ...otherLabels } = allLabels.value
  return otherLabels
})

const cleanUrl = (url: string) => {
  if (!url) return ''
  return url.replace(/[`\s]/g, '')
}

const getLabelName = (key: string) => {
  const labelMap: Record<string, string> = {
    label_one: '年份',
    label_two: '季节',
    label_three: '分类',
    label_four: '品类',
    label_five: '款式',
    label_six: '面料',
    label_seven: '品类细'
  }
  return labelMap[key] || key
}

const getDisplayPictures = () => {
  if (!productData.value || !productData.value.display_pictures) {
    return {}
  }
  return productData.value.display_pictures
}

const goBack = () => {
  router.back()
}

const fetchProductDetail = async () => {
  try {
    await getToken()
    const code = route.params.styleCode as string
    styleCode.value = code
    
    const res = await getStyleCodeCommodity({
      shopname: 'youlan_kids',
      style_code: code
    })
    
    console.log('API返回数据:', res)
    
    if (res.code === 200 && res.data) {
      productData.value = res.data
      console.log('商品数据:', productData.value)
      console.log('类别:', productData.value.category)
      console.log('标签:', productData.value.labels)
    } else {
      ElMessage.error(res.msg || '获取商品详情失败')
    }
  } catch (error) {
    console.error('获取商品详情失败:', error)
    ElMessage.error('获取商品详情失败')
  } finally {
    loading.value = false
  }
}

const fetchLabelsAndCategories = async () => {
  try {
    const [labelsRes, categoriesRes] = await Promise.all([
      getAllLabels({ shopname: 'youlan_kids' }),
      getAllCategories({ shopname: 'youlan_kids' })
    ])
    
    if (labelsRes.code === 200 && labelsRes.data) {
      allLabels.value = labelsRes.data
    }
    
    if (categoriesRes.code === 200 && categoriesRes.data) {
      allCategories.value = categoriesRes.data.categories || []
    }
  } catch (error) {
    console.error('获取标签和类别失败:', error)
  }
}

const toggleEdit = () => {
  if (isEditing.value) {
    cancelEdit()
  } else {
    startEdit()
  }
}

const startEdit = () => {
  if (!productData.value) return
  
  mainImageFileList.value = []
  displayPicturesFileList.value = []
  
  // 初始化展示图片位置
  const displayPictureSlots: Array<{
    position: number
    originalUrl: string | null
    newFile: File | null
    previewUrl: string
    cleared: boolean
  }> = []
  
  if (productData.value.display_pictures) {
    // 找到最大位置号
    let maxPosition = 0
    Object.keys(productData.value.display_pictures).forEach(key => {
      const pos = parseInt(key)
      if (!isNaN(pos) && pos > maxPosition) {
        maxPosition = pos
      }
    })
    
    // 初始化位置，至少有一个位置
    const positionsToInit = Math.max(1, maxPosition)
    for (let i = 1; i <= positionsToInit; i++) {
      const url = productData.value.display_pictures[i]
      displayPictureSlots.push({
        position: i,
        originalUrl: url ? String(url) : null,
        newFile: null,
        previewUrl: url ? cleanUrl(String(url)) : '',
        cleared: false
      })
    }
  } else {
    // 默认一个空位置
    displayPictureSlots.push({
      position: 1,
      originalUrl: null,
      newFile: null,
      previewUrl: '',
      cleared: false
    })
  }
  
  editForm.value = {
    name: productData.value.name,
    category: productData.value.category,
    price: productData.value.price,
    labels: { ...productData.value.labels },
    mainImageFile: null,
    mainImagePreview: '',
    displayPictureSlots
  }
  
  isEditing.value = true
}

const cancelEdit = () => {
  isEditing.value = false
  editForm.value = null
  mainImageFileList.value = []
  displayPicturesFileList.value = []
}

const handleMainImageChange = (file: UploadFile) => {
  if (!editForm.value) return
  
  if (file.raw) {
    editForm.value.mainImageFile = file.raw
    const reader = new FileReader()
    reader.onload = (e) => {
      if (editForm.value && e.target?.result) {
        editForm.value.mainImagePreview = e.target.result as string
      }
    }
    reader.readAsDataURL(file.raw)
  }
}

const handleMainImageExceed = () => {
  ElMessage.warning('只能上传一张主图')
}

// 当前正在编辑的图片位置
const currentEditingSlotIndex = ref<number | null>(null)
// 隐藏的文件输入引用
const hiddenFileInput = ref<HTMLInputElement | null>(null)

const handleSlotReplace = (index: number) => {
  currentEditingSlotIndex.value = index
  // 触发隐藏文件输入点击
  if (hiddenFileInput.value) {
    hiddenFileInput.value.click()
  }
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (!files || !files.length || currentEditingSlotIndex.value === null || !editForm.value) {
    return
  }
  
  const file = files[0]
  const slotIndex = currentEditingSlotIndex.value
  
  const reader = new FileReader()
  reader.onload = (e) => {
    if (editForm.value && e.target?.result) {
      editForm.value.displayPictureSlots[slotIndex].newFile = file
      editForm.value.displayPictureSlots[slotIndex].previewUrl = e.target.result as string
      editForm.value.displayPictureSlots[slotIndex].cleared = false
    }
  }
  reader.readAsDataURL(file)
  
  // 清空输入以便再次选择相同文件
  target.value = ''
  currentEditingSlotIndex.value = null
}

const handleSlotClear = (index: number) => {
  if (!editForm.value) return
  editForm.value.displayPictureSlots[index].cleared = true
  editForm.value.displayPictureSlots[index].newFile = null
  editForm.value.displayPictureSlots[index].previewUrl = ''
}

const handleSlotRestore = (index: number) => {
  if (!editForm.value) return
  editForm.value.displayPictureSlots[index].cleared = false
  editForm.value.displayPictureSlots[index].newFile = null
  editForm.value.displayPictureSlots[index].previewUrl = editForm.value.displayPictureSlots[index].originalUrl 
    ? cleanUrl(editForm.value.displayPictureSlots[index].originalUrl) 
    : ''
}

const handleAddSlot = () => {
  if (!editForm.value) return
  const newPosition = editForm.value.displayPictureSlots.length + 1
  editForm.value.displayPictureSlots.push({
    position: newPosition,
    originalUrl: null,
    newFile: null,
    previewUrl: '',
    cleared: false
  })
}

const handleSave = async () => {
  if (!editForm.value) return
  
  try {
    await ElMessageBox.confirm('确定要保存修改吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  
  try {
    saving.value = true
    
    const formData = new FormData()
    formData.append('shopname', 'youlan_kids')
    formData.append('style_code', styleCode.value)
    formData.append('name', editForm.value.name)
    formData.append('category', editForm.value.category)
    formData.append('price', editForm.value.price.toString())
    
    Object.entries(editForm.value.labels).forEach(([key, value]) => {
      if (value) {
        formData.append(`labels[${key}]`, value)
      }
    })
    
    if (editForm.value.mainImageFile) {
      formData.append('image', editForm.value.mainImageFile)
    }
    
    // 处理展示图片：按位置处理
    editForm.value.displayPictureSlots.forEach(slot => {
      if (slot.cleared) {
        // 清空该位置：传空字符串
        formData.append(`display_pictures[${slot.position}]`, '')
      } else if (slot.newFile) {
        // 有新文件：传文件
        formData.append(`display_pictures[${slot.position}]`, slot.newFile)
      } else if (slot.originalUrl) {
        // 保留原图片：传原URL
        formData.append(`display_pictures[${slot.position}]`, cleanUrl(slot.originalUrl))
      }
    })
    
    const res = await http.post('/commodity/update_style_code_info', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    if (res.code === 200) {
      ElMessage.success('保存成功')
      isEditing.value = false
      editForm.value = null
      mainImageFileList.value = []
      displayPicturesFileList.value = []
      await fetchProductDetail()
    } else {
      ElMessage.error(res.msg || '保存失败')
    }
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchProductDetail()
  fetchLabelsAndCategories()
})
</script>

<style scoped>
.product-detail {
  padding: 20px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-shrink: 0;
  position: relative;
}

.header-edit-btn {
  position: fixed;
  right: 20px;
  top: 20px;
  z-index: 100;
}

.page-title {
  font-size: 18px;
  font-weight: 500;
  color: #1a1a1a;
  flex: 1;
}

.basic-info {
  display: flex;
  gap: 24px;
}

.product-image-section {
  flex-shrink: 0;
  position: relative;
}

.main-image {
  width: 200px;
  height: 200px;
  border-radius: 8px;
  object-fit: cover;
  background: #f5f5f5;
}

.edit-image-input {
  margin-top: 8px;
  width: 200px;
}

.product-info {
  flex: 1;
}

.product-name {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #1a1a1a;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
}

.info-item .label {
  color: #999;
  width: 80px;
}

.info-item .price {
  font-size: 20px;
  font-weight: 600;
  color: #f56c6c;
}

.image-list {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.image-item {
  width: 100px;
  height: 100px;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  border: 2px solid #e0e0e0;
  transition: all 0.3s ease;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
}

.image-item.cleared {
  border-color: #f56c6c;
  background: #fef0f0;
}

.image-item.cleared .image-placeholder span {
  color: #f56c6c;
}

.image-item.empty {
  border: 2px dashed #d9d9d9;
  background: #fafafa;
}

.image-item.add-slot {
  border: 2px dashed #409eff;
  background: #f0f7ff;
  cursor: pointer;
}

.image-item.add-slot:hover {
  border-color: #66b1ff;
  background: #ecf5ff;
}

.image-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-wrapper img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: #999;
  font-size: 14px;
}

.add-icon {
  margin-bottom: 8px;
  color: #409eff;
}

.image-item:hover .image-overlay {
  opacity: 1;
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 2;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.position-badge {
  position: absolute;
  top: 4px;
  left: 4px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  z-index: 3;
}

.image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-item .el-tag {
  position: absolute;
  top: 8px;
  right: 8px;
}

.color-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.color-preview {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
}

.labels-container {
  min-height: 40px;
}

.label-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-right: 16px;
  margin-bottom: 8px;
}

.edit-labels {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.label-edit-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.label-key {
  color: #666;
  font-size: 14px;
}

.action-buttons {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  gap: 16px;
}

.fixed-save-buttons {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 12px;
  z-index: 100;
}

.square-btn {
  min-width: 80px;
  height: 40px;
  border-radius: 4px;
}
</style>
