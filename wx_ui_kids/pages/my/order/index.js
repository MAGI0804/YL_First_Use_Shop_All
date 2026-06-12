// pages/my/order/index.js
const app = getApp();
const globalUserInfo = app.globalData.userInfo || {};
const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);

const statusTextMap = {
  pending: '待发货',
  shipped: '已发货',
  delivered: '已签收',
  canceled: '已取消',
  processing: '售后中'
};

const payStatusTextMap = {
  unpaid: '未支付',
  paid: '已支付',
  partial_paid: '部分支付'
};

function formatMoney(value) {
  return Number(value || 0).toFixed(2);
}

function mapOrderItem(item) {
  const payStatus = item.pay_status || 'unpaid';
  const orderAmount = Number(item.order_amount || 0);
  const finalPayAmount = Number(item.final_pay_amount || item.order_amount || 0);

  return {
    id: item.order_id,
    status: item.status,
    statusText: statusTextMap[item.status] || item.status,
    payStatus,
    payStatusText: payStatusTextMap[payStatus] || '未支付',
    createTime: item.order_time,
    totalPrice: formatMoney(orderAmount),
    finalPayAmount: formatMoney(finalPayAmount),
    productCount: Array.isArray(item.product_list) ? item.product_list.length : 0,
    process_num: item.process_num || item.return_order_id || '',
    products: Array.isArray(item.product_list) ? item.product_list.map((productId) => ({
      id: productId.toString(),
      name: '加载中...',
      price: '0.00',
      quantity: 1,
      image: '/images/products.png'
    })) : []
  };
}

