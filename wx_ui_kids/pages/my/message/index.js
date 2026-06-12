// pages/my/message/index.js
const app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {
    // 消息列表数据
    messageList: [],
    // 加载状态
    loading: false,
    // 是否有更多数据
    hasMore: true,
    // 当前页码
    page: 1,
    // 每页数量
    pageSize: 10
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 页面加载时的初始化操作
    this.loadMessages(true);
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 页面显示时的操作
    if (this.data.messageList.length === 0) {
      this.loadMessages(true);
    }
  },

  /**
   * 加载消息分类数据
   */
  loadMessages(isRefresh = false) {
    if (this.data.loading) {
      return;
    }

    this.setData({ loading: true });
    
    // 实时获取用户ID
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 774231);
    
    // 准备请求参数
    const requestData = {
      user_id: userId
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起消息分类查询请求:', '/message/categories', requestData);
    app.req.post('/message/categories', requestData, 
      (res) => {
        console.log('消息分类查询响应:', res);
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 将API返回的数据转换为页面需要的数据格式
          const messages = res.data.data.map((item, index) => {
            // 根据消息类型设置图标和标题
            const typeConfig = {
              'Order': { icon: '📦', title: '订单消息' },
              'Promotion': { icon: '🔥', title: '促销消息' },
              'System': { icon: '📢', title: '系统消息' },
              'Notice': { icon: '⏰', title: '通知消息' },
              'return_order': { icon: '🔄', title: '退货消息' }
            };
            
            const config = typeConfig[item.message_type] || { icon: '📢', title: '其他消息' };
            
            return {
              id: index + 1,
              title: config.title,
              description: item.last_message,
              time: item.last_message_time,
              type: item.message_type,
              icon: config.icon,
              unread: true
            };
          });
          
          // 更新页面数据
          this.setData({
            messageList: messages,
            loading: false,
            hasMore: false
          });
          
          // 停止下拉刷新
          if (isRefresh) {
            wx.stopPullDownRefresh();
          }
        } else {
          // 请求成功但数据格式不正确
          this.setData({ loading: false });
          wx.showToast({
            title: '获取消息数据失败',
            icon: 'none'
          });
        }
      },
      (err) => {
        // 处理请求失败
        console.error('请求消息数据失败:', err);
        this.setData({ loading: false });
        wx.showToast({
          title: '网络请求失败',
          icon: 'none'
        });
        
        // 停止下拉刷新
        if (isRefresh) {
          wx.stopPullDownRefresh();
        }
      }
    );
  },

  /**
   * 查看消息详情
   */
  viewMessageDetail(e) {
    const messageId = e.currentTarget.dataset.id;
    const message = this.data.messageList.find(item => item.id === messageId);
    console.log('查看消息详情，消息ID:', messageId, '消息类型:', message.type);
    
    // 标记消息为已读
    const messageList = this.data.messageList.map(item => {
      if (item.id === messageId) {
        return { ...item, unread: false };
      }
      return item;
    });
    this.setData({ messageList });
    
    // 跳转到消息详情页面
    if (message) {
      app.navigateTo({
        url: `/pages/my/message_details/index?type=${message.type}`
      });
    }
  },

  /**
   * 联系客服
   */
  contactService() {
    wx.openCustomerServiceChat({
      success: function(res) {
        console.log('进入客服聊天成功', res)
      },
      fail: function(res) {
        console.error('进入客服聊天失败', res)
      }
    });
  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {
    // 下拉刷新，重新加载消息列表
    console.log('下拉刷新消息列表');
    this.loadMessages(true);
  },



  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {
    return {
      title: '我的消息',
      path: '/pages/my/message/index'
    };
  }
});
