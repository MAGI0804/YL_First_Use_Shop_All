// pages/my/order/return_detail/index.js
const app = getApp();

Page({
  /**
   * 页面的初始数据
   */
  data: {
    // 加载状态
    loading: true,
    // 退货订单ID
    returnOrderId: '',
    // 原订单ID
    orderId: '',
    // 售后订单信息
    returnOrderInfo: null,
    // 原订单信息
    originalOrderInfo: null,
    // 错误信息
    errorMessage: ''
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    console.log('页面加载，收到的参数:', options);
    // 从URL参数中获取退货订单ID和原订单ID
    const returnOrderId = options.returnOrderId || options.return_order_no;
    if (returnOrderId) {
      this.setData({ 
        returnOrderId: returnOrderId,
        orderId: options.orderId || ''
      });
      console.log('更新退货订单ID为:', returnOrderId, '原订单ID为:', options.orderId);
      this.loadReturnOrderDetail();
    } else {
      this.setData({
        loading: false,
        errorMessage: '退货订单ID不存在'
      });
    }
  },

  /**
   * 加载售后订单详情
   */
  loadReturnOrderDetail() {
    if (!this.data.returnOrderId) {
      this.setData({
        loading: false,
        errorMessage: '退货订单ID不存在'
      });
      return;
    }

    this.setData({ loading: true, errorMessage: '' });
    
    // 准备请求参数
    const requestData = {
      return_order_id: this.data.returnOrderId
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起售后订单详情请求:', '/return_order/detail', requestData);
    app.req.post('/return_order/detail', requestData, 
      (res) => {
        console.log('售后订单详情响应:', res);
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.return_order) {
          const returnOrderData = res.data.return_order;
          
          // 处理商品列表
          let productList = [];
          if (Array.isArray(returnOrderData.product_list)) {
            productList = returnOrderData.product_list;
          } else if (typeof returnOrderData.product_list === 'string') {
            try {
              productList = JSON.parse(returnOrderData.product_list);
            } catch (e) {
              console.error('解析商品列表失败:', e);
              productList = [];
            }
          }
          
          // 转换售后类型和状态为中文
          const typeMap = {
            'return': '退货',
            'exchange': '换货退款',
            'refund': '仅退款'
          };
          
          const statusMap = {
            'pending': '待处理',
            'processing': '处理中',
            'completed': '已完成',
            'canceled': '已取消',
            'shipped': '已发货'
          };
          
          // 构造售后订单信息对象
          const returnOrderInfo = {
            ...returnOrderData,
            type_text: typeMap[returnOrderData.type] || '售后',
            status_text: statusMap[returnOrderData.status] || '未知状态',
            products: []
          };
          
          this.setData({ returnOrderInfo });
          
          // 获取商品详情
          if (productList.length > 0) {
            this.getProductDetails(productList, returnOrderInfo);
          } else {
            // 没有商品列表，直接加载原订单信息
            this.loadOriginalOrderInfo(returnOrderData.order_id);
          }
        } else {
          // 请求成功但数据格式不正确
          this.setData({
            loading: false,
            errorMessage: res && res.msg ? res.msg : '获取售后订单详情失败'
          });
        }
      },
      (err) => {
        // 处理请求失败
        console.error('请求售后订单详情失败:', err);
        this.setData({
          loading: false,
          errorMessage: '网络请求失败'
        });
      }
    );
  },

  /**
   * 获取商品详情
   */
  getProductDetails(productIds, returnOrderInfo) {
    const requestData = {
      commodity_ids: productIds
    };
    
    app.req.post('/commodity/batch_get_products_by_ids', requestData, 
      (res) => {
        console.log('商品详情响应:', res);
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 创建商品信息映射表
          const productMap = {};
          res.data.data.forEach(product => {
            productMap[product.commodity_id] = product;
          });
          
          // 更新商品信息
          const products = productIds.map(productId => {
            const realProduct = productMap[productId];
            // 提取款式编码（商品编码的前七位）
            const styleCode = productId.substring(0, 9);
            return {
              id: productId,
              styleCode: styleCode,
              name: realProduct ? (realProduct.name || '商品名称') : '商品名称',
              price: realProduct && realProduct.price ? parseFloat(realProduct.price).toFixed(2) : '0.00',
              image: realProduct ? (realProduct.image || realProduct.promo_image_url || '/images/products.png') : '/images/products.png'
            };
          });
          
          returnOrderInfo.products = products;
          this.setData({ returnOrderInfo });
        } else {
          console.log('获取商品信息失败，响应数据:', res);
        }
        
        // 加载原订单信息
        this.loadOriginalOrderInfo(returnOrderInfo.order_id);
      },
      (err) => {
        console.error('获取商品信息失败:', err);
        // 即使商品信息获取失败，也继续加载原订单信息
        this.loadOriginalOrderInfo(returnOrderInfo.order_id);
      }
    );
  },

  /**
   * 加载原订单信息
   */
  loadOriginalOrderInfo(orderId) {
    if (!orderId) {
      this.setData({ loading: false });
      return;
    }
    
    // 准备请求参数
    const requestData = {
      order_id: orderId
    };
    
    // 使用app.js中的req.post方法调用API
    console.log('发起原订单详情请求:', '/order/query_order_data', requestData);
    app.req.post('/order/query_order_data', requestData, 
      (res) => {
        console.log('原订单详情响应:', res);
        // 处理成功响应
        if (res && res.code === 200 && res.data) {
          const orderData = res.data.data || res.data;
          
          // 转换订单状态为中文
          const statusMap = {
            'pending': '待处理',
            'shipped': '已发货',
            'delivered': '已送达',
            'canceled': '已取消',
            'processing': '售后中'
          };
          
          // 解析商品列表
          let productList = [];
          try {
            if (Array.isArray(orderData.product_list)) {
              productList = orderData.product_list.map((productId) => ({
                id: productId.toString(),
                name: '加载中...',
                price: '0.00',
                quantity: 1,
                image: '/images/products.png'
              }));
            } else if (typeof orderData.product_list === 'string') {
              const products = JSON.parse(orderData.product_list);
              productList = products.map((productId) => ({
                id: productId.toString(),
                name: '加载中...',
                price: '0.00',
                quantity: 1,
                image: '/images/products.png'
              }));
            }
          } catch (e) {
            console.error('解析原订单商品列表失败:', e);
          }
          
          // 构造原订单信息对象
          const originalOrderInfo = {
            ...orderData,
            status_text: statusMap[orderData.status] || orderData.status,
            products: productList
          };
          
          this.setData({ originalOrderInfo });
          
          // 获取原订单商品详情
          if (productList.length > 0) {
            const productIds = productList.map(product => product.id);
            this.getOriginalOrderProductDetails(productIds, originalOrderInfo);
          }
        }
        
        this.setData({ loading: false });
      },
      (err) => {
        console.error('请求原订单详情失败:', err);
        this.setData({ loading: false });
      }
    );
  },

  /**
   * 获取原订单商品详情
   */
  getOriginalOrderProductDetails(productIds, originalOrderInfo) {
    const requestData = {
      commodity_ids: productIds
    };
    
    app.req.post('/commodity/batch_get_products_by_ids', requestData, 
      (res) => {
        console.log('原订单商品详情响应:', res);
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 创建商品信息映射表
          const productMap = {};
          res.data.data.forEach(product => {
            productMap[product.commodity_id] = product;
          });
          
          // 更新商品信息
          const products = originalOrderInfo.products.map(product => {
            const realProduct = productMap[product.id];
            // 提取款式编码（商品编码的前七位）
            const styleCode = product.id.substring(0, 9);
            return {
              ...product,
              styleCode: styleCode,
              name: realProduct ? (realProduct.name || product.name) : product.name,
              price: realProduct && realProduct.price ? parseFloat(realProduct.price).toFixed(2) : product.price,
              image: realProduct ? (realProduct.image || realProduct.promo_image_url || product.image) : product.image
            };
          });
          
          originalOrderInfo.products = products;
          this.setData({ originalOrderInfo });
        }
      },
      (err) => {
        console.error('获取原订单商品信息失败:', err);
      }
    );
  },

  /**
   * 跳转到商品详情页面
   */
  navigateToProductDetail(e) {
    const { productId } = e.currentTarget.dataset;
    console.log('跳转到商品详情页面，商品ID:', productId, '事件数据:', e.currentTarget.dataset);
    if (productId) {
      wx.navigateTo({
        url: `/pages/commodity/goods/index?id=${productId}`
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
   * 返回上一页
   */
  navigateBack() {
    wx.navigateBack({
      delta: 1
    });
  },

  /**
   * 跳转到物流详情页面
   */
  navigateToLogistics() {
    const { returnOrderInfo } = this.data;
    console.log('跳转到售后物流详情页面，售后订单详情:', returnOrderInfo);
    if (returnOrderInfo && returnOrderInfo.express_company && returnOrderInfo.express_number) {
      wx.navigateTo({
        url: `/pages/my/order/logistics/index?company=${returnOrderInfo.express_company}&trackingNumber=${returnOrderInfo.express_number}&type=return`
      });
    } else {
      wx.showToast({
        title: '物流信息不存在',
        icon: 'none'
      });
    }
  },

  /**
   * 跳转到原订单物流详情页面
   */
  navigateToOriginalLogistics() {
    const { originalOrderInfo } = this.data;
    console.log('跳转到原订单物流详情页面，原订单详情:', originalOrderInfo);
    if (originalOrderInfo && originalOrderInfo.express_company && originalOrderInfo.express_number) {
      wx.navigateTo({
        url: `/pages/my/order/logistics/index?company=${originalOrderInfo.express_company}&trackingNumber=${originalOrderInfo.express_number}&type=original`
      });
    } else {
      wx.showToast({
        title: '物流信息不存在',
        icon: 'none'
      });
    }
  },

  /**
   * 重试请求
   */
  retryRequest() {
    this.loadReturnOrderDetail();
  }
});