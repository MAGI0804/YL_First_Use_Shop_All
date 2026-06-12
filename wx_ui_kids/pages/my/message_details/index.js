// pages/my/message_details/index.js
const app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {
    // 消息列表数据
    messageList: [],
    // 消息类型
    messageType: '',
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
    console.log('页面加载，收到的参数:', options);
    if (options.type) {
      this.setData({ messageType: options.type });
      this.loadMessages(true);
    } else {
      wx.showToast({
        title: '消息类型不存在',
        icon: 'none'
      });
    }
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    if (this.data.messageList.length === 0 && this.data.messageType) {
      this.loadMessages(true);
    }
  },

  /**
   * 加载消息数据
   */
  loadMessages(isRefresh = false) {
    if (this.data.loading || (!this.data.hasMore && !isRefresh)) {
      return;
    }

    this.setData({ loading: true });
    
    // 实时获取用户ID
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 774231);
    
    // 准备请求参数
    const { page, pageSize, messageType } = this.data;
    const requestData = {
      user_id: userId,
      message_type: messageType,
      page: isRefresh ? 1 : page,
      page_size: pageSize
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起消息查询请求:', '/message/query', requestData);
    app.req.post('/message/query', requestData, 
      (res) => {
        console.log('消息查询响应:', res);
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 将API返回的数据转换为页面需要的数据格式
          const messages = res.data.data.map(item => ({
            id: item.MessageID,
            title: item.MessageTitleOne,
            subtitle: item.MessageTitleTwo,
            content: item.MessageBody,
            time: item.CreatedAt,
            type: item.MessageType,
            relatedNum: item.RelatedNum,
            displayImg: item.DisplayImg
          }));
          
          // 更新导航栏标题
          if (messages.length > 0) {
            const messageType = messages[0].type;
            let navTitle = '消息详情';
            switch (messageType) {
              case 'Order':
                navTitle = '订单消息';
                break;
              case 'return_order':
                navTitle = '退货消息';
                break;
              case 'Promotion':
                navTitle = '促销消息';
                break;
              case 'System':
                navTitle = '系统消息';
                break;
              case 'Notice':
                navTitle = '通知消息';
                break;
            }
            wx.setNavigationBarTitle({ title: navTitle });
          }
          
          // 更新页面数据
          this.setData({
            messageList: isRefresh ? messages : [...this.data.messageList, ...messages],
            loading: false,
            page: isRefresh ? 2 : page + 1,
            hasMore: messages.length === pageSize
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
    console.log('下拉刷新消息列表');
    this.loadMessages(true);
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {
    console.log('上拉加载更多消息');
    if (!this.data.loading && this.data.hasMore) {
      this.loadMessages(false);
    }
  },

  /**
   * 处理消息点击事件
   */
  handleMessageTap(e) {
    const message = e.currentTarget.dataset.message;
    console.log('点击消息:', message);
    
    // 检查是否有RelatedNum
    if (!message.relatedNum) {
      console.log('消息没有RelatedNum，不进行跳转');
      return;
    }
    
    // 根据消息类型跳转到不同页面
    switch (message.type) {
      case 'Order':
        // 跳转到订单详情页
        console.log('跳转到订单详情页，订单号:', message.relatedNum);
        app.navigateTo({
          url: `/pages/my/order/detail/index?order_no=${message.relatedNum}`
        });
        break;
      case 'return_order':
        // 跳转到售后详情页
        console.log('跳转到售后详情页，售后单号:', message.relatedNum);
        app.navigateTo({
          url: `/pages/my/order/return_detail/index?return_order_no=${message.relatedNum}`
        });
        break;
      default:
        console.log('未知消息类型，不进行跳转:', message.type);
        break;
    }
  }
})
