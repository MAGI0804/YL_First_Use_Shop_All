<template>
  <div :class="pickerClass">
    <VueDatePicker
      :model-value="startValue"
      class="compact-date-input"
      :model-type="modelType"
      :format="formatInput"
      :enable-time-picker="isTimeRange"
      :enable-seconds="isTimeRange"
      :start-time="isTimeRange ? defaultTime : undefined"
      :auto-apply="!isTimeRange"
      placeholder="开始"
      locale="zh-CN"
      position="left"
      teleport="body"
      select-text="确定"
      cancel-text="取消"
      :text-input="textInputConfig"
      clearable
      @open="refreshDefaultTime"
      @update:model-value="value => updateRange(0, value)"
    />
    <span class="compact-date-separator">-</span>
    <VueDatePicker
      :model-value="endValue"
      class="compact-date-input"
      :model-type="modelType"
      :format="formatInput"
      :enable-time-picker="isTimeRange"
      :enable-seconds="isTimeRange"
      :start-time="isTimeRange ? defaultTime : undefined"
      :auto-apply="!isTimeRange"
      placeholder="结束"
      locale="zh-CN"
      position="left"
      teleport="body"
      select-text="确定"
      cancel-text="取消"
      :text-input="textInputConfig"
      clearable
      @open="refreshDefaultTime"
      @update:model-value="value => updateRange(1, value)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import VueDatePicker from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

type DateRangeValue = [string, string] | null

const props = withDefaults(defineProps<{
  modelValue: DateRangeValue
  type?: 'daterange' | 'datetimerange'
  valueFormat?: string
}>(), {
  type: 'daterange',
  valueFormat: 'YYYY-MM-DD'
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: DateRangeValue): void
}>()

const isTimeRange = computed(() => props.type === 'datetimerange')
const startValue = computed(() => props.modelValue?.[0] || null)
const endValue = computed(() => props.modelValue?.[1] || null)

const getCurrentTime = () => {
  const now = new Date()
  return {
    hours: now.getHours(),
    minutes: now.getMinutes(),
    seconds: now.getSeconds()
  }
}

const defaultTime = ref(getCurrentTime())

const refreshDefaultTime = () => {
  defaultTime.value = getCurrentTime()
}

const modelType = computed(() => {
  return props.valueFormat
    .replace('YYYY', 'yyyy')
    .replace('DD', 'dd')
})

const textInputConfig = computed(() => ({
  enterSubmit: true,
  tabSubmit: true,
  selectOnFocus: true,
  format: parseTextInput
}))

const formatDatePart = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatTimePart = (date: Date) => {
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${formatDatePart(date)} ${hours}:${minutes}:${seconds}`
}

const formatInput = (date: Date) => {
  if (!(date instanceof Date)) return ''
  return isTimeRange.value ? formatTimePart(date) : formatDatePart(date)
}

const parseTextInput = (value: string) => {
  const pattern = isTimeRange.value
    ? /^(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})$/
    : /^(\d{4})-(\d{2})-(\d{2})$/
  const match = value.trim().match(pattern)
  if (!match) return null

  const [, year, month, day, hours = '0', minutes = '0', seconds = '0'] = match
  const date = new Date(
    Number(year),
    Number(month) - 1,
    Number(day),
    Number(hours),
    Number(minutes),
    Number(seconds)
  )

  return Number.isNaN(date.getTime()) ? null : date
}

const updateRange = (index: 0 | 1, value: unknown) => {
  const next: [string, string] = [props.modelValue?.[0] || '', props.modelValue?.[1] || '']
  next[index] = typeof value === 'string' ? value : ''
  emit('update:modelValue', next[0] || next[1] ? next : null)
}

const pickerClass = computed(() => [
  'compact-date-range',
  props.type === 'datetimerange' ? 'compact-date-range--time' : ''
])
</script>

<style scoped>
.compact-date-range {
  width: 292px;
  flex: 0 0 292px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 12px minmax(0, 1fr);
  align-items: center;
  gap: 6px;
}

.compact-date-range--time {
  width: 440px;
  flex-basis: 440px;
}

.compact-date-input,
:deep(.dp__main) {
  width: 100%;
  min-width: 0;
}

.compact-date-separator {
  color: #909399;
  text-align: center;
}

:deep(.dp__input) {
  height: 34px;
  border-color: #e4e7ed;
  border-radius: 0;
  color: #303133;
  font-size: 14px;
  padding: 6px 26px 6px 28px;
}

:deep(.dp__input:hover),
:deep(.dp__input_focus) {
  border-color: #87ceeb;
}

:deep(.dp__input_icon) {
  color: #909399;
  width: 14px;
  height: 14px;
}

</style>
