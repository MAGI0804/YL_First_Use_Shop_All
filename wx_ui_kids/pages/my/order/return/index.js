// pages/my/order/return/index.js
const app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {
    orderId: '',
    orderStatus: '',
    orderInfo: null,
    // 售后类型
    afterSalesTypes: [
      { label: '仅退款', value: 'refund' },
      { label: '退货退款', value: 'return' },
      { label: '换货', value: 'exchange' },
      { label: '补发', value: 'reissue' }
    ],
    selectedType: '',
    // 售后原因
    afterSalesReasons: [
      { label: '商品质量问题', value: 'quality_issue' },
      { label: '商品与描述不符', value: 'description_mismatch' },
      { label: '商品损坏', value: 'damaged' },
      { label: '拍错/不想要了', value: 'not_wanted' },
      { label: '其他原因', value: 'other' }
    ],
    selectedReason: '',
    // 选择的商品（数组格式，用于提交）
    selectedProducts: [],
    // 商品选中状态（对象格式，用于UI显示）
    productSelected: {},
    // 问题描述
    description: '',
    // 上传的图片
    uploadedImages: [],
    // 选中的地址
    selectedAddress: null,
    // 是否可以提交
    canSubmit: false
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    console.log('页面加载，收到的参数:', options);
    if (options.id) {
      // 强制更新 orderId，确保即使页面被缓存也能获取最新的订单ID
      this.setData({ 
        orderId: options.id,
        orderStatus: options.order_status || options.status || ''
      });
      console.log('更新订单ID为:', options.id, '订单状态为:', options.order_status || options.status);
      
      // 根据订单状态设置可用的售后类型和原因
      this.setAvailableOptions();
      
      this.loadOrderInfo();
    }
  },

  /**
   * 根据订单状态设置可用的售后类型和原因
   */
  setAvailableOptions() {
    const { orderStatus } = this.data;
    console.log('根据订单状态设置可用的售后类型和原因，订单状态:', orderStatus);
    
    // 默认的售后类型和原因
    let afterSalesTypes = [
      { label: '仅退款', value: 'refund' },
      { label: '退货退款', value: 'return' },
      { label: '换货', value: 'exchange' },
      { label: '补发', value: 'reissue' }
    ];
    
    let afterSalesReasons = [
      { label: '商品质量问题', value: 'quality_issue' },
      { label: '商品与描述不符', value: 'description_mismatch' },
      { label: '商品损坏', value: 'damaged' },
      { label: '拍错/不想要了', value: 'not_wanted' },
      { label: '其他原因', value: 'other' }
    ];
    
    // 当订单状态为已发货时，只能选择退货退款，原因只能选不想要了/拍错和其它原因
    if (orderStatus === 'shipped') {
      console.log('订单状态为已发货，限制售后类型和原因');
      afterSalesTypes = [
        { label: '退货退款', value: 'return' }
      ];
      
      afterSalesReasons = [
        { label: '拍错/不想要了', value: 'not_wanted' },
        { label: '其他原因', value: 'other' }
      ];
      
      // 默认选择退货退款
      this.setData({ selectedType: 'return' });
    }
    
    this.setData({
      afterSalesTypes,
      afterSalesReasons
    });
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 从地址选择页面返回时，检查是否有选择的地址
    const selectedAddress = wx.getStorageSync('selectedAddress');
    if (selectedAddress) {
      console.log('从地址选择页面返回，获取到地址:', selectedAddress);
      this.setData({ selectedAddress });
      wx.removeStorageSync('selectedAddress'); // 清除临时存储
      this.checkCanSubmit();
    } else {
      // 只有在不是从地址选择页面返回时，才重新加载订单信息
      // 这样可以避免覆盖用户选择的地址
      if (this.data.orderId) {
        console.log('页面显示，重新加载订单信息，订单ID:', this.data.orderId);
        this.loadOrderInfo();
      }
    }
  },

  /**
   * 加载订单信息
   */
  loadOrderInfo() {
    const { orderId, orderStatus } = this.data;
    console.log('加载订单信息，订单ID:', orderId, '订单状态:', orderStatus);
    
    // 重置页面状态，确保加载新订单时不保留之前的状态
    this.setData({
      selectedProducts: [],
      productSelected: {},
      selectedReason: '',
      description: '',
      uploadedImages: [],
      selectedAddress: {}, // 初始化为空对象，以便地址组件正常显示
      canSubmit: false
    });
    
    // 当订单状态不是已发货时，加载默认地址
    if (orderStatus !== 'shipped') {
      this.loadDefaultAddress();
    }
    
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    app.req.post('/order/query_by_user_id', {
      page: 1,
      page_size: 10,
      shopname: 'youlan_kids',
      user_id: userId,
      order_id: orderId
    }, (res) => {
      console.log('订单信息响应:', res);
      
      if (res && res.code === 200 && res.data && res.data.data && res.data.data.length > 0) {
        // 查找与请求的 orderId 匹配的订单
        const order = res.data.data.find(item => item.order_id === orderId) || res.data.data[0];
        console.log('订单数据:', order);
        console.log('请求的订单ID:', orderId);
        console.log('返回的订单ID:', order.order_id);
        console.log('订单是否匹配:', order.order_id === orderId);
        
        const statusMap = {
          'pending': '待处理',
          'shipped': '已发货',
          'delivered': '已送达',
          'canceled': '已取消',
          'processing': '售后中'
        };
        
        const orderInfo = {
          id: order.order_id,
          status: order.status,
          statusText: statusMap[order.status] || order.status,
          createTime: order.order_time,
          totalPrice: parseFloat(order.order_amount).toFixed(2),
          products: []
        };
        
        // 获取商品信息
        if (Array.isArray(order.product_list)) {
          console.log('商品ID列表:', order.product_list);
          this.getProductDetails(order.product_list, orderInfo);
        } else {
          console.log('订单中没有商品列表');
          this.setData({ orderInfo });
        }
      } else {
        wx.showToast({
          title: '获取订单信息失败',
          icon: 'none'
        });
      }
    }, (err) => {
      console.error('获取订单信息失败:', err);
      wx.showToast({
        title: '网络请求失败',
        icon: 'none'
      });
    });
  },

  /**
   * 获取商品详情
   */
  getProductDetails(productIds, orderInfo) {
    console.log('获取商品详情，商品ID列表:', productIds);
    
    app.req.post('/commodity/batch_get_products_by_ids', {
      commodity_ids: productIds
    }, (res) => {
      console.log('商品详情响应:', res);
      
      if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
        console.log('商品数据:', res.data.data);
        
        const products = res.data.data.map(product => {
          const productId = String(product.commodity_id); // 确保ID为字符串类型
          const productData = {
            id: productId,
            name: product.name || '商品名称',
            price: product.price ? parseFloat(product.price).toFixed(2) : '0.00',
            quantity: 1,
            image: product.image || product.promo_image_url || '/images/products.png'
          };
          console.log('处理后的商品数据:', productData, 'ID类型:', typeof productId);
          return productData;
        });
        
        console.log('所有商品数据:', products);
        orderInfo.products = products;
        console.log('更新后的订单信息:', orderInfo);
        
        // 调试：直接打印商品ID和类型
        products.forEach((product, index) => {
          console.log(`商品${index + 1} ID:`, product.id, '类型:', typeof product.id);
        });
        
        this.setData({ orderInfo });
      } else {
        console.log('获取商品信息失败，响应数据:', res);
        wx.showToast({
          title: '获取商品信息失败',
          icon: 'none'
        });
      }
    }, (err) => {
      console.error('获取商品信息失败:', err);
      wx.showToast({
        title: '网络请求失败',
        icon: 'none'
      });
    });
  },

  /**
   * 选择售后类型
   */
  selectType(e) {
    const value = e.currentTarget.dataset.value;
    this.setData({ selectedType: value });
    this.checkCanSubmit();
  },

  /**
   * 选择售后原因
   */
  selectReason(e) {
    const value = e.currentTarget.dataset.value;
    this.setData({ selectedReason: value });
    this.checkCanSubmit();
  },

  /**
   * 切换商品选择
   */
  toggleProduct(e) {
    const productId = e.currentTarget.dataset.id; // 微信小程序中data-*属性的值已经是字符串类型
    console.log('切换商品选择，商品ID:', productId, '类型:', typeof productId);
    
    // 获取当前选中状态
    const productSelected = { ...this.data.productSelected };
    const isSelected = productSelected[productId] || false;
    console.log('当前商品选中状态:', isSelected);
    
    // 切换选中状态
    productSelected[productId] = !isSelected;
    console.log('更新后商品选中状态:', productSelected);
    
    // 更新 selectedProducts 数组
    const selectedProducts = Object.keys(productSelected).filter(id => productSelected[id]);
    console.log('更新后选中商品数组:', selectedProducts);
    
    // 设置数据
    this.setData({ 
      productSelected: productSelected,
      selectedProducts: selectedProducts
    });
    
    this.checkCanSubmit();
  },

  /**
   * 输入问题描述
   */
  inputDescription(e) {
    const value = e.detail.value;
    this.setData({ description: value });
    this.checkCanSubmit();
  },

  /**
   * 判断商品是否被选中
   */
  isProductSelected(productId) {
    let selectedProducts = this.data.selectedProducts;
    // 确保 selectedProducts 是数组
    if (typeof selectedProducts === 'string') {
      selectedProducts = selectedProducts.split(',').filter(id => id);
    } else if (!Array.isArray(selectedProducts)) {
      selectedProducts = [];
    }
    // 使用字符串比较，确保类型一致
    return selectedProducts.some(id => String(id) === String(productId));
  },

  /**
   * 选择图片
   */
  chooseImage() {
    wx.chooseImage({
      count: 5 - this.data.uploadedImages.length,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const newImages = res.tempFilePaths;
        this.setData({
          uploadedImages: [...this.data.uploadedImages, ...newImages]
        });
      }
    });
  },

  /**
   * 删除图片
   */
  removeImage(e) {
    const index = e.currentTarget.dataset.index;
    const images = [...this.data.uploadedImages];
    images.splice(index, 1);
    this.setData({ uploadedImages: images });
  },

  /**
   * 选择地址
   */
  selectAddress() {
    console.log('跳转到地址选择页面，传递from=return参数');
    // 跳转到地址选择页面，确保传递from=return参数
    app.navigateTo({
      url: '/pages/my/address/index?from=return'
    });
  },

  /**
   * 加载默认地址
   */
  loadDefaultAddress() {
    wx.showLoading({
      title: '加载地址中...',
    });
    
    // 按照应用标准方式获取用户ID：先检查全局变量，再检查本地存储
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    // 确保user_id存在且不为空
    if (!userId) {
      console.error('用户ID不存在，无法获取地址列表');
      wx.hideLoading();
      // 使用模拟数据作为备选
      this.setDefaultAddressMockData();
      return;
    }
    
    // 构建请求数据
    const requestData = { user_id: userId };
    
    // 发送请求获取地址列表
    app.req.post('/address/get_addresses', requestData, 
      (res) => {
        wx.hideLoading();
        console.log('获取地址列表响应:', res);
        // 根据响应格式处理
        if (res.code === 200 && res.data && Array.isArray(res.data.addresses) && res.data.addresses.length > 0) {
          // 查找默认地址
          const defaultAddress = res.data.addresses.find(item => item.is_default);
          // 如果有默认地址，使用默认地址；否则使用第一个地址
          const addressToUse = defaultAddress || res.data.addresses[0];
          
          // 转换地址格式以适配页面显示
          const formattedAddress = {
            id: addressToUse.address_id,
            name: addressToUse.receiver_name || '',
            phone: addressToUse.phone_number || '',
            province: addressToUse.province || '',
            city: addressToUse.city || '',
            county: addressToUse.county || '',
            address: addressToUse.detailed_address || '',
            is_default: !!addressToUse.is_default
          };
          
          this.setData({
            selectedAddress: formattedAddress
          });
        } else {
          console.warn('未获取到地址列表或列表为空');
          // 使用模拟数据
          this.setDefaultAddressMockData();
        }
      },
      (err) => {
        console.error('获取地址列表网络错误:', err);
        wx.hideLoading();
        // 使用模拟数据
        this.setDefaultAddressMockData();
      }
    );
  },

  /**
   * 设置默认地址模拟数据
   */
  setDefaultAddressMockData() {
    const defaultAddress = {
      id: 1,
      name: '张三',
      phone: '13800138000',
      province: '上海市',
      city: '上海市',
      county: '浦东新区',
      address: '张江高科技园区博云路2号',
      is_default: true
    };
    
    this.setData({
      selectedAddress: defaultAddress
    });
  },

  /**
   * 检查是否可以提交
   */
  checkCanSubmit() {
    const { selectedType, selectedReason, selectedProducts, description, selectedAddress, orderStatus } = this.data;
    let canSubmit = false;
    
    if (selectedType && selectedReason && selectedProducts.length > 0 && description) {
      // 对于退货和换货类型，需要选择地址且地址信息完整
      // 但当订单状态为已发货时，不需要选择地址
      if ((selectedType === 'return' || selectedType === 'exchange') && (orderStatus === 'shipped' || (selectedAddress && selectedAddress.name))) {
        canSubmit = true;
      } else if (selectedType === 'refund') {
        // 仅退款不需要地址
        canSubmit = true;
      }
    }
    
    this.setData({ canSubmit });
  },

  /**
   * 提交售后申请
   */
  submitAfterSales() {
    const { orderId, selectedType, selectedReason, selectedAddress } = this.data;
    
    // 显示加载中
    wx.showLoading({
      title: '提交中...'
    });
    
    // 构建请求数据
    // 找到对应的中文原因
    const reasonObj = this.data.afterSalesReasons.find(item => item.value === selectedReason);
    const reasonText = reasonObj ? reasonObj.label : selectedReason;
    
    // 获取用户ID
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(wx.getStorageSync('user_id') || globalUserInfo.user_id || app.globalData.user_id || 0);
    
    const requestData = {
      order_id: orderId,
      reason: reasonText,
      specific_reasons: this.data.description,
      type: selectedType,
      user_id: userId,
      order_status: this.data.orderStatus,
      product_ids: this.data.selectedProducts
    };
    
    // 对于退货和换货类型，添加地址信息
    // 但当订单状态为已发货时，不需要添加地址信息
    if ((selectedType === 'return' || selectedType === 'exchange') && selectedAddress && this.data.orderStatus !== 'shipped') {
      requestData.buyer_province = selectedAddress.province;
      requestData.buyer_city = selectedAddress.city;
      requestData.buyer_county = selectedAddress.county;
      requestData.buyer_address = selectedAddress.address;
      requestData.buyer_phone = selectedAddress.phone;
    }
    
    console.log('提交售后申请，请求数据:', requestData);
    
    // 调用API提交售后申请
    app.req.post('/return_order/create', requestData, (res) => {
      wx.hideLoading();
      console.log('提交售后申请响应:', res);
      if (res && res.code === 200) {
        const returnOrderId = res.data && res.data.return_order_id ? res.data.return_order_id : '';
        wx.showToast({
          title: '提交成功',
          icon: 'success'
        });
        
        // 跳转到售后详情页，展示审核和寄回进度
        setTimeout(() => {
          if (returnOrderId) {
            app.redirectTo({
              url: `/pages/my/order/return_detail/index?returnOrderId=${returnOrderId}&orderId=${orderId}`
            });
          } else {
            app.navigateTo({
              url: `/pages/my/order/detail/index?id=${orderId}`
            });
          }
        }, 1500);
      } else {
        wx.showToast({
          title: res && res.msg ? res.msg : '提交失败',
          icon: 'none'
        });
      }
    }, (err) => {
      wx.hideLoading();
      console.error('提交售后申请失败:', err);
      wx.showToast({
        title: '网络请求失败',
        icon: 'none'
      });
    });
  }
});
