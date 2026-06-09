// pages/accUser/index.js
const app = getApp()
const req = app.req
const request = require('../../api/request').default

Page({
  data: {
    bar: {
      hide: true
    },
    agreementChecked: false,
    avatarUrl: '',
    nickname: '',
    loginLoading: false
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
    if (this.data.loginLoading) {
      return
    }
    if (!this.data.agreementChecked) {
      wx.showToast({
        title: '请先同意协议',
        icon: 'none'
      })
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
      success: (loginRes) => {
        if (!loginRes.code) {
          this.setData({ loginLoading: false })
          wx.showToast({
            title: '获取登录凭证失败',
            icon: 'none'
          })
          return
        }
        this.loginWithWechatCode(loginRes.code, e.detail.code)
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

  loginWithWechatCode(code, phoneCode) {
    const avatarUrl = this.isTemporaryAvatar(this.data.avatarUrl) ? '' : this.data.avatarUrl
    req.post('/ordinary_user/wechat_login', {
      code,
      phone_code: phoneCode,
      userInfo: {
        nickName: this.data.nickname,
        nickname: this.data.nickname,
        avatarUrl,
        avatar_url: avatarUrl
      }
    }, (res) => {
      this.setData({ loginLoading: false })
      if (res.code === 200 && res.data) {
        this.persistLogin(res.data)
        this.uploadAvatarIfNeeded(res.data.user_id, () => {
          this.redirectAfterLogin()
        })
        return
      }
      wx.showToast({
        title: res.message || res.error || '登录失败',
        icon: 'none'
      })
    }, (err) => {
      this.setData({ loginLoading: false })
      const data = err && err.data ? err.data : {}
      wx.showToast({
        title: data.message || data.error || err.message || '登录失败',
        icon: 'none'
      })
      console.error('登录失败:', err)
    })
  },

  persistLogin(data) {
    const token = data.token || {}
    const userInfo = {
      user_id: data.user_id,
      member_no: data.member_no,
      mobile: data.mobile,
      nickname: data.nickname || this.data.nickname || '微信用户',
      user_img: data.avatar_url || this.data.avatarUrl || '/images/home.png',
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

  uploadAvatarIfNeeded(userId, done) {
    if (!this.isTemporaryAvatar(this.data.avatarUrl) || !userId) {
      done()
      return
    }

    const accessToken = wx.getStorageSync('access_token')
    wx.uploadFile({
      url: `${request.getHost()}/ordinary_user/Modify_data?access_token=${encodeURIComponent(accessToken)}`,
      filePath: this.data.avatarUrl,
      name: 'user_img',
      formData: {
        user_id: userId.toString(),
        nickname: this.data.nickname
      },
      success: (res) => {
        try {
          const data = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
          const avatarURL = data && (data.avatar_url || (data.data && data.data.avatar_url))
          if (avatarURL) {
            const userInfo = {
              ...app.globalData.userInfo,
              user_img: avatarURL
            }
            app.globalData.userInfo = userInfo
            wx.setStorageSync('userInfo', userInfo)
          }
        } catch (parseErr) {
          console.warn('头像上传响应解析失败:', parseErr)
        }
      },
      fail: (err) => {
        console.warn('头像上传失败，保留本地头像:', err)
      },
      complete: done
    })
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
