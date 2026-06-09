// pages/accUser/index.js
const app = getApp()
const req = app.req
const request = require('../../api/request').default

const LOGIN_REQUEST_TIMEOUT = 20000
const WX_LOGIN_TIMEOUT = 8000

Page({
  data: {
    bar: {
      hide: true
    },
    agreementChecked: false,
    phoneAuthed: false,
    avatarUrl: '',
    nickname: '',
    loginLoading: false,
    profileSaving: false
  },

  toggleAgreement() {
    this.setData({
      agreementChecked: !this.data.agreementChecked
    })
  },

  navigateToAgreement() {
    wx.navigateTo({
      url: '/pages/accUser/privacy/index?type=agreement'
    })
  },

  navigateToPrivacy() {
    wx.navigateTo({
      url: '/pages/accUser/privacy/index?type=privacy'
    })
  },

  onChooseAvatar(e) {
    const avatarUrl = e.detail && e.detail.avatarUrl ? e.detail.avatarUrl : ''
    if (avatarUrl) {
      this.setData({ avatarUrl })
    }
  },

  onNicknameInput(e) {
    this.setData({
      nickname: (e.detail.value || '').trim()
    })
  },

  promptAgreement() {
    wx.showToast({
      title: '请先同意协议',
      icon: 'none'
    })
  },

  handlePhoneLogin(e) {
    const requestId = Date.now().toString()
    console.log('[wechat-login] getPhoneNumber event', {
      requestId,
      errMsg: e.detail && e.detail.errMsg,
      hasPhoneCode: !!(e.detail && e.detail.code)
    })

    if (this.data.loginLoading) {
      return
    }
    if (!this.data.agreementChecked) {
      this.promptAgreement()
      return
    }
    if (!e.detail || e.detail.errMsg !== 'getPhoneNumber:ok' || !e.detail.code) {
      wx.showToast({
        title: '需要授权手机号登录',
        icon: 'none'
      })
      return
    }

    this.setData({ loginLoading: true })
    wx.login({
      timeout: WX_LOGIN_TIMEOUT,
      success: (loginRes) => {
        console.log('[wechat-login] wx.login success', {
          requestId,
          hasCode: !!loginRes.code
        })
        if (!loginRes.code) {
          this.setData({ loginLoading: false })
          wx.showToast({
            title: '获取登录凭证失败',
            icon: 'none'
          })
          return
        }
        this.loginWithWechatCode(loginRes.code, e.detail.code, requestId)
      },
      fail: (err) => {
        this.setData({ loginLoading: false })
        wx.showToast({
          title: '微信登录失败',
          icon: 'none'
        })
        console.error('wx.login失败:', err)
      }
    })
  },

  loginWithWechatCode(code, phoneCode, requestId) {
    console.log('[wechat-login] request start', {
      requestId,
      host: request.getHost()
    })
    wx.request({
      url: `${request.getHost()}/ordinary_user/wechat_login`,
      method: 'POST',
      timeout: LOGIN_REQUEST_TIMEOUT,
      header: {
        'content-type': 'application/json'
      },
      data: {
        code,
        phone_code: phoneCode,
        userInfo: {}
      },
      success: (httpRes) => {
        const res = httpRes.data || {}
        console.log('[wechat-login] request success', {
          requestId,
          statusCode: httpRes.statusCode,
          code: res.code,
          message: res.message || res.error || ''
        })
        if (httpRes.statusCode >= 200 && httpRes.statusCode < 300 && res.code === 200 && res.data) {
          this.persistLogin(res.data)
          this.setData({
            phoneAuthed: true,
            nickname: res.data.nickname || '',
            avatarUrl: res.data.avatar_url || ''
          })
          return
        }
        wx.showToast({
          title: res.message || res.error || `登录失败(${httpRes.statusCode})`,
          icon: 'none'
        })
      },
      fail: (err) => {
        const isTimeout = err && err.errMsg && err.errMsg.includes('timeout')
        wx.showToast({
          title: isTimeout ? '登录超时，请重试' : '网络错误，登录失败',
          icon: 'none'
        })
        console.error('[wechat-login] request fail:', {
          requestId,
          err
        })
      },
      complete: () => {
        this.setData({ loginLoading: false })
      }
    })
  },

  persistLogin(data) {
    const token = data.token || {}
    const userInfo = {
      user_id: data.user_id,
      member_no: data.member_no,
      mobile: data.mobile,
      nickname: data.nickname || '微信用户',
      user_img: data.avatar_url || '/images/home.png',
      phone_bound: data.phone_bound === true
    }

    wx.setStorageSync('token', token.access)
    wx.setStorageSync('refresh_token', token.refresh)
    wx.setStorageSync('user_id', data.user_id)
    wx.setStorageSync('userInfo', userInfo)

    app.globalData.userInfo = userInfo
    app.globalData.token = token.access
    app.globalData.user_id = data.user_id
  },

  saveProfileAndEnter() {
    if (this.data.profileSaving) {
      return
    }
    const userId = wx.getStorageSync('user_id')
    if (!userId) {
      wx.showToast({
        title: '请先授权手机号',
        icon: 'none'
      })
      return
    }

    this.setData({ profileSaving: true })
    if (this.isTemporaryAvatar(this.data.avatarUrl)) {
      this.uploadAvatarProfile(userId)
      return
    }
    this.updateNicknameProfile(userId)
  },

  uploadAvatarProfile(userId) {
    const accessToken = wx.getStorageSync('access_token')
    wx.uploadFile({
      url: `${request.getHost()}/ordinary_user/Modify_data?access_token=${encodeURIComponent(accessToken)}`,
      filePath: this.data.avatarUrl,
      name: 'user_img',
      formData: {
        user_id: userId.toString(),
        nickname: this.data.nickname || '微信用户'
      },
      success: (res) => {
        try {
          const data = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
          const avatarURL = data && (data.avatar_url || (data.data && data.data.avatar_url))
          this.mergeStoredUserInfo({
            nickname: this.data.nickname || app.globalData.userInfo.nickname,
            user_img: avatarURL || this.data.avatarUrl || app.globalData.userInfo.user_img
          })
        } catch (parseErr) {
          console.warn('头像上传响应解析失败:', parseErr)
          this.mergeStoredUserInfo({
            nickname: this.data.nickname || app.globalData.userInfo.nickname
          })
        }
      },
      fail: (err) => {
        console.warn('头像上传失败:', err)
        this.mergeStoredUserInfo({
          nickname: this.data.nickname || app.globalData.userInfo.nickname
        })
      },
      complete: () => {
        this.setData({ profileSaving: false })
        this.redirectAfterLogin()
      }
    })
  },

  updateNicknameProfile(userId) {
    const nickname = this.data.nickname
    if (!nickname || nickname === (app.globalData.userInfo && app.globalData.userInfo.nickname)) {
      this.setData({ profileSaving: false })
      this.redirectAfterLogin()
      return
    }
    req.post('/ordinary_user/Modify_data', {
      user_id: userId,
      nickname
    }, () => {
      this.mergeStoredUserInfo({ nickname })
      this.setData({ profileSaving: false })
      this.redirectAfterLogin()
    }, (err) => {
      console.warn('昵称保存失败:', err)
      this.setData({ profileSaving: false })
      this.redirectAfterLogin()
    })
  },

  skipProfile() {
    this.redirectAfterLogin()
  },

  mergeStoredUserInfo(nextInfo) {
    const current = app.globalData.userInfo || wx.getStorageSync('userInfo') || {}
    const userInfo = {
      ...current,
      ...nextInfo
    }
    app.globalData.userInfo = userInfo
    wx.setStorageSync('userInfo', userInfo)
  },

  isTemporaryAvatar(avatarUrl) {
    return !!avatarUrl && (avatarUrl.startsWith('wxfile://') || avatarUrl.startsWith('http://tmp/'))
  },

  redirectAfterLogin() {
    if (this.returnPath === 'goods' && this.id) {
      wx.redirectTo({
        url: `/pages/commodity/goods/index?id=${this.id}`,
        fail: () => {
          wx.switchTab({ url: '/pages/index/index' })
        }
      })
      return
    }
    wx.switchTab({
      url: '/pages/index/index'
    })
  },

  onLoad(options) {
    this.returnPath = options.returnPath
    this.id = options.id
    wx.hideHomeButton()
  },

  onReady() {
    wx.hideHomeButton()
  },

  onShow() {
    wx.hideHomeButton()
  }
})
