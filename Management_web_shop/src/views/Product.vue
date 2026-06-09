<template>
  <div class="product-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="searchCode"
          placeholder="编码搜索"
          style="width: 180px;"
          clearable
        />

        <CompactDateRangePicker v-model="dateRange" />

        <el-select v-model="queryParams.category" placeholder="商品类别" style="width: 150px;">
          <el-option
            v-for="cat in categories"
            :key="cat"
            :label="cat"
            :value="cat"
          />
        </el-select>

        <el-select v-model="selectedClothingType" placeholder="服装分类" style="width: 150px;">
          <el-option label="全部" value="" />
          <el-option label="童装" value="kids" />
          <el-option label="成人装" value="adult" />
        </el-select>

        <el-select v-model="queryParams.status" placeholder="商品状态" style="width: 150px;">
          <el-option label="全部" value="" />
          <el-option label="上架" value="online" />
          <el-option label="下架" value="offline" />
        </el-select>

        <el-popover
          placement="bottom"
          :width="500"
          trigger="click"
          v-model:visible="labelPopoverVisible"
        >
          <div class="label-popover">
            <div class="popover-label-group" v-for="(labelList, labelKey) in otherLabels" :key="labelKey">
              <span class="popover-label-title">{{ getLabelGroupName(labelKey) }}:</span>
              <div class="popover-label-tags">
                <el-tag
                  v-for="label in labelList"
                  :key="label"
                  :type="selectedLabels[labelKey]?.includes(label) ? 'primary' : 'info'"
                  class="label-tag"
                  @click="toggleLabel(labelKey, label)"
                >
                  {{ label }}
                </el-tag>
              </div>
            </div>
          </div>
          <template #reference>
            <el-button>标签</el-button>
          </template>
        </el-popover>

        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button :icon="Download" @click="handleExportTask">导出</el-button>
      </div>
    </div>

    <div v-if="initialLoading" class="loading-container">
      <el-loading-spinner style="width: 32px; height: 32px;" />
      <span style="margin-left: 10px;">加载中...</span>
    </div>

    <div v-else>
      <el-table :data="goodsList" style="width: 100%; margin-top: 20px;">
        <el-table-column label="商品图片" width="100">
          <template #default="{ row }">
            <div v-if="row.promo_image_url" class="product-image-wrapper">
              <img :src="row.promo_image_url" class="product-image" loading="lazy" decoding="async" />
            </div>
            <div v-else class="product-image"></div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="商品名称" min-width="250" />
        <el-table-column prop="style_code" label="款式编码" width="140" />
        <el-table-column prop="price" label="价格" width="100">
          <template #default="{ row }">
            ¥{{ row.price }}
          </template>
        </el-table-column>
        <el-table-column prop="online_status" label="编码状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.online_status === 'online' ? 'success' : 'info'" size="small">
              {{ row.online_status === 'online' ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column prop="online_time" label="操作时间" width="160" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="primary" link @click="viewDetail(row.style_code)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import CompactDateRangePicker from '@/components/CompactDateRangePicker.vue'
import { getToken, getAllLabels, getAllCategories, goodsQuery, createDownloadTask } from '@/api'

const router = useRouter()
const initialLoading = ref(true)
const categories = ref<string[]>([])
const labels = ref<Record<string, string[]>>({})
const selectedLabels = reactive<Record<string, string[]>>({})
const selectedClothingType = ref('')
const labelPopoverVisible = ref(false)
const goodsList = ref<any[]>([])
const total = ref(0)
const dateRange = ref<[string, string] | null>(null)
const searchCode = ref('')

const CACHE_KEY = 'product_first_page_cache'
const CACHE_EXPIRE = 5 * 60 * 1000 // 5分钟缓存

const loadFromCache = () => {
  try {
    const cached = localStorage.getItem(CACHE_KEY)
    if (cached) {
      const { data, timestamp } = JSON.parse(cached)
      if (Date.now() - timestamp < CACHE_EXPIRE) {
        goodsList.value = data.goodsList
        total.value = data.total
        return true
      }
    }
  } catch (e) {
    console.error('读取缓存失败:', e)
  }
  return false
}

const saveToCache = (goods: any[], totalCount: number) => {
  try {
    const data = {
      goodsList: goods,
      total: totalCount
    }
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      data,
      timestamp: Date.now()
    }))
  } catch (e) {
    console.error('保存缓存失败:', e)
  }
}

const queryParams = reactive({
  shopname: 'youlan_kids',
  page: 1,
  page_size: 10,
  demand: 'style_code',
  category: '全部',
  status: '',
  begin_time: '',
  end_time: ''
})

const otherLabels = computed(() => {
  const result: Record<string, string[]> = {}
  Object.keys(labels.value).forEach(key => {
    if (key !== 'label_three') {
      result[key] = labels.value[key]
    }
  })
  return result
})

