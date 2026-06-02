// pages/my/index/index.js
const app = getApp()

Page({

  /**
   * 页面的初始数据
   */
  data: {
    isLogin: false,
    userInfo: null
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // console.log('页面加载，开始检查登录状态')
    // 检查登录状态
    this.checkLoginStatus()
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // console.log('页面显示，开始检查登录状态')
    // 每次页面显示时检查登录状态
    this.checkLoginStatus()
  },

  /**
   * 检查登录状态
   */
  checkLoginStatus() {
    // 直接从全局状态获取用户信息
    const globalUserInfo = app.globalData.userInfo;
    const globalUserId = app.globalData.user_id;
    
    // 初始化用户数据
    this.setData({
      isLogin: false,
      userInfo: {
        nickname: '点击登录',
        user_img: '/images/home.png'
      }
    });
    
    // 检查是否有有效的全局用户信息
    let userId = '';
    if (globalUserInfo && globalUserInfo.user_id) {
      userId = globalUserInfo.user_id;
    } else if (globalUserId) {
      userId = globalUserId;
    } else {
      // 尝试从本地存储恢复用户信息
      const storedUserId = wx.getStorageSync('user_id');
      if (storedUserId) {
        userId = storedUserId;
      } else {
        // 尝试从本地存储恢复完整用户信息
        const storedUserInfo = wx.getStorageSync('userInfo');
        if (storedUserInfo && storedUserInfo.user_id) {
          userId = storedUserInfo.user_id;
        }
      }
    }
    
    if (userId) {
      this.setData({
        isLogin: true,
        userInfo: {
          user_id: userId,
          nickname: '加载中...',
          user_img: '/images/home.png'
        }
      });
      
      // 异步请求最新的用户数据
      this.fetchUserData(userId);
    } else {
      // 未登录状态，使用默认值
      console.log('用户未登录');
    }
  },
  
  /**
   * 获取用户数据
   */
  fetchUserData(user_id) {
    const that = this;
    
    app.req.post('/ordinary_user/find_data', {
      user_id: user_id
    }, function(res) {
      // 成功回调
      // console.log('获取用户数据成功:', res);
      
      // 统一处理响应结果
      // 根据新的API返回格式，数据在res.data中
      if (res && res.code === 200 && res.data && res.data.user_id) {
        // 格式化用户数据，确保字段名与WXML模板匹配
        const formattedUserInfo = {
          user_id: res.data.user_id,
          nickname: res.data.nickname || '微信用户',
          user_img: res.data.user_img || '/images/home.png',
          openid: res.data.openid,
          mobile: res.data.mobile,
          membership_level: res.data.membership_level,
          registration_date: res.data.registration_date,
          total_spending: res.data.total_spending || '0.00',
          last_login: res.data.last_login,
          is_active: res.data.is_active
        };
        
        // 更新页面数据
        that.setData({
          userInfo: formattedUserInfo
        });
        
        // 更新全局和本地存储
        app.globalData.userInfo = formattedUserInfo;
        wx.setStorageSync('userInfo', formattedUserInfo);
      } else if (res && res.code === 201 && res.msg && res.msg.includes('用户不存在')) {
        // 处理用户不存在的情况
        console.log('用户不存在，清除所有用户信息');
        that.clearUserInfo();
      } else {
        // 其他数据问题，使用默认值
        console.warn('返回数据格式异常，使用默认值');
        that.setData({
          userInfo: {
            user_id: user_id,
            nickname: '微信用户',
            user_img: '/images/home.png'
          }
        });
      }
    }, function(err) {
      // 错误回调
      console.error('获取用户数据失败:', err);
      
      // 检查是否是用户不存在的错误
      if (err && (err.error === '用户不存在' || (err.message && err.message.includes('用户不存在')))) {
        that.clearUserInfo();
      } else {
        // 其他错误，保持已有状态
        console.log('网络或服务器错误，保持已有状态');
      }
    });
  },

  /**
   * 跳转到登录页面
   */
  navigateToLogin() {
    // 清空所有缓存，确保登录前状态干净
    console.log('登录前清空所有缓存');
    this.clearUserInfo();
    
    wx.navigateTo({
      url: '/pages/accUser/index'
    })
  },

  /**
   * 跳转到编辑资料页面
   */
  navigateToEditProfile() {
    wx.navigateTo({
      url: '/pages/my/modify/index'
    })
  },

  /**
   * 跳转到全部订单页面
   */
  navigateToAllOrders() {
    wx.navigateTo({
      url: '/pages/my/order/index?status=all'
    })
  },
  
  /**
   * 跳转到特定状态的订单页面
   */
  navigateToOrderByStatus(e) {
    const status = e.currentTarget.dataset.status;
    wx.navigateTo({
      url: `/pages/my/order/index?status=${status}`
    })
  },

  /**
   * 跳转到地址管理页面
   */
  navigateToAddress() {
    wx.navigateTo({
      url: '/pages/my/address/index'
    })
  },

  /**
   * 跳转到消息页面
   */
  navigateToMessage() {
    wx.navigateTo({
      url: '/pages/my/message/index'
    })
  },

  /**
   * 跳转到联系客服页面
   */
  navigateToCustomerService() {
    wx.openCustomerServiceChat({
      success: function(res) {
        console.log('进入客服聊天成功', res)
      },
      fail: function(res) {
        console.error('进入客服聊天失败', res)
      }
    })
  },
  
  /**
   * 清除用户信息
   */
  clearUserInfo() {
    wx.clearStorageSync();
    app.globalData.userInfo = null;
    app.globalData.user_id = null;
    this.setData({
      isLogin: false,
      userInfo: {
        nickname: '点击登录',
        user_img: '/images/home.png'
      }
    });
  },

  /**
   * 退出登录
   */
  handleLogout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          this.clearUserInfo();
          wx.showToast({
            title: '已退出登录',
            icon: 'success'
          });
          setTimeout(() => {
            wx.reLaunch({
              url: '/pages/index/index'
            });
          }, 1500);
        }
      }
    });
  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {
    this.checkLoginStatus()
    wx.stopPullDownRefresh()
  }
})