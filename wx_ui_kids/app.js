// app.js
import request from './api/request'
import router from './utils/router'
App({
  bindGetUserInfo (e) {
    console.log(e.detail.userInfo)
  },
  onLaunch(option) {
    const normalScenes = [1001, 1005, 1101, 1011, 1012, 1013, 1089]
    this.globalData.share = !normalScenes.includes(option.scene)
    
    try {
      const systemInfo = wx.getSystemInfoSync()
      const barTop = systemInfo.statusBarHeight || 0
      const menuButtonInfo = wx.getMenuButtonBoundingClientRect()
      const barHeight = menuButtonInfo && menuButtonInfo.height
        ? menuButtonInfo.height + (menuButtonInfo.top - barTop) * 2
        : 44

      this.globalData.barTop = barTop
      this.globalData.barHeight = barHeight
      this.globalData.placeHolderHeight = barHeight + barTop
      this.globalData.height = barTop + 40
      this.globalData.windowWidth = systemInfo.windowWidth || 375
    } catch (err) {
      console.warn('获取系统信息失败，使用默认导航高度', err)
      this.globalData.barTop = 0
      this.globalData.barHeight = 44
      this.globalData.placeHolderHeight = 44
      this.globalData.height = 44
      this.globalData.windowWidth = 375
    }

    const userInfo = wx.getStorageSync('userInfo')
    if(userInfo && (userInfo.avatar || userInfo.user_img || userInfo.user_id)){
      this.globalData.userInfo = userInfo
      this.globalData.user_id = userInfo.user_id || wx.getStorageSync('user_id') || null
    }
  },
  globalData: {
    userInfo: null,
    height: 0,
    barTop: 0,
    barHeight: 0,
    placeHolderHeight: 0,
    code: undefined,
    token: undefined ,
    user_id: null,
    share: false
  },
  req:{
    get: (url, data, response, error) => request.http('GET', url, data, response, error),
    post: (url, data, response, error) => request.http('POST', url, data, response, error),
    put: (url, data, response, error) => request.http('PUT', url, data, response, error),
    delete: (url, data, response, error) => request.http('DELETE', url, data, response, error),
    getHost: () => request.getHost(),
    getAccessToken: (success, fail) => request.getAccessToken(success, fail),
  },
  router,
  navigateTo: router.navigateTo,
  switchTab: router.switchTab,
  redirectTo: router.redirectTo,
  navigateBack: router.navigateBack,
  parseTime(time, cFormat) {
    if (arguments.length === 0) {
      return null
    }
    const format = cFormat || '{y}-{m}-{d} {h}:{i}:{s}'
    let date
    if (typeof time === 'object') {
      date = time
    } else {
      if ((typeof time === 'string') && (/^[0-9]+$/.test(time))) {
        time = parseInt(time)
      }
      if ((typeof time === 'number') && (time.toString().length === 10)) {
        time = time * 1000
      }
      date = new Date(time)
    }
    const formatObj = {
      y: date.getFullYear(),
      m: date.getMonth() + 1,
      d: date.getDate(),
      h: date.getHours(),
      i: date.getMinutes(),
      s: date.getSeconds(),
      a: date.getDay()
    }
    return format.replace(/{([ymdhisa])+}/g, (result, key) => {
      const value = formatObj[key]
      // Note: getDay() returns 0 on Sunday
      if (key === 'a') {
        return ['日', '一', '二', '三', '四', '五', '六'][value]
      }
      return value.toString().padStart(2, '0')
    })
  },
  haveUserInfo(){
    let info = wx.getStorageSync('userInfo')
    if(wx.getStorageSync('token')!='' || wx.getStorageSync('user_id')!=''){
      return !info || !info.avatar && !info.user_img
    } else {
      return false
    }
  }
})