const getLabelGroupName = (key: string) => {
  const map: Record<string, string> = {
    label_one: '年份',
    label_two: '季节',
    label_three: '分类',
    label_four: '品类',
    label_seven: '品类细'
  }
  return map[key] || key
}

const toggleLabel = (labelKey: string, label: string) => {
  if (!selectedLabels[labelKey]) {
    selectedLabels[labelKey] = []
  }
  const index = selectedLabels[labelKey].indexOf(label)
  if (index > -1) {
    selectedLabels[labelKey].splice(index, 1)
  } else {
    selectedLabels[labelKey].push(label)
  }
}

const fetchLabelsAndCategories = async () => {
  try {
    await getToken()
    const [labelsRes, categoriesRes] = await Promise.all([
      getAllLabels({ shopname: 'youlan_kids' }),
      getAllCategories({ shopname: 'youlan_kids' })
    ])

    if (labelsRes.code === 200 && labelsRes.data) {
      labels.value = labelsRes.data
      Object.keys(labelsRes.data).forEach(key => {
        if (!selectedLabels[key]) {
          selectedLabels[key] = []
        }
      })
    }

    if (categoriesRes.code === 200 && categoriesRes.data) {
      categories.value = categoriesRes.data.categories
    }
  } catch (error) {
    console.error('获取标签和类目失败:', error)
    ElMessage.error('获取标签和类目失败')
  }
}

const fetchGoodsList = async () => {
  const isFirstPage = queryParams.page === 1 && !searchCode.value && !dateRange.value && queryParams.category === '全部' && !selectedClothingType.value && !queryParams.status && Object.values(selectedLabels).every(labels => labels.length === 0)
  
  if (isFirstPage && loadFromCache()) {
    initialLoading.value = false
    return
  }
  
  try {
    const params: any = {
      ...queryParams
    }

    if (dateRange.value && dateRange.value.length === 2) {
      params.begin_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }

    if (selectedClothingType.value === 'kids') {
      params.label_three = ['ACC', 'KIDS']
    } else if (selectedClothingType.value === 'adult') {
      params.label_three = ['成人']
    }

    if (searchCode.value) {
      params.style_code = searchCode.value
    }

    Object.keys(selectedLabels).forEach(key => {
      if (key !== 'label_three' && selectedLabels[key].length > 0) {
        params[key] = selectedLabels[key]
      }
    })

    const res = await goodsQuery(params)
    if (res.code === 200 && res.data) {
      goodsList.value = res.data.data
      total.value = res.data.total
      
      if (isFirstPage) {
        saveToCache(res.data.data, res.data.total)
      }
    } else {
      ElMessage.error(res.msg || '查询商品失败')
    }
  } catch (error) {
    console.error('查询商品失败:', error)
    ElMessage.error('查询商品失败')
  }
}

const handleSearch = () => {
  queryParams.page = 1
  fetchGoodsList()
}

const handleReset = () => {
  searchCode.value = ''
  queryParams.category = '全部'
  queryParams.status = ''
  queryParams.page = 1
  queryParams.page_size = 10
  dateRange.value = null
  selectedClothingType.value = ''
  Object.keys(selectedLabels).forEach(key => {
    selectedLabels[key] = []
  })
  fetchGoodsList()
}

const handleExportTask = async () => {
  try {
    await createDownloadTask({
      template_code: 'product_export',
      file_format: 'xlsx',
      filters: {
        category: queryParams.category && queryParams.category !== '全部' ? queryParams.category : undefined,
        style_code: searchCode.value || undefined,
        begin_time: dateRange.value?.[0] || undefined,
        end_time: dateRange.value?.[1] || undefined
      }
    })
    ElMessage.success('商品下载任务已创建，请到下载中心查看')
  } catch (error) {
    console.error('create product download task failed:', error)
    ElMessage.error('商品下载任务创建失败')
  }
}

const handleSizeChange = (size: number) => {
  queryParams.page_size = size
  fetchGoodsList()
}

const handleCurrentChange = (page: number) => {
  queryParams.page = page
  fetchGoodsList()
}

const viewDetail = (styleCode: string) => {
  router.push(`/product/${styleCode}`)
}

onMounted(async () => {
  await fetchLabelsAndCategories()
  await fetchGoodsList()
  initialLoading.value = false
})
</script>

<style scoped>
.product-page {
  padding: 20px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.search-bar {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.search-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  cursor: pointer;
}

.product-image-wrapper {
  width: 60px;
  height: 60px;
}

.product-image {
  width: 60px;
  height: 60px;
  background: #f5f5f5;
  border-radius: 4px;
  object-fit: cover;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.label-popover {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.popover-label-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.popover-label-title {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.popover-label-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
