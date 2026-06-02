// pages/my/order/logistics/index.js
const app = getApp();

Page({
  /**
   * 页面的初始数据
   */
  data: {
    // 加载状态
    loading: true,
    // 订单ID
    orderId: '',
    // 物流信息
    logisticsInfo: null,
    // 错误信息
    errorMessage: ''
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 从URL参数中获取订单ID
    if (options.orderId) {
      this.setData({
        orderId: options.orderId
      });
      // 请求物流信息
      this.fetchLogisticsInfo();
    } else {
      this.setData({
        loading: false,
        errorMessage: '订单ID不存在'
      });
    }
  },

  /**
   * 请求物流信息
   */
  fetchLogisticsInfo() {
    if (!this.data.orderId) {
      this.setData({
        loading: false,
        errorMessage: '订单ID不存在'
      });
      return;
    }

    this.setData({ loading: true, errorMessage: '' });
    
    // 准备请求参数
    const requestData = {
      order_id: this.data.orderId
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起物流信息请求:', '/order/sync_logistics_info', requestData);
    app.req.post('/order/sync_logistics_info', requestData, 
      (res) => {
        console.log('物流信息响应:', res);
        // 处理成功响应
        if (res && res.status === 'success' && res.data) {
          const logisticsData = res.data;
          
          // 构造物流信息对象
          // 使用reverse()方法反转物流进程顺序，让最新的信息显示在顶部
          const logisticsInfo = {
            orderId: logisticsData.order_id,
            expressCompany: logisticsData.express_company || '',
            expressNumber: logisticsData.express_number || '',
            logisticsProcess: (logisticsData.logistics_process || []).reverse()
          };
          
          this.setData({
            logisticsInfo: logisticsInfo,
            loading: false
          });
        } else {
          // 请求成功但数据格式不正确
          this.setData({
            loading: false,
            errorMessage: res && res.message ? res.message : '获取物流信息失败'
          });
        }
      },
      (err) => {
        // 处理请求失败
        console.error('请求物流信息失败:', err);
        this.setData({
          loading: false,
          errorMessage: '网络请求失败'
        });
      }
    );
  },

  /**
   * 返回上一页
   */
  navigateBack() {
    wx.navigateBack({
      delta: 1
    });
  },

  /**
   * 重试请求
   */
  retryRequest() {
    this.fetchLogisticsInfo();
  }
});