Page({
  /**
   * 页面的初始数据
   */
  data: {
    // 订单状态列表
    statusList: [
      { key: 'all', name: '全部' },
      { key: 'pending', name: '待发货' },
      { key: 'shipped', name: '已发货' },
      { key: 'delivered', name: '已签收' },
      { key: 'canceled', name: '已取消' },
      { key: 'processing', name: '售后中' }
    ],
    // 当前选中的状态
    currentStatus: 'all',
    // 订单列表数据
    orderList: [],
    // 页面加载状态
    loading: false,
    // 是否有更多数据
    hasMore: true,
    // 当前页码
    page: 1,
    // 每页数量
    pageSize: 10,
    // 搜索关键字
    searchKeyword: '',
    // 是否处于搜索状态
    isSearching: false
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 从URL参数中获取状态，如果没有则默认为全部
    if (options.status) {
      // 处理从主页传来的中文状态名
      const statusMap = {
        '全部': 'all',
        '待发货': 'pending',
        '已发货': 'shipped',
        '已签收': 'delivered',
        '已送达': 'delivered',
        '已取消': 'canceled',
        '售后中': 'processing'
      };
      
      this.setData({
        currentStatus: statusMap[options.status] || options.status || 'all'
      });
    }
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 每次显示页面时重新加载订单数据
    this.loadOrders(true);
  },

  /**
   * 切换订单状态
   */
  switchStatus(e) {
    const status = e.currentTarget.dataset.status;
    if (status !== this.data.currentStatus) {
      this.setData({
        currentStatus: status,
        page: 1,
        hasMore: true,
        orderList: [],
        isSearching: false,
        searchKeyword: ''
      });
      this.loadOrders(true);
    }
  },

  /**
   * 搜索输入变化
   */
  onSearchInput(e) {
    this.setData({
      searchKeyword: e.detail.value
    });
  },

  /**
   * 清空搜索
   */
  clearSearch() {
    this.setData({
      searchKeyword: '',
      isSearching: false,
      page: 1,
      hasMore: true,
      orderList: []
    });
    this.loadOrders(true);
  },

  /**
   * 搜索确认
   */
  onSearchConfirm() {
    const { searchKeyword } = this.data;
    if (searchKeyword.trim()) {
      this.setData({
        isSearching: true,
        page: 1,
        hasMore: true,
        orderList: []
      });
      this.loadSearchOrders(true);
    }
  },

  /**
   * 加载搜索订单数据
   */
  loadSearchOrders(isRefresh = false) {
    if (this.data.loading || (!this.data.hasMore && !isRefresh)) {
      return;
    }

    this.setData({ loading: true });
    
    // 实时获取用户ID
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    // 准备请求参数
    const { searchKeyword, page, pageSize } = this.data;
    const requestData = {
      shopname: 'youlan_kids',
      user_id: userId,
      product_name: searchKeyword,
      page: isRefresh ? 1 : page,
      page_size: pageSize
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起订单搜索请求:', '/order/search_by_product_name', requestData);
    app.req.post('/order/search_by_product_name', requestData, 
      (res) => {
        console.log('订单搜索响应:', res);
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 将API返回的数据转换为页面需要的数据格式
          const orders = res.data.data.map(mapOrderItem);
          
          // 提取所有商品ID
          const allProductIds = [];
          orders.forEach(order => {
            if (Array.isArray(order.products)) {
              order.products.forEach(product => {
                if (product.id && !allProductIds.includes(product.id)) {
                  allProductIds.push(product.id);
                }
              });
            }
          });
          
          // 如果有商品ID需要查询，先不更新页面数据，等待获取到商品详情后再更新
          if (allProductIds.length > 0) {
            this.getProductDetails(allProductIds, orders, isRefresh);
          } else {
            // 没有商品ID需要查询，直接更新页面数据
            this.setData({
              orderList: isRefresh ? orders : [...this.data.orderList, ...orders],
              loading: false,
              page: isRefresh ? 2 : page + 1,
              hasMore: orders.length === pageSize
            });
            // 停止下拉刷新
            if (isRefresh) {
              wx.stopPullDownRefresh();
            }
          }
        } else {
          // 请求成功但数据格式不正确
          this.setData({ loading: false });
          wx.showToast({
            title: '获取订单数据失败',
            icon: 'none'
          });
        }
        

      },
      
      (err) => {
        // 处理请求失败
        console.error('请求订单数据失败:', err);
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
   * 加载订单数据
   */
  loadOrders(isRefresh = false) {
    if (this.data.loading || (!this.data.hasMore && !isRefresh)) {
      return;
    }

    this.setData({ loading: true });
    
    // 实时获取用户ID
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    // 准备请求参数
    const { currentStatus, page, pageSize } = this.data;
    const requestData = {
      page: isRefresh ? 1 : page,
      page_size: pageSize,
      shopname: 'youlan_kids',
      user_id: userId
    };
    
    // 如果不是全部订单，添加状态参数
    if (currentStatus !== 'all') {
      requestData.status = currentStatus;
    }
    
    // 使用app.js中的req.post方法调用API
    console.log('发起订单查询请求:', '/order/query_by_user_id', requestData);
    app.req.post('/order/query_by_user_id', requestData, 
      (res) => {
        console.log('订单查询响应:', res);
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 将API返回的数据转换为页面需要的数据格式
          const orders = res.data.data.map(mapOrderItem);
          
          // 提取所有商品ID
          const allProductIds = [];
          orders.forEach(order => {
            if (Array.isArray(order.products)) {
              order.products.forEach(product => {
                if (product.id && !allProductIds.includes(product.id)) {
                  allProductIds.push(product.id);
                }
              });
            }
          });
          
          // 如果有商品ID需要查询，先不更新页面数据，等待获取到商品详情后再更新
          if (allProductIds.length > 0) {
            this.getProductDetails(allProductIds, orders, isRefresh);
          } else {
            // 没有商品ID需要查询，直接更新页面数据
            this.setData({
              orderList: isRefresh ? orders : [...this.data.orderList, ...orders],
              loading: false,
              page: isRefresh ? 2 : page + 1,
              hasMore: orders.length === pageSize
            });
            // 停止下拉刷新
            if (isRefresh) {
              wx.stopPullDownRefresh();
            }
          }
        } else {
          // 请求成功但数据格式不正确
          this.setData({ loading: false });
          wx.showToast({
            title: '获取订单数据失败',
            icon: 'none'
          });
        }
        

      },
      
      (err) => {
        // 处理请求失败
        console.error('请求订单数据失败:', err);
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
   * 根据商品ID列表获取商品详情
   */
  getProductDetails(productIds, orders, isRefresh) {
    const requestData = {
      commodity_ids: productIds
    };
    
    app.req.post('/commodity/batch_get_products_by_ids', requestData, 
      (res) => {
        let updatedOrders = orders;
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 创建商品信息映射表
          const productMap = {};
          res.data.data.forEach(product => {
            productMap[product.commodity_id] = product;
          });
          
          // 更新订单中的商品信息，合并相同商品并计算数量
          updatedOrders = orders.map(order => {
            // 先统计每种商品的数量
            const productCountMap = {};
            order.products.forEach(product => {
              if (!productCountMap[product.id]) {
                productCountMap[product.id] = 0;
              }
              productCountMap[product.id] += product.quantity;
            });
            
            // 创建合并后的商品列表
            const mergedProducts = [];
            Object.keys(productCountMap).forEach(productId => {
              const realProduct = productMap[productId];
              mergedProducts.push({
                id: productId,
                name: realProduct ? (realProduct.name || '商品名称') : '商品名称',
                price: realProduct && realProduct.price ? parseFloat(realProduct.price).toFixed(2) : '0.00',
                quantity: productCountMap[productId],
                image: realProduct ? (realProduct.image || realProduct.promo_image_url || '/images/products.png') : '/images/products.png'
              });
            });
            
            return {
              ...order,
              products: mergedProducts,
              productCount: Object.values(productCountMap).reduce((sum, count) => sum + count, 0) // 重新计算商品总数
            };
          });
        } else {
          // 请求商品详情失败，显示提示
          wx.showToast({
            title: '获取商品信息失败',
            icon: 'none'
          });
          
          // 在失败情况下也尝试合并相同商品
          updatedOrders = orders.map(order => {
            const productCountMap = {};
            order.products.forEach(product => {
              if (!productCountMap[product.id]) {
                productCountMap[product.id] = 0;
              }
              productCountMap[product.id] += product.quantity;
            });
            
            const mergedProducts = [];
            Object.keys(productCountMap).forEach(productId => {
              mergedProducts.push({
                id: productId,
                name: '加载中...',
                price: '0.00',
                quantity: productCountMap[productId],
                image: '/images/products.png'
              });
            });
            
            return {
              ...order,
              products: mergedProducts,
              productCount: Object.values(productCountMap).reduce((sum, count) => sum + count, 0)
            };
          });
        }
        
        // 更新页面数据
        this.setData({
          orderList: isRefresh ? updatedOrders : [...this.data.orderList, ...updatedOrders],
          loading: false,
          page: isRefresh ? 2 : this.data.page + 1,
          hasMore: updatedOrders.length === this.data.pageSize
        });
        
        // 停止下拉刷新
        if (isRefresh) {
          wx.stopPullDownRefresh();
        }
      },
      (err) => {
        // 处理请求失败
        console.error('请求商品详情失败:', err);
        
        // 更新页面数据，即使商品详情获取失败，也显示订单信息
        this.setData({
          orderList: isRefresh ? orders : [...this.data.orderList, ...orders],
          loading: false,
          page: isRefresh ? 2 : this.data.page + 1,
          hasMore: orders.length === this.data.pageSize
        });
        
        // 停止下拉刷新
        if (isRefresh) {
          wx.stopPullDownRefresh();
        }
        
        // 显示错误提示
        wx.showToast({
          title: '获取商品信息失败',
          icon: 'none'
        });
      }
    );
  },

  

  /**
   * 查看订单详情
   */
  viewOrderDetail(e) {
    const orderId = e.currentTarget.dataset.id;
    const orderStatus = e.currentTarget.dataset.status;
    const processNum = e.currentTarget.dataset.processNum;
    
    console.log('查看订单详情，订单ID:', orderId, '状态:', orderStatus, '退货订单号:', processNum);
    
    // 如果订单状态为售后中，跳转到售后订单详情页面
    if (orderStatus === 'processing' && processNum) {
      console.log('订单状态为售后中，跳转到售后订单详情页面');
      app.navigateTo({
        url: `/pages/my/order/return_detail/index?returnOrderId=${processNum}&orderId=${orderId}`
      });
    } else {
      console.log('订单状态不是售后中，跳转到普通订单详情页面');
      // 其他状态跳转到普通订单详情页面
      app.navigateTo({
        url: `/pages/my/order/detail/index?id=${orderId}`
      });
    }
  },

  /**
   * 跳转到首页
   */
  navigateToHome() {
    app.switchTab('/pages/index/index');
  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {
    if (this.data.isSearching) {
      this.loadSearchOrders(true);
    } else {
      this.loadOrders(true);
    }
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {
    console.log('触发上拉加载，当前状态:', {
      loading: this.data.loading,
      hasMore: this.data.hasMore,
      currentPage: this.data.page,
      orderListLength: this.data.orderList.length,
      isSearching: this.data.isSearching
    });
    // 只有当不是加载中且有更多数据时才加载
    if (!this.data.loading && this.data.hasMore) {
      console.log('开始加载下一页数据');
      if (this.data.isSearching) {
        this.loadSearchOrders(false);
      } else {
        this.loadOrders(false);
      }
    } else {
      console.log('不需要加载数据:', {
        loading: this.data.loading,
        hasMore: this.data.hasMore
      });
    }
  },

  /**
   * 取消订单
   */
  cancelOrder(e) {
    const orderId = e.currentTarget.dataset.id;
    const that = this;
    
    wx.showModal({
      title: '取消订单',
      content: '确定要取消该订单吗？',
      success(res) {
        if (res.confirm) {
          // 显示加载中提示
          wx.showLoading({
            title: '处理中...',
          });
          console.log('发起订单查询请求:', '/order/cancel', { order_id: orderId,user_id: userId, });
          // 发送取消订单请求
          app.req.post('/order/cancel', { order_id: orderId ,user_id: userId,},
            (res) => {
              wx.hideLoading();
              if (res && res.code === 200 ) {
                wx.showToast({
                  title: res.data.data.message || '订单已取消',
                  icon: 'success'
                });
                
                // 刷新订单列表
                that.loadOrders(true);
              } else {
                wx.showToast({
                  title: res && res.data && res.data.data && res.data.data.message ? res.data.data.message : '取消订单失败',
                  icon: 'none'
                });
              }
            },
            (err) => {
              wx.hideLoading();
              console.error('取消订单失败:', err);
              wx.showToast({
                title: '网络请求失败',
                icon: 'none'
              });
            }
          );
        }
      }
    });
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
    })
  },
  
  /**
   * 申请售后
   */
  navigateToAfterSales(e) {
    const orderId = e.currentTarget.dataset.id;
    const orderStatus = e.currentTarget.dataset.status;
    // 跳转到售后申请页面
    console.log('申请售后，订单ID:', orderId, '订单状态:', orderStatus);
    app.navigateTo({
      url: `/pages/my/order/return/index?id=${orderId}&order_status=${orderStatus}`
    });
  },
  
  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {
    return {
      title: '我的订单',
      path: '/pages/my/order/index'
    };
  }
})
