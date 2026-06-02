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
</style>
