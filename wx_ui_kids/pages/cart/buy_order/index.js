// pages/cart/buy_order/index.js
const app = getApp()
Page({

  /**
   * 页面的初始数据
   */
  data: {
    selectedItems: [
      {
        id: '1',
        name: '[春Vol.3] 泡泡纱防晒服 UPF50+抗皱外套',
        specification: '甘蓝色;165cm/M',
        price: 339,
        priceFormatted: '339.00',
        quantity: 1,
        image: '/images/products.png'
      }
    ],
    totalPrice: 339,
    totalPriceFormatted: '339.00',
    discount: 0,
    finalPrice: 339,
    finalPriceFormatted: '339.00',
    coupon: '',
    address: {
      id: 1,
      name: '张三',
      phone: '13800138000',
      province: '上海市',
      city: '上海市',
      district: '浦东新区',
      detail: '张江高科技园区博云路2号'
    }, // 收货地址
    orderNo: '', // 订单编号
    orderTime: '', // 下单时间
    remark: '' // 订单备注
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 监听购物车页面传递的数据
    const eventChannel = this.getOpenerEventChannel()
    eventChannel.on('selectedItems', (data) => {
      // 确保接收到的数据正确
      const items = data.items || this.getMockProducts()
      // 转换购物车数据格式以匹配下单页面的期望格式
      const formattedItems = items.map(item => {
        // 确保price是有效数字
        const price = typeof item.price === 'number' && !isNaN(item.price) ? item.price : 0;
        // 确保quantity是有效数字
        const quantity = typeof (item.count || item.quantity) === 'number' && !isNaN(item.count || item.quantity) && (item.count || item.quantity) > 0 ? (item.count || item.quantity) : 1;
        
        return {
          id: item.id,
          name: item.title || item.name || '',  // 处理title和name的差异
          specification: item.spec || item.specification || '',  // 处理spec和specification的差异
          price: price,
          priceFormatted: price.toFixed(2),  // 预先格式化价格
          quantity: quantity,
          image: item.image || '/images/products.png'
        };
      })
      const totalPrice = typeof data.totalPrice === 'number' && !isNaN(data.totalPrice) ? data.totalPrice : this.calculateTotalPrice(formattedItems)
      const finalPrice = this.calculateFinalPrice(totalPrice, this.data.discount)
      this.setData({
        selectedItems: formattedItems,
        totalPrice: totalPrice,
        totalPriceFormatted: totalPrice.toFixed(2),  // 预先格式化总价
        finalPrice: finalPrice,
        finalPriceFormatted: finalPrice.toFixed(2)  // 预先格式化订单金额
      })
      console.log('接收购物车数据并转换格式:', formattedItems)
      console.log('计算后的总价:', totalPrice)
      // 生成订单信息
      this.generateOrderInfo()
    })
    
    // 尝试加载默认地址
    this.loadDefaultAddress()
  },
  


  /**
   * 清空下单内容
   */
  clearOrderContent() {
    this.setData({
      selectedItems: [],
      totalPrice: 0,
      finalPrice: 0,
      discount: 0,
      address: {},
      remark: '',
      coupon: ''
    })
    wx.showToast({
      title: '下单内容已清空',
      icon: 'success',
      duration: 2000
    })
  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {

  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 从地址选择页面返回时，检查是否有选择的地址
    const address = wx.getStorageSync('selectedAddress')
    if (address) {
      this.setData({
        address: address
      })
      wx.removeStorageSync('selectedAddress') // 清除临时存储
    }
    
    // 从优惠券选择页面返回时，检查是否有选择的优惠券
    const coupon = wx.getStorageSync('selectedCoupon')
    if (coupon) {
      this.setData({
        coupon: coupon.name,
        discount: coupon.discount
      })
      wx.removeStorageSync('selectedCoupon') // 清除临时存储
    }
  },

  /**
   * 生成订单信息
   */
  generateOrderInfo() {
    // 生成订单编号
    const timestamp = Date.now()
    const random = Math.floor(Math.random() * 10000)
    const orderNo = `ORD${timestamp}${random}`
    
    // 生成当前时间
    const now = new Date()
    const year = now.getFullYear()
    const month = String(now.getMonth() + 1).padStart(2, '0')
    const day = String(now.getDate()).padStart(2, '0')
    const hours = String(now.getHours()).padStart(2, '0')
    const minutes = String(now.getMinutes()).padStart(2, '0')
    const seconds = String(now.getSeconds()).padStart(2, '0')
    const orderTime = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    
    this.setData({
      orderNo: orderNo,
      orderTime: orderTime
    })
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
    const userId = globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || '';
    
    console.log('当前用户ID:', userId);
    
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
            district: addressToUse.county || '',
            detail: addressToUse.detailed_address || '',
            is_default: !!addressToUse.is_default
          };
          
          this.setData({
            address: formattedAddress
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
      district: '浦东新区',
      detail: '张江高科技园区博云路2号',
      is_default: true
    };
    
    this.setData({
      address: defaultAddress
    });
  },

  /**
   * 选择地址
   */
  selectAddress() {
    console.log('跳转到地址选择页面，传递from=order参数');
    // 跳转到地址选择页面，确保传递from=order参数
    wx.navigateTo({
      url: '/pages/my/address/index?from=order'
    })
  },

  /**
   * 选择优惠券
   */
  selectCoupon() {
    // 跳转到优惠券选择页面
    wx.navigateTo({
      url: '/pages/my/coupon/index?from=order'
    })
  },

  /**
   * 设置优惠券
   */
  setCoupon(coupon, discount) {
    const newFinalPrice = this.calculateFinalPrice(this.data.totalPrice, discount)
    this.setData({
      coupon: coupon,
      discount: discount,
      finalPrice: newFinalPrice,
      finalPriceFormatted: newFinalPrice.toFixed(2)
    })
  },

  /**
   * 减少商品数量
   */
  decreaseQuantity(e) {
    const { id } = e.currentTarget.dataset
    const items = this.data.selectedItems
    const index = items.findIndex(item => item.id === id)
    
    if (index !== -1 && items[index].quantity > 1) {
      items[index].quantity--
      const newTotalPrice = this.calculateTotalPrice(items)
      const newFinalPrice = this.calculateFinalPrice(newTotalPrice, this.data.discount)
      this.setData({
        selectedItems: items,
        totalPrice: newTotalPrice,
        totalPriceFormatted: newTotalPrice.toFixed(2),
        finalPrice: newFinalPrice,
        finalPriceFormatted: newFinalPrice.toFixed(2)
      })
    }
  },

  /**
   * 增加商品数量
   */
  increaseQuantity(e) {
    const { id } = e.currentTarget.dataset
    const items = this.data.selectedItems
    const index = items.findIndex(item => item.id === id)
    
    if (index !== -1) {
      items[index].quantity++
      const newTotalPrice = this.calculateTotalPrice(items)
      const newFinalPrice = this.calculateFinalPrice(newTotalPrice, this.data.discount)
      this.setData({
        selectedItems: items,
        totalPrice: newTotalPrice,
        totalPriceFormatted: newTotalPrice.toFixed(2),
        finalPrice: newFinalPrice,
        finalPriceFormatted: newFinalPrice.toFixed(2)
      })
    }
  },

  /**
   * 计算总价
   */
  calculateTotalPrice(items) {
    if (!Array.isArray(items) || items.length === 0) {
      return 0
    }
    return items.reduce((total, item) => {
      // 添加防御性检查
      const price = typeof item.price === 'number' && !isNaN(item.price) ? item.price : 0
      const quantity = typeof item.quantity === 'number' && !isNaN(item.quantity) && item.quantity > 0 ? item.quantity : 0
      return total + (price * quantity)
    }, 0)
  },

  /**
   * 计算订单金额
   */
  calculateFinalPrice(totalPrice, discount) {
    return totalPrice - discount
  },

  /**
   * 处理订单备注输入
   */
  onRemarkInput(e) {
    this.setData({
      remark: e.detail.value
    })
  },

  /**
   * 提交订单
   */
  submitOrder() {
    // 检查是否选择了地址
    if (!this.data.address.name) {
      wx.showToast({
        title: '请选择收货地址',
        icon: 'none'
      })
      return
    }
    
    // 检查是否有商品
    if (this.data.selectedItems.length === 0) {
      wx.showToast({
        title: '订单中没有商品',
        icon: 'none'
      })
      return
    }
    
    // 构建订单数据
    const address = this.data.address;
    
    // 构建商品列表，按购买数量重复商品编码
    const product_list = [];
    this.data.selectedItems.forEach(item => {
      for (let i = 0; i < item.quantity; i++) {
        product_list.push(item.id);
      }
    });
    
    // 构建请求体数据
    const requestData = {
      receiver_name: address.name,
      receiver_phone: address.phone,
      province: address.province,
      city: address.city,
      county: address.district,
      detailed_address: address.detail,
      order_amount: this.data.finalPrice,
      product_list: product_list,
      user_id: app.globalData.userInfo && app.globalData.userInfo.user_id ? app.globalData.userInfo.user_id : 774231,
      remark: this.data.remark
    };
    
    // 显示加载中
    wx.showLoading({
      title: '提交订单中...',
    })
    
    // 发送订单数据到服务器
    app.req.post(
      '/order/add_order',
      requestData,
      (res) => {
        wx.hideLoading();
        
        // 详细记录响应内容，便于调试
        console.log('订单创建响应完整数据:', res);
        
        // 处理响应 - 增加兼容性，处理不同的响应格式
        let orderId = null;
        let isSuccess = false;
        
        if (res.status === 'success' && res.data) {
          isSuccess = true;
          orderId = res.data.order_id || (res.data.data && res.data.data.order_id);
        } else if (res.code === 200 && res.data) {
          isSuccess = true;
          orderId = res.data.order_id || (res.data.data && res.data.data.order_id);
        } else if (res.success && res.data) {
          isSuccess = true;
          orderId = res.data.order_id || (res.data.data && res.data.data.order_id);
        }
        
        if (isSuccess && orderId) {
          console.log('订单创建成功，订单ID:', orderId);
          
          // 从购物车删除已购买的商品
          const globalUserInfo = app.globalData.userInfo || {};
          const userId = globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || '';
          const commodityCodes = this.data.selectedItems.map(item => item.id);
          app.req.post('/cart/batch_delete_from_cart', {
            user_id: userId,
            commodity_codes: commodityCodes
          }, (deleteRes) => {
            // 无论删除是否成功，都继续下单流程
            if (deleteRes && deleteRes.code !== 200) {
              console.warn('删除购物车商品失败:', deleteRes);
            } else if (!deleteRes) {
              console.warn('删除购物车商品响应为空');
            }
          }, (deleteErr) => {
            console.error('删除购物车商品网络错误:', deleteErr);
          });
          
          // 显示成功提示
          wx.showToast({
            title: '订单已提交',
            icon: 'success',
            duration: 2000,
            success: () => {
              // 跳转订单详情页面
              setTimeout(() => {
                try {
                  console.log('尝试跳转到订单详情页，路径: /pages/my/order/detail/index?id=' + orderId);
                  wx.navigateTo({
                    url: '/pages/my/order/detail/index?id=' + orderId,
                    success: function(res) {
                      console.log('跳转成功');
                    },
                    fail: function(err) {
                      console.error('跳转失败:', err);
                      wx.showToast({
                        title: '跳转订单详情页失败',
                        icon: 'none'
                      });
                    }
                  })
                } catch (e) {
                  console.error('跳转异常:', e);
                  wx.showToast({
                    title: '跳转失败，请手动查看订单',
                    icon: 'none'
                  });
                }
              }, 2000)
            }
          })
        } else {
          // 显示失败提示，提供更具体的错误信息
          const errorMsg = res.message || res.errMsg || '订单创建失败';
          console.error('订单创建失败:', errorMsg, res);
          wx.showToast({
            title: errorMsg,
            icon: 'none'
          })
        }
      },
      (err) => {
        wx.hideLoading();
        console.error('提交订单网络失败:', err);
        wx.showToast({
          title: '网络异常，请重试',
          icon: 'none'
        })
      }
    );
  },
  
  /**
   * 获取模拟商品数据，用于展示效果
   */
  getMockProducts() {
    return [{
      id: '123456',
      name: '幼岚"出汗不会黏身"女士泡泡纱防晒服',
      image: '/images/products.png',
      price: 339.00,
      quantity: 1,
      specification: '甘蓝色；165cmM'
    }]
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
