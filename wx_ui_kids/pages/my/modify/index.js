// pages/my/modify/index.js
const app = getApp()
const request = require('../../../api/request').default

Page({
  /**
   * 页面的初始数据
   */
  data: {
    userInfo: {
      nickName: '',
      avatarUrl: '',
      mobile: ''
    },
    nav: {
      title: '修改个人信息',
      showHome: true
    }
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 初始化用户信息
    this.initUserInfo()
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 每次显示页面时刷新用户信息
    this.initUserInfo()
  },

  /**
   * 初始化用户信息
   */
  initUserInfo() {
    // 从本地存储获取用户信息，确保能获取到完整的用户数据
    const storedUserInfo = wx.getStorageSync('userInfo') || {}
    // 结合全局数据，确保信息完整
    const globalUserInfo = app.globalData.userInfo || {}
    // 合并用户信息，优先使用本地存储的数据
    const userInfo = {
      ...globalUserInfo,
      ...storedUserInfo
    }
    
    // 确保用户信息正确显示，根据my/index/index.js中的定义，头像是user_img字段
    this.setData({
      userInfo: {
        nickName: userInfo.nickName || userInfo.nickname || '',
        avatarUrl: userInfo.user_img || userInfo.avatarUrl || userInfo.avatar || '',
        mobile: userInfo.mobile || ''
      }
    })
    
    console.log('初始化用户信息:', this.data.userInfo)
    console.log('原始用户数据(包含user_img):', userInfo)
  },

  /**
   * 上传头像
   */
  uploadAvatar() {
    const that = this
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'], // 压缩图片
      success(res) {
        const tempFilePath = res.tempFiles[0].tempFilePath
        const fileSize = res.tempFiles[0].size
        
        // 检查文件大小，限制在5MB以内
        if (fileSize > 5 * 1024 * 1024) {
          wx.showToast({
            title: '图片大小不能超过5MB',
            icon: 'none'
          })
          return
        }
        
        // 获取access_token
        const accessToken = wx.getStorageSync('access_token')
        if (!accessToken) {
          wx.showToast({
            title: '请先登录',
            icon: 'none'
          })
          return
        }
        
        // 获取用户ID
        const user_id = app.globalData.userInfo.user_id || 
                      (wx.getStorageSync('userInfo') ? wx.getStorageSync('userInfo').user_id : null)
        
        if (!user_id) {
          wx.showToast({
            title: '用户信息不完整',
            icon: 'none'
          })
          return
        }
        
        // 显示加载提示
        wx.showLoading({
          title: '上传中...',
        })
        
        // 使用wx.uploadFile上传图片，自动使用multipart/form-data格式
        wx.uploadFile({
          url: `${request.getHost()}/ordinary_user/Modify_data?access_token=${accessToken}`,
          filePath: tempFilePath,
          name: 'user_img', // 根据后端要求，文件参数名必须是user_img
          formData: {
            user_id: user_id.toString() // 后端需要user_id参数
          },
          header: {
            'content-type': 'multipart/form-data'
          },
          success(res) {
            try {
              const data = JSON.parse(res.data)
              if (data.message === '用户信息更新成功') {
                // 上传成功，设置头像URL
                that.setData({
                  'userInfo.avatarUrl': tempFilePath // 临时使用本地路径，后续会在保存时更新
                })
                wx.showToast({
                  title: '上传成功',
                  icon: 'success'
                })
              } else {
                wx.showToast({
                  title: data.error || data.message || '上传失败',
                  icon: 'none'
                })
              }
            } catch (e) {
              console.error('解析上传结果失败:', e)
              wx.showToast({
                title: '上传失败，请重试',
                icon: 'none'
              })
            }
          },
          fail(err) {
            console.error('上传图片失败:', err)
            wx.showToast({
              title: '网络异常，请重试',
              icon: 'none'
            })
          },
          complete() {
            wx.hideLoading()
          }
        })
      },
      fail(err) {
        console.error('选择图片失败:', err)
      }
    })
  },

  /**
   * 监听昵称输入
   */
  onNickNameInput(e) {
    this.setData({
      'userInfo.nickName': e.detail.value
    })
  },

  /**
   * 监听手机号输入
   */
  onMobileInput(e) {
    this.setData({
      'userInfo.mobile': e.detail.value
    })
  },

  /**
   * 保存用户信息
   */
  saveUserInfo() {
    const { nickName, avatarUrl } = this.data.userInfo
    // 从全局或本地存储获取用户ID
    const user_id = app.globalData.userInfo.user_id || 
                  (wx.getStorageSync('userInfo') ? wx.getStorageSync('userInfo').user_id : null)

    if (!user_id) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    // 验证昵称不能为空
    if (!nickName || nickName.trim() === '') {
      wx.showToast({
        title: '昵称不能为空',
        icon: 'none'
      })
      return
    }

    // 获取本地存储的access_token
    const accessToken = wx.getStorageSync('access_token')
    
    if (!accessToken) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    // 检查avatarUrl是否为临时文件路径（包含wxfile://前缀）
    const isTempFilePath = avatarUrl && avatarUrl.startsWith('wxfile://')
    
    if (isTempFilePath) {
      // 当用户选择了新头像时，使用wx.uploadFile
      wx.uploadFile({
        url: `${request.getHost()}/ordinary_user/Modify_data?access_token=${accessToken}`,
        filePath: avatarUrl,
        name: 'user_img',
        formData: {
          user_id: user_id.toString(),
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
        }
      })
    } else {
      app.req.post('/ordinary_user/Modify_data', {
        user_id: user_id,
        nickname: nickName
      }, (res) => {
        this.handleSaveResult(res)
      }, (err) => {
        console.error('保存用户信息失败:', err)
        wx.showToast({
          title: '保存失败',
          icon: 'none'
        })
      })
    }
  },
  
  /**
   * 处理保存结果
   */
  handleSaveResult(res) {
    try {
      const data = typeof res.data === 'string' ? JSON.parse(res.data) : res
      if (data.code === 200 || data.message === '用户信息更新成功' || data.msg === '信息修改成功') {
        // 更新成功，更新全局和本地存储的用户信息
        const updatedUserInfo = {
          ...app.globalData.userInfo,
          nickName: this.data.userInfo.nickName,
          avatarUrl: data.avatar_url || this.data.userInfo.avatarUrl
        }
        app.globalData.userInfo = updatedUserInfo
        wx.setStorageSync('userInfo', updatedUserInfo)

        wx.showToast({
          title: '保存成功',
          icon: 'success',
          duration: 2000,
          success: () => {
            // 延迟跳转回用户页面
            setTimeout(() => {
              wx.navigateBack()
            }, 1500)
          }
        })
      } else {
        wx.showToast({
          title: data.message || '保存失败',
          icon: 'none'
        })
      }
    } catch (e) {
      console.error('保存用户信息失败:', e)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    }
  }
})
