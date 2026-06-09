<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  width: 100%;
  height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

#app {
  width: 100%;
  height: 100%;
}

.el-main {
  padding: 0 !important;
}

.el-card__body {
  padding: 16px;
}

.el-button,
.el-input__wrapper,
.el-select__wrapper,
.el-card,
.el-dialog,
.el-message-box,
.el-dropdown-menu,
.el-table,
.el-form-item__content,
.el-pagination {
  border-radius: 0 !important;
  box-shadow: none !important;
}

.el-card {
  border: 1px solid #e4e7ed;
}

.el-button--primary {
  background: #87ceeb;
  border-color: #87ceeb;
}

.el-button--primary:hover {
  background: #5bc0de;
  border-color: #5bc0de;
}

.el-input__wrapper {
  box-shadow: none !important;
  border: 1px solid #e4e7ed;
}

.el-input__wrapper:hover,
.el-input__wrapper.is-focus {
  border-color: #87ceeb;
}

.dp__theme_light {
  --dp-primary-color: #5bc0de;
  --dp-border-color-hover: #87ceeb;
  --dp-border-radius: 0;
  --dp-font-family: inherit;
}

.dp__menu {
  border-color: #e5e7eb;
  border-radius: 0;
  box-shadow: none;
  z-index: 3000;
}

.dp__outer_menu_wrap {
  z-index: 3000 !important;
}

.dp__preset_dates {
  width: 72px;
  padding: 8px 0;
  border-right-color: #edf0f4;
}

.dp__preset_date {
  padding: 7px 10px;
  font-size: 12px;
}

.dp__calendar_header,
.dp__calendar {
  font-size: 12px;
}

.dp__action_row {
  padding: 8px 12px;
}
</style>
