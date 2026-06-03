// pages/my/order/detail/index.js
const app = getApp();

function getUserId() {
  const globalUserInfo = app.globalData.userInfo || {};
  return globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || '';
}

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

Page({
  /**
   * 页面的初始数据
   */
  data: {
    // 订单详情数据
    orderDetail: null,
    // 加载状态
    loading: true,
    // 订单ID
    orderId: '',
    // 来源页面
    from: ''
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 从URL参数中获取订单ID和来源页面
    const orderId = options.id || options.order_no;
    if (orderId) {
      this.setData({
        orderId: orderId,
        from: options.from || ''
      });
      // 加载订单详情
      this.loadOrderDetail();
    }
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 如果已经加载过订单详情，再次显示时可以选择是否刷新
    if (this.data.orderDetail && !this.data.loading) {
      // 这里可以根据实际需求决定是否刷新订单详情
    }
  },

  /**
   * 加载订单详情
   */
  loadOrderDetail() {
    if (this.data.loading) {
      this.setData({ loading: true });
    }
    
    // 准备请求参数
    const requestData = {
      order_id: this.data.orderId,
      inquired_list: ["order_id", "order_amount", "final_pay_amount", "discount_amount", "discount_reason", "pay_status", "payment_time", "payment_remark", "product_list", "province", "city", "county", "detailed_address", "status", "remarks", "order_time", "receiver_name", "receiver_phone", "express_company", "express_number", "shipped_time", "delivered_time", "canceled_time", "process_num", "processing_time"]
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起订单详情请求:', '/order/query_order_data', requestData);
    app.req.post('/order/query_order_data', requestData, 
      (res) => {
        console.log('订单详情响应:', res);
        // 处理成功响应
        if (res && res.code === 200 && res.data) {
          // 将API返回的数据转换为页面需要的数据格式
          const orderData = res.data.data || res.data;
          
          const payStatus = orderData.pay_status || 'unpaid';
          
          // 解析商品列表（兼容字符串和数组格式）
          let productList = [];
          try {
            // 检查订单数据中是否有详细的商品列表信息
            if (orderData.product_details && Array.isArray(orderData.product_details)) {
              // 如果订单数据中直接包含详细的商品信息，直接使用
              productList = orderData.product_details.map((product) => ({
                id: (product.product_id || product.commodity_id || '').toString(),
                name: product.product_name || product.name || '加载中...',
                price: product.price ? parseFloat(product.price).toFixed(2) : '0.00',
                quantity: product.quantity || 1,
                subtotal: product.subtotal ? parseFloat(product.subtotal).toFixed(2) : (product.price ? parseFloat(product.price).toFixed(2) : '0.00'),
                image: product.image || product.promo_image_url || '/images/products.png'
              }));
            } else {
              // 否则使用product_list数组
              const products = typeof orderData.product_list === 'string' ? JSON.parse(orderData.product_list) : orderData.product_list;
              productList = products.map((productId) => ({
                id: productId.toString(),
                name: '加载中...',
                price: '0.00',
                quantity: 1,
                subtotal: '0.00',
                image: '/images/home.png'
              }));
            }
          } catch (e) {
            console.error('解析商品列表失败:', e);
          }
          
          // 构造收货地址信息 - 尝试从订单数据中获取真实的收货人信息
          const address = {
            name: orderData.receiver_name || orderData.name || '张三',
            phone: orderData.receiver_phone || orderData.phone || '138****1234',
            address: `${orderData.province || ''}${orderData.city || ''}${orderData.county || ''}${orderData.detailed_address || ''}`
          };
          
          // 构造物流信息 - 尝试从订单数据中获取真实的物流信息
          const orderStatus = orderData.status;
          // 检查是否有真实的物流信息（快递公司和运单号）
          const hasRealLogisticsInfo = !!(orderData.express_company && orderData.express_number);
          
          const logisticsInfo = {
            company: orderData.express_company || orderData.delivery_company || '',
            trackingNumber: orderData.express_number || orderData.logistics_no || '',
            status: orderStatus !== 'pending' && hasRealLogisticsInfo ? '运输中' : '暂无物流信息',
            updates: []
          };
          
          // 构造完整的订单详情数据，根据订单状态设置对应的时间
          const orderDetail = {
            id: orderData.order_id,
            status: orderData.status,
            statusText: statusTextMap[orderData.status] || orderData.status,
            payStatus,
            payStatusText: payStatusTextMap[payStatus] || '未支付',
            // 下单时间
            orderTime: orderData.order_time || new Date().toLocaleString(),
            // 发货时间
            shippedTime: orderData.shipped_time || '',
            // 完成时间
            deliveredTime: orderData.delivered_time || '',
            // 取消时间
            canceledTime: orderData.canceled_time || '',
            // 售后发起时间
            processingTime: orderData.processing_time || '',
            // 售后订单号
            process_num: orderData.process_num || '',
            totalPrice: formatMoney(orderData.order_amount),
            finalPayAmount: formatMoney(orderData.final_pay_amount || orderData.order_amount),
            discountAmount: formatMoney(orderData.discount_amount),
            discountReason: orderData.discount_reason || '',
            paymentTime: orderData.payment_time || '',
            paymentRemark: orderData.payment_remark || '',
            productCount: productList.length,
            products: productList,
            address: address,
            logistics: logisticsInfo,
            paymentMethod: payStatus === 'paid' ? '线下结算' : '收货后线下结算',
            orderNumber: orderData.order_id,
            remarks: orderData.remarks || ''
          };
          
          this.setData({
            orderDetail: orderDetail,
            loading: false
          });
          
          // 提取所有商品ID
          const allProductIds = [];
          if (Array.isArray(productList)) {
            productList.forEach(product => {
              if (product.id && !allProductIds.includes(product.id)) {
                allProductIds.push(product.id);
              }
            });
          }
          
          // 调用商品详情API获取真实商品信息
          if (allProductIds.length > 0) {
            this.getProductDetails(allProductIds);
          } else {
            this.loadSubOrders();
          }
          
          // 不再自动处理物流信息，改为点击跳转查看
        } else {
          // 请求成功但数据格式不正确
          this.setData({ loading: false });
          wx.showToast({
            title: '获取订单详情失败',
            icon: 'none'
          });
        }
      },
      (err) => {
        // 处理请求失败
        console.error('请求订单详情失败:', err);
        this.setData({ loading: false });
        wx.showToast({
          title: '网络请求失败',
          icon: 'none'
        });
      }
    );
  },
  
  /**
   * 根据商品ID列表获取商品详情
   */
  getProductDetails(productIds) {
    const requestData = {
      commodity_ids: productIds
    };
    console.log(requestData)
    
    app.req.post('/commodity/batch_get_products_by_ids', requestData, 
      (res) => {
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 创建商品信息映射表
          const productMap = {};
          res.data.data.forEach(product => {
            productMap[product.commodity_id] = product;
          });
          
          // 获取当前订单详情
          const currentOrderDetail = this.data.orderDetail;
          if (currentOrderDetail && Array.isArray(currentOrderDetail.products)) {
            // 先统计每种商品的数量
            const productCountMap = {};
            const productPriceMap = {}; // 存储商品单价用于计算小计
            
            currentOrderDetail.products.forEach(product => {
              if (!productCountMap[product.id]) {
                productCountMap[product.id] = 0;
              }
              productCountMap[product.id] += product.quantity || 1;
              
              // 记录商品单价
              const realProduct = productMap[product.id];
              const unitPrice = realProduct && realProduct.price ? parseFloat(realProduct.price) : (product.price ? parseFloat(product.price) : 0);
              productPriceMap[product.id] = unitPrice;
            });
            
            // 创建合并后的商品列表
            const mergedProducts = [];
            Object.keys(productCountMap).forEach(productId => {
              const realProduct = productMap[productId];
              const product = currentOrderDetail.products.find(p => p.id === productId);
              
              if (product) {
                const unitPrice = productPriceMap[productId];
                const quantity = productCountMap[productId];
                const subtotal = unitPrice * quantity;
                
                mergedProducts.push({
                  id: productId,
                  name: realProduct ? (realProduct.name || product.name) : product.name || '商品名称',
                  price: unitPrice.toFixed(2),
                  quantity: quantity,
                  subtotal: subtotal.toFixed(2),
                  image: realProduct ? (realProduct.image || realProduct.promo_image_url || product.image) : (product.image || '/images/products.png'),
                  styleCode: realProduct ? (realProduct.style_code || product.styleCode) : (product.styleCode || ''),
                  category: realProduct ? (realProduct.category || product.category) : (product.category || '')
                });
              }
            });
            
            // 更新订单详情
            this.setData({
              orderDetail: {
                ...currentOrderDetail,
                products: mergedProducts,
                productCount: Object.values(productCountMap).reduce((sum, count) => sum + count, 0) // 重新计算商品总数
              }
            });
            this.loadSubOrders();
          }
        } else {
          // 请求商品详情失败，在失败情况下也尝试合并相同商品
          const currentOrderDetail = this.data.orderDetail;
          if (currentOrderDetail && Array.isArray(currentOrderDetail.products)) {
            const productCountMap = {};
            currentOrderDetail.products.forEach(product => {
              if (!productCountMap[product.id]) {
                productCountMap[product.id] = 0;
              }
              productCountMap[product.id] += product.quantity || 1;
            });
            
            const mergedProducts = [];
            Object.keys(productCountMap).forEach(productId => {
              const product = currentOrderDetail.products.find(p => p.id === productId);
              if (product) {
                mergedProducts.push({
                  ...product,
                  quantity: productCountMap[product.id]
                });
              }
            });
            
            this.setData({
              orderDetail: {
                ...currentOrderDetail,
                products: mergedProducts,
                productCount: Object.values(productCountMap).reduce((sum, count) => sum + count, 0)
              }
            });
            this.loadSubOrders();
          }
          
          // 显示提示
          wx.showToast({
            title: '获取商品信息失败',
            icon: 'none'
          });
        }
      },
      (err) => {
        // 处理请求失败
        console.error('请求商品详情失败:', err);
        
        // 请求失败时也尝试合并相同商品
        const currentOrderDetail = this.data.orderDetail;
        if (currentOrderDetail && Array.isArray(currentOrderDetail.products)) {
          const productCountMap = {};
          currentOrderDetail.products.forEach(product => {
            if (!productCountMap[product.id]) {
              productCountMap[product.id] = 0;
            }
            productCountMap[product.id] += product.quantity || 1;
          });
          
          const mergedProducts = [];
          Object.keys(productCountMap).forEach(productId => {
            const product = currentOrderDetail.products.find(p => p.id === productId);
            if (product) {
              mergedProducts.push({
                ...product,
                quantity: productCountMap[product.id]
              });
            }
          });
          
          this.setData({
            orderDetail: {
              ...currentOrderDetail,
              products: mergedProducts,
              productCount: Object.values(productCountMap).reduce((sum, count) => sum + count, 0)
            }
          });
          this.loadSubOrders();
        }
        
        // 显示错误提示
        wx.showToast({
          title: '获取商品信息失败',
          icon: 'none'
        });
      }
    );
  },

  loadSubOrders() {
    const { orderId, orderDetail } = this.data;
    if (!orderId || !orderDetail || !Array.isArray(orderDetail.products)) {
      return;
    }

    app.req.post('/order/query_sub_order_data', {
      order_id: orderId
    }, (res) => {
      const subOrders = res && res.code === 200 && res.data && Array.isArray(res.data.sub_orders)
        ? res.data.sub_orders
        : [];
      if (subOrders.length === 0) {
        return;
      }

      const subOrderMap = {};
      subOrders.forEach((subOrder) => {
        const commodityId = (subOrder.commodity_id || '').toString();
        if (commodityId && !subOrderMap[commodityId]) {
          subOrderMap[commodityId] = subOrder;
        }
      });

      const products = orderDetail.products.map((product) => {
        const subOrder = subOrderMap[(product.id || '').toString()];
        if (!subOrder) {
          return product;
        }
        return {
          ...product,
          subOrderId: subOrder.sub_order_id,
          reviewable: orderDetail.status === 'delivered' && subOrder.status === 'delivered'
        };
      });

      this.setData({
        orderDetail: {
          ...orderDetail,
          products
        }
      });
    }, (err) => {
      console.error('查询子订单失败:', err);
    });
  },
  
  /**
   * 跳转到物流详情页面
   */
  navigateToLogisticsPage() {
    const { orderDetail } = this.data;
    console.log('跳转到物流详情页面，订单详情:', orderDetail);
    if (orderDetail && orderDetail.id) {
      console.log('订单状态:', orderDetail.status);
      // 如果订单状态为售后中，跳转到售后订单详情页面
      if (orderDetail.status === 'processing' || orderDetail.statusText === '售后中') {
        console.log('订单状态为售后中，准备跳转到售后订单详情页面');
        // 从订单详情中获取退货订单号
        const returnOrderId = orderDetail.process_num || orderDetail.return_order_id || '';
        console.log('退货订单号:', returnOrderId);
        if (returnOrderId) {
          console.log('跳转到售后订单详情页面，退货订单号:', returnOrderId);
          wx.navigateTo({
            url: `/pages/my/order/return_detail/index?returnOrderId=${returnOrderId}`
          });
        } else {
          console.log('退货订单号不存在');
          wx.showToast({
            title: '退货订单号不存在',
            icon: 'none'
          });
        }
      } else {
        console.log('订单状态不是售后中，跳转到普通物流页面');
        // 其他状态跳转到普通物流页面
        wx.navigateTo({
          url: `/pages/my/order/logistics/index?orderId=${orderDetail.id}`
        });
      }
    } else {
      console.log('订单详情不存在');
    }
  },

  /**
   * 跳转到商品详情页面
   */
  navigateToProductDetail(e) {
    const { productId } = e.currentTarget.dataset;
    if (productId) {
      wx.navigateTo({
        url: `/pages/commodity/goods/index?id=${productId}`
      });
    }
  },

  navigateToReview(e) {
    const { productId, subOrderId, styleCode, productName } = e.currentTarget.dataset;
    const { orderDetail } = this.data;
    if (!productId || !subOrderId) {
      wx.showToast({
        title: '评价信息缺失',
        icon: 'none'
      });
      return;
    }
    wx.navigateTo({
      url: `/pages/my/order/review/index?orderId=${orderDetail.id}&subOrderId=${subOrderId}&commodityId=${productId}&styleCode=${styleCode || ''}&productName=${encodeURIComponent(productName || '')}`
    });
  },

  /**
   * 联系客服
   */
  contactService() {
    wx.showToast({
      title: '正在连接客服...',
      icon: 'none'
    });
  },

  /**
   * 再次购买：将商品添加到购物车
   */
  buyAgain() {
    const { orderDetail } = this.data;
    const app = getApp();
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    // 验证用户是否登录
    if (!userId) {
      wx.showToast({
        title: '用户未登录',
        icon: 'none'
      });
      return;
    }
    
    // 显示加载状态
    wx.showLoading({
      title: '添加中...',
    });
    
    // 获取第一个商品（可以根据实际需求修改为选择特定商品）
    const firstProduct = orderDetail && orderDetail.products && orderDetail.products.length > 0 ? orderDetail.products[0] : null;
    
    if (!firstProduct || !firstProduct.id) {
      wx.hideLoading();
      wx.showToast({
        title: '商品信息获取失败',
        icon: 'none'
      });
      return;
    }
    
    // 构建请求数据
    const requestData = {
      user_id: userId,
      commodity_code: firstProduct.id
    };
    
    // 发送POST请求到/cart/add_to_cart接口
    app.req.post('/cart/add_to_cart', requestData, 
      (res) => {
        wx.hideLoading();
        // 根据响应处理结果
        if (res.code === 200) {
          wx.showToast({
            title: res.message || '已加入感兴趣的商品',
            icon: 'success'
          });
        } else {
          wx.showToast({
            title: res.message || '添加失败',
            icon: 'none'
          });
        }
      },
      (err) => {
        console.error('加入购物车请求失败:', err);
        wx.hideLoading();
        wx.showToast({
          title: '网络异常',
          icon: 'none'
        });
      }
    );
  },

  /**
   * 确认签收
   */
  confirmReceipt() {
    const that = this;
    wx.showModal({
      title: '确认签收',
      content: '确认已收到商品并完成签收吗？',
      success(res) {
        if (res.confirm) {
          wx.showLoading({ title: '确认中...' });
          app.req.post('/order/order_receive', {
            order_id: that.data.orderId,
            user_id: Number(getUserId())
          }, () => {
            wx.hideLoading();
            wx.showToast({
              title: '已签收',
              icon: 'success'
            });
            that.loadOrderDetail();
          }, (err) => {
            console.error('确认签收失败:', err);
            wx.hideLoading();
            wx.showToast({
              title: '确认签收失败',
              icon: 'none'
            });
          });
        }
      }
    });
  },

  /**
   * 取消订单
   */
  cancelOrder() {
    const that = this;
    const orderId = this.data.orderId;
    
    wx.showModal({
      title: '取消订单',
      content: '确定要取消该订单吗？',
      success(res) {
        if (res.confirm) {
          // 显示加载中提示
          wx.showLoading({
            title: '处理中...',
          });
          
          // 发送取消订单请求
          app.req.post('/order/cancel', { order_id: orderId,
          "user_id": getUserId() },
            (res) => {
              wx.hideLoading();
              if (res && res.status === 'success') {
                // 更新订单状态
                const updatedOrder = {
                  ...that.data.orderDetail,
                  status: 'canceled',
                  statusText: '已取消'
                };
                
                that.setData({
                  orderDetail: updatedOrder
                });
                
                wx.showToast({
                  title: '订单已取消',
                  icon: 'success'
                });
              } else {
                wx.showToast({
                  title: '取消订单失败',
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
   * 申请售后
   */
  applyAfterSale() {
    const { orderId, orderDetail } = this.data;
    const orderStatus = orderDetail ? orderDetail.status : '';
    console.log('申请售后，订单ID:', orderId, '订单状态:', orderStatus);
    wx.navigateTo({
      url: `/pages/my/order/return/index?id=${orderId}&status=${orderStatus}`
    });
  },

  /**
   * 页面卸载时处理返回逻辑
   */
  onUnload() {
    // 如果是从购买页面跳转过来的，直接跳转到个人主页
    if (this.data.from === 'buy_order') {
      wx.switchTab({
        url: '/pages/my/index/index'
      });
    }
  },

  /**
   * 返回上一页
   */
  navigateBack() {
    // 如果是从购买页面跳转过来的，直接跳转到个人主页
    if (this.data.from === 'buy_order') {
      wx.switchTab({
        url: '/pages/my/index/index',
        fail: function() {
          wx.reLaunch({
            url: '/pages/my/index/index'
          });
        }
      });
    } else {
      // 正常的返回逻辑
      wx.navigateBack();
    }
  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {
    return {
      title: '订单详情',
      path: `/pages/my/order/detail/index?id=${this.data.orderId}`
    };
  }
})
