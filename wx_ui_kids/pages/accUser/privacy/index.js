// pages/accUser/privacy/index.js
const app = getApp()

Page({
  /**
   * 页面的初始数据
   */
  data: {
    type: 'privacy' // 默认显示隐私政策，可以是 'privacy' 或 'agreement'
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 从URL参数中获取要显示的协议类型
    if (options && options.type) {
      this.setData({
        type: options.type
      });
    }
    // 根据协议类型设置页面标题
    const title = this.data.type === 'privacy' ? '隐私政策' : '用户协议';
    wx.setNavigationBarTitle({ title });
  },

  /**
   * 返回上一页
   */
  navigateBack() {
    app.navigateBack();
  }
});
