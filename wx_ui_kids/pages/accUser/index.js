// pages/accUser/index.js
const app = getApp()
const req = app.req
Page({
  /**
   * 页面的初始数据
   */
  data: {
    bar:{
      hide: true
    },
    // 协议勾选状态
    agreementChecked: false
  },

  /**
   * 切换协议勾选状态
   */
  toggleAgreement() {
    this.setData({
      agreementChecked: !this.data.agreementChecked
    });
  },
  
  /**
   * 跳转到用户协议页面
   */
  navigateToAgreement() {
    wx.navigateTo({
      url: '/pages/accUser/privacy/index?type=agreement'
    });
  },
  
  /**
   * 跳转到隐私政策页面
   */
  navigateToPrivacy() {
    wx.navigateTo({
      url: '/pages/accUser/privacy/index?type=privacy'
    });
  },
  // 一键登录
  getUserProfile() {
    // 检查协议是否已勾选
    if (!this.data.agreementChecked) {
      wx.showToast({
        title: '请先阅读并同意用户协议和隐私政策',
        icon: 'none'
      });
      return;
    }
    
    // 1. 获取登录code
    wx.login({
      success: (loginRes) => {
        if (loginRes.code) {
          console.log("code: loginRes.code")
          // 2. 调用微信登录接口
          req.post('/ordinary_user/wechat_login', {
            code: loginRes.code
          }, (res) => {
            if (res.code === 200) {
              // 登录成功
              const { token, user_id } = res.data;
              
              // 存储token和user_id
              wx.setStorageSync('token', token.access);
              wx.setStorageSync('refresh_token', token.refresh);
              wx.setStorageSync('user_id', user_id);
              
              // 更新全局用户信息
              app.globalData.userInfo = { user_id };
              app.globalData.token = token.access;
              app.globalData.user_id = user_id;
              
              // 重定向到首页
              wx.switchTab({
                url: '/pages/index/index'
              });
            } else {
              wx.showToast({
                title: res.message || '登录失败',
                icon: 'none'
              });
            }
          }, (err) => {
            wx.showToast({
              title: '网络错误',
              icon: 'none'
            });
            console.error('登录失败:', err);
          });
        } else {
          wx.showToast({
            title: '获取code失败',
            icon: 'none'
          });
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '登录失败',
          icon: 'none'
        });
        console.error('wx.login失败:', err);
      }
    });
  },
  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    wx.hideHomeButton()
  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {
    wx.hideHomeButton()
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    wx.hideHomeButton()
  },

  /**
   * 生命周期函数--监听页面隐藏
   */
  onHide() {

  },

  /**
   * 生命周期函数--监听页面卸载
   */
  onUnload() {

  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {

  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {

  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {

  }
})