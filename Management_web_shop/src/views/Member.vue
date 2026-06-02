<template>
  <div class="member-page">
    <div class="top-bar">
      <div class="search-bar">
        <el-input
          v-model="phoneSearch"
          placeholder="请输入手机号"
          style="width: 150px;"
          clearable
        />
        <el-select v-model="levelFilter" placeholder="会员等级" style="width: 100px; margin-left: 12px;">
          <el-option label="全部" value="" />
          <el-option label="lv1" value="normal" />
          <el-option label="lv2" value="silver" />
          <el-option label="lv3" value="gold" />
          <el-option label="lv4" value="black" />
        </el-select>
        <div class="points-range" style="margin-left: 12px; display: flex; align-items: center;">
          <el-input-number v-model="pointsMin" :min="0" placeholder="最小" style="width: 120px;" />
          <span style="margin: 0 8px; font-size: 12px;">至</span>
          <el-input-number v-model="pointsMax" :min="0" placeholder="最大" style="width: 120px;" />
          <span style="margin-left: 4px; font-size: 12px; color: #999;">积分</span>
        </div>
        <el-select v-model="tagFilter" placeholder="用户标签" multiple collapse-tags collapse-tags-tooltip style="width: 140px; margin-left: 12px;">
          <el-option label="活跃用户" value="active" />
          <el-option label="高消费" value="high消费" />
          <el-option label="新用户" value="new" />
          <el-option label="沉睡用户" value="sleep" />
          <el-option label="VIP" value="vip" />
        </el-select>
        <el-button type="primary" style="margin-left: 12px;" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
      <div class="action-buttons">
        <el-button type="primary" @click="openAddDialog">新增会员</el-button>
        <el-button @click="handleBatchImport">批量新增</el-button>
      </div>
    </div>

    <el-table :data="filteredList" style="width: 100%; margin-top: 20px;" row-key="id">
      <el-table-column label="会员信息" min-width="160">
        <template #default="{ row }">
          <div class="member-info">
            <div class="member-name">{{ row.username }}</div>
            <div class="member-phone">{{ row.phone }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="level" label="会员等级" width="120">
        <template #default="{ row }">
          <el-tag :type="getLevelType(row.level)" size="small">
            {{ getLevelText(row.level) }}
          </el-tag>
          <el-button type="primary" link style="margin-left: 4px;" @click="openLevelDialog(row)">调整</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="points" label="积分" width="100" />
      <el-table-column label="标签" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="tag in row.tags" :key="tag" size="small" style="margin-right: 4px;">
            {{ getTagText(tag) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="注册时间" width="160" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
            {{ row.status === 'active' ? '正常' : '已暂停' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="viewDetail(row.id)">详情</el-button>
          <el-button type="warning" link @click="addTag(row)">加标签</el-button>
          <el-button :type="row.status === 'active' ? 'info' : 'success'" link @click="toggleStatus(row)">
            {{ row.status === 'active' ? '暂停' : '启用' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="10"
        :total="filteredList.length"
        layout="total, prev, pager, next"
      />
    </div>

    <el-dialog v-model="addDialogVisible" title="新增会员" width="500px">
      <el-form :model="newMember" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="newMember.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="newMember.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="会员等级">
          <el-select v-model="newMember.level" placeholder="请选择会员等级" style="width: 100%;">
            <el-option label="lv1" value="normal" />
            <el-option label="lv2" value="silver" />
            <el-option label="lv3" value="gold" />
            <el-option label="lv4" value="black" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAddMember">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="tagDialogVisible" title="添加标签" width="450px">
      <div style="margin-bottom: 16px;">
        <div style="margin-bottom: 8px; color: #666; font-size: 14px;">选择已有标签</div>
        <el-select v-model="selectedTags" multiple placeholder="请选择标签" style="width: 100%;">
          <el-option label="活跃用户" value="active" />
          <el-option label="高消费" value="high消费" />
          <el-option label="新用户" value="new" />
          <el-option label="沉睡用户" value="sleep" />
          <el-option label="VIP" value="vip" />
        </el-select>
      </div>
      <div>
        <div style="margin-bottom: 8px; color: #666; font-size: 14px;">自定义新标签</div>
        <div style="display: flex; gap: 8px;">
          <el-input v-model="customTag" placeholder="请输入新标签" />
          <el-button @click="addCustomTag">添加</el-button>
        </div>
        <div v-if="customTags.length > 0" style="margin-top: 12px;">
          <el-tag
            v-for="(tag, index) in customTags"
            :key="index"
            closable
            @close="removeCustomTag(index)"
            style="margin-right: 8px; margin-bottom: 8px;"
          >
            {{ tag }}
          </el-tag>
        </div>
      </div>
      <template #footer>
        <el-button @click="tagDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAddTag">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="levelDialogVisible" title="调整会员等级" width="400px">
      <el-form label-width="80px">
        <el-form-item label="会员昵称">
          <span>{{ levelRow?.username }}</span>
        </el-form-item>
        <el-form-item label="当前等级">
          <el-tag :type="getLevelType(levelRow?.level || '')" size="small">
            {{ getLevelText(levelRow?.level || '') }}
          </el-tag>
        </el-form-item>
        <el-form-item label="调整等级">
          <el-select v-model="selectedLevel" placeholder="请选择等级" style="width: 100%;">
            <el-option label="lv1" value="normal" />
            <el-option label="lv2" value="silver" />
            <el-option label="lv3" value="gold" />
            <el-option label="lv4" value="black" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="levelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmLevelChange">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const phoneSearch = ref('')
const levelFilter = ref('')
const pointsMin = ref<number | undefined>(undefined)
const pointsMax = ref<number | undefined>(undefined)
const tagFilter = ref<string[]>([])
const currentPage = ref(1)
const tagDialogVisible = ref(false)
const addDialogVisible = ref(false)
const levelDialogVisible = ref(false)
const selectedTags = ref<string[]>([])
const selectedLevel = ref('')
const customTag = ref('')
const customTags = ref<string[]>([])
const levelRow = ref<any>(null)
const currentRow = ref<any>(null)

const newMember = reactive({
  username: '',
  phone: '',
  level: 'normal',
  points: 0
})

const allMembers = ref([
  { id: 1, username: '张三', phone: '13888888888', level: 'gold', points: 5000, tags: ['active', 'vip'], createTime: '2024-03-15 10:30:00', status: 'active' },
  { id: 2, username: '李四', phone: '13999999999', level: 'silver', points: 2000, tags: ['new'], createTime: '2024-04-01 11:20:00', status: 'active' },
  { id: 3, username: '王五', phone: '13666666666', level: 'normal', points: 500, tags: [], createTime: '2024-04-10 14:15:00', status: 'active' },
  { id: 4, username: '赵六', phone: '13777777777', level: 'black', points: 10000, tags: ['high消费', 'vip'], createTime: '2024-02-20 09:20:00', status: 'paused' },
  { id: 5, username: '钱七', phone: '13555555555', level: 'normal', points: 800, tags: ['sleep'], createTime: '2024-01-05 16:30:00', status: 'active' },
])

const filteredList = computed(() => {
  let list = [...allMembers.value]
  
  if (phoneSearch.value) {
    list = list.filter(item => item.phone.includes(phoneSearch.value))
  }
  
  if (levelFilter.value) {
    list = list.filter(item => item.level === levelFilter.value)
  }
  
  if (pointsMin.value !== undefined) {
    list = list.filter(item => item.points >= pointsMin.value!)
  }
  
  if (pointsMax.value !== undefined) {
    list = list.filter(item => item.points <= pointsMax.value!)
  }
  
  if (tagFilter.value.length > 0) {
    list = list.filter(item => 
      tagFilter.value.some(tag => item.tags.includes(tag))
    )
  }
  
  return list
})

const getLevelType = (level: string) => {
  const map: Record<string, string> = {
    normal: 'info',
    silver: '',
    gold: 'warning',
    black: 'danger'
  }
  return map[level] || 'info'
}

const getLevelText = (level: string) => {
  const map: Record<string, string> = {
    normal: 'lv1',
    silver: 'lv2',
    gold: 'lv3',
    black: 'lv4'
  }
  return map[level] || level
}

const getTagText = (tag: string) => {
  const map: Record<string, string> = {
    active: '活跃用户',
    high消费: '高消费',
    new: '新用户',
    sleep: '沉睡用户',
    vip: 'VIP'
  }
  return map[tag] || tag
}

const handleSearch = () => {
  ElMessage.success(`找到 ${filteredList.value.length} 位会员`)
}

const handleReset = () => {
  phoneSearch.value = ''
  levelFilter.value = ''
  pointsMin.value = undefined
  pointsMax.value = undefined
  tagFilter.value = []
  ElMessage.success('已重置')
}

const openAddDialog = () => {
  newMember.username = ''
  newMember.phone = ''
  newMember.level = 'normal'
  newMember.points = 0
  addDialogVisible.value = true
}

const confirmAddMember = () => {
  if (!newMember.username || !newMember.phone) {
    ElMessage.warning('请填写完整信息')
    return
  }
  allMembers.value.unshift({
    id: Date.now(),
    ...newMember,
    tags: [],
    createTime: new Date().toLocaleString(),
    status: 'active'
  })
  ElMessage.success('会员添加成功')
  addDialogVisible.value = false
}

const handleBatchImport = () => {
  ElMessage.info('批量导入功能开发中，请使用Excel文件导入')
}

const viewDetail = (id: number) => {
  router.push(`/member/${id}`)
}

const addTag = (row: any) => {
  currentRow.value = row
  selectedTags.value = [...row.tags]
  tagDialogVisible.value = true
}

const confirmAddTag = () => {
  if (currentRow.value) {
    currentRow.value.tags = [...selectedTags.value, ...customTags.value]
    ElMessage.success('标签添加成功')
  }
  tagDialogVisible.value = false
  customTag.value = ''
  customTags.value = []
}

const addCustomTag = () => {
  if (customTag.value && !customTags.value.includes(customTag.value)) {
    customTags.value.push(customTag.value)
    customTag.value = ''
  }
}

const removeCustomTag = (index: number) => {
  customTags.value.splice(index, 1)
}

const openLevelDialog = (row: any) => {
  levelRow.value = row
  selectedLevel.value = row.level
  levelDialogVisible.value = true
}

const confirmLevelChange = () => {
  if (levelRow.value && selectedLevel.value) {
    levelRow.value.level = selectedLevel.value
    ElMessage.success('等级调整成功')
  }
  levelDialogVisible.value = false
}

const toggleStatus = (row: any) => {
  const action = row.status === 'active' ? '暂停' : '启用'
  ElMessageBox.confirm(
    `是否确认要将 ${row.phone} 的会员${action}会员服务？`,
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    if (row.status === 'active') {
      row.status = 'paused'
      ElMessage.success('会员已暂停')
    } else {
      row.status = 'active'
      ElMessage.success('会员已启用')
    }
  }).catch(() => {})
}
</script>

<style scoped>
.member-page {
  padding: 20px;
}

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.search-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.points-range {
  display: flex;
  align-items: center;
}

.member-info {
  display: flex;
  flex-direction: column;
}

.member-name {
  font-weight: 500;
  color: #1a1a1a;
}

.member-phone {
  font-size: 12px;
  color: #999;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
