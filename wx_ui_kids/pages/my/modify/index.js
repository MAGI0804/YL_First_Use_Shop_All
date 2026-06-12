// pages/my/modify/index.js
const app = getApp()
const request = require('../../../api/request').default

const getCurrentUser = () => {
  return {
    ...(app.globalData.userInfo || {}),
    ...(wx.getStorageSync('userInfo') || {})
  }
}

const getCurrentUserId = () => {
  const userInfo = getCurrentUser()
  return userInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || null
}

const isTemporaryAvatar = (avatarUrl) => {
  return !!avatarUrl && (avatarUrl.startsWith('wxfile://') || avatarUrl.startsWith('http://tmp/'))
}

Page({
  data: {
    userInfo: {
      nickName: '',
      avatarUrl: '',
      mobile: ''
    },
    originalUserInfo: {
      nickName: '',
      avatarUrl: '',
      mobile: ''
    },
    wechatNicknameDraft: '',
    saving: false,
    nav: {
      title: '修改个人信息',
      showHome: true
    }
  },

  onLoad() {
    this.initUserInfo()
  },

  onShow() {
    if (!this.data.originalUserInfo.nickName && !this.data.originalUserInfo.avatarUrl) {
      this.initUserInfo()
    }
  },

  initUserInfo() {
    const userInfo = getCurrentUser()
    const nextUserInfo = {
      nickName: userInfo.nickName || userInfo.nickname || '',
      avatarUrl: userInfo.user_img || userInfo.avatarUrl || userInfo.avatar || '',
      mobile: userInfo.mobile || ''
    }

    this.setData({
      userInfo: { ...nextUserInfo },
      originalUserInfo: { ...nextUserInfo },
      wechatNicknameDraft: ''
    })
  },

  uploadAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        const tempFile = res.tempFiles && res.tempFiles[0]
        if (!tempFile || !tempFile.tempFilePath) {
          return
        }
        if (tempFile.size > 5 * 1024 * 1024) {
          wx.showToast({
            title: '图片大小不能超过5MB',
            icon: 'none'
          })
          return
        }
        this.setData({
          'userInfo.avatarUrl': tempFile.tempFilePath
        })
      },
      fail: (err) => {
        console.error('选择图片失败:', err)
      }
    })
  },

  onChooseWechatAvatar(e) {
    const avatarUrl = e.detail && e.detail.avatarUrl ? e.detail.avatarUrl : ''
    if (avatarUrl) {
      this.setData({
        'userInfo.avatarUrl': avatarUrl
      })
    }
  },

  onWechatNicknameInput(e) {
    const nickname = (e.detail.value || '').trim()
    if (nickname) {
      this.setData({
        'userInfo.nickName': nickname,
        wechatNicknameDraft: nickname
      })
    }
  },

  onNickNameInput(e) {
    this.setData({
      'userInfo.nickName': e.detail.value
    })
  },

  onMobileInput(e) {
    this.setData({
      'userInfo.mobile': e.detail.value
    })
  },

  saveUserInfo() {
    if (this.data.saving) {
      return
    }

    const { nickName, avatarUrl } = this.data.userInfo
    const userId = getCurrentUserId()

    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }
    if (!nickName || nickName.trim() === '') {
      wx.showToast({
        title: '昵称不能为空',
        icon: 'none'
      })
      return
    }

    this.setData({ saving: true })
    wx.showLoading({
      title: '保存中...',
      mask: true
    })

    if (isTemporaryAvatar(avatarUrl)) {
      this.uploadProfileWithAvatar(userId, nickName.trim(), avatarUrl)
      return
    }
    this.saveProfileWithoutAvatar(userId, nickName.trim())
  },

  getAccessToken(callback) {
    const accessToken = wx.getStorageSync('access_token')
    if (accessToken) {
      callback(accessToken)
      return
    }
    request.getAccessToken((token) => {
      callback(token)
    }, () => {
      this.finishSaving()
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
    })
  },

  uploadProfileWithAvatar(userId, nickName, avatarUrl) {
    this.getAccessToken((accessToken) => {
      wx.uploadFile({
        url: `${request.getHost()}/ordinary_user/Modify_data?access_token=${encodeURIComponent(accessToken)}`,
        filePath: avatarUrl,
        name: 'user_img',
        formData: {
          user_id: userId.toString(),
          nickname: nickName
        },
        success: (res) => {
          this.handleSaveResult(res)
        },
        fail: (err) => {
          console.error('保存用户信息失败:', err)
          wx.showToast({
            title: '保存失败',
            icon: 'none'
          })
        },
        complete: () => {
          this.finishSaving()
        }
      })
    })
  },

  saveProfileWithoutAvatar(userId, nickName) {
    app.req.post('/ordinary_user/Modify_data', {
      user_id: userId,
      nickname: nickName
    }, (res) => {
      this.finishSaving()
      this.handleSaveResult(res)
    }, (err) => {
      this.finishSaving()
      console.error('保存用户信息失败:', err)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    })
  },

  finishSaving() {
    wx.hideLoading()
    this.setData({ saving: false })
  },

  parseSaveResponse(res) {
    if (res && typeof res.data === 'string') {
      return JSON.parse(res.data)
    }
    if (res && (res.code !== undefined || res.msg || res.message || res.error)) {
      return res
    }
    if (res && res.statusCode && res.data) {
      return res.data
    }
    return res && res.data ? res.data : res
  },

  isSaveSuccess(data) {
    return data && (
      data.code === 200 ||
      data.message === '用户信息更新成功' ||
      data.message === '信息修改成功' ||
      data.msg === '信息修改成功'
    )
  },

  updateStoredUserInfo(data) {
    const current = getCurrentUser()
    const avatarUrl = data.avatar_url ||
      data.user_img ||
      (data.data && (data.data.avatar_url || data.data.user_img)) ||
      this.data.userInfo.avatarUrl
    const nickname = this.data.userInfo.nickName
    const updatedUserInfo = {
      ...current,
      nickName: nickname,
      nickname,
      avatarUrl,
      user_img: avatarUrl
    }

    app.globalData.userInfo = updatedUserInfo
    wx.setStorageSync('userInfo', updatedUserInfo)
    this.setData({
      originalUserInfo: {
        nickName: updatedUserInfo.nickname || '',
        avatarUrl: updatedUserInfo.user_img || '',
        mobile: updatedUserInfo.mobile || ''
      },
      wechatNicknameDraft: ''
    })
  },

  handleSaveResult(res) {
    try {
      const data = this.parseSaveResponse(res)
      if (this.isSaveSuccess(data)) {
        this.updateStoredUserInfo(data)
        wx.showToast({
          title: '保存成功',
          icon: 'success',
          duration: 2000,
          success: () => {
            setTimeout(() => {
              app.navigateBack()
            }, 1500)
          }
        })
        return
      }
      wx.showToast({
        title: (data && (data.message || data.msg || data.error)) || '保存失败',
        icon: 'none'
      })
    } catch (e) {
      console.error('保存用户信息失败:', e)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    }
  }
})
