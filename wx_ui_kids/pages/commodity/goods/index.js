// pages/commodity/goods/index.js

const app = getApp();

Page({

  /**
   * 页面的初始数据
   */
  data: {
    indicatorDots: true,  // 显示轮播图指示器
    autoplay: true,       // 自动播放
    interval: 3000,       // 自动播放间隔时间(ms)
    duration: 500,        // 滑动动画时长(ms)
    currentSwiper: 0,     // 当前轮播图索引
    goodsImages: [
    ],
    displayPicturesList: [], // display_pictures列表，用于下方纵向展示
    selectedColorImage: '', // 选中颜色对应的图片
    goodsTitle: '[秋Vol.1] allblu幼岚【复古天鹅绒束口裤】',
    goodsSubtitle: '儿童裤子 25秋新款男女童童柔软ZY',
    currentPrice: 179,    // 当前价格
    originalPrice: 199,   // 原价
    stock: 978,           // 总库存
    currentItemStock: 0,   // 当前选择商品的库存
    serviceInfo: '线下门店 · 快递发货 · 收货后结算',
    loading: false,
    error: false,
    // 直接初始化默认的商品选项列表，确保选择界面始终有内容显示
    items: [
      {
        color: '奶油白',
        sizes: [
          { size: '90cm', commodity_id: '1001', inventory: 100 },
          { size: '100cm', commodity_id: '1002', inventory: 100 },
          { size: '110cm', commodity_id: '1003', inventory: 100 }
        ]
      },
      {
        color: '樱花粉',
        sizes: [
          { size: '90cm', commodity_id: '1004', inventory: 100 },
          { size: '100cm', commodity_id: '1005', inventory: 100 },
          { size: '110cm', commodity_id: '1006', inventory: 100 }
        ]
      }
    ],
    selectedColor: '',    // 选中的颜色
    selectedSize: '',     // 选中的尺码
    selectedCommodityId: '', // 选中的商品ID
    showSelectModal: false,  // 是否显示选择弹出层
    actionType: '',        // 操作类型：cart 或 buy
    quantity: 1,           // 选择的数量，默认为1
    currentStyleCode: '',
    reviewStats: {
      total: 0,
      averageRating: '0.00',
      goodRate: '0.00'
    },
    productReviews: [],
    reviewLoading: false
  },

  /**
   * 跳转到搜索页面
   */
  navigateToSearch() {
    wx.navigateTo({
      url: '/pages/search/index/index'
    });
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 保存options到实例，以便后续使用
    this.options = options;
    
    // 从options中获取商品ID或款式编码并请求商品详情数据
    if (options.id || options.style_code) {
      const goodsId = options.style_code || options.id;
      console.log('款式编码:', goodsId);
      this.setData({
        currentStyleCode: goodsId
      });
      this.fetchGoodsDetail(goodsId);
      this.loadProductReviews(goodsId);
    } else {
      this.setData({
        error: true,
        loading: false
      })
      wx.showToast({
        title: '商品ID不存在',
        icon: 'none'
      });
    }
  },

  /**
   * 根据商品ID获取商品详情
   */
  fetchGoodsDetail(goodsId) {
    const that = this;
    this.setData({
      loading: true,
      error: false
    });

    app.req.post(
      '/commodity/stylecode_commodities',
      {
        style_code: goodsId,
        shopname:'youlan_kids'
      },
      function(res) {
        if (res.code === 200) {
          const goodsData = res.data;
          // 构建轮播图图片数组，images中的主图作为第一张，display_pictures作为后续图片
      const goodsImages = [];
      // 找到主图
      if (goodsData.images && Array.isArray(goodsData.images)) {
        const mainImage = goodsData.images.find(img => img.is_main);
        if (mainImage && mainImage.url) {
          goodsImages.push(mainImage.url);
        } else if (goodsData.images.length > 0) {
          // 如果没有主图，使用第一张图
          goodsImages.push(goodsData.images[0].url);
        }
      }
      // 处理display_pictures，按顺序添加到轮播图中
      if (goodsData.display_pictures) {
        // 获取display_pictures中的键，并按数字排序
        const displayKeys = Object.keys(goodsData.display_pictures).sort((a, b) => parseInt(a) - parseInt(b));
        // 按排序后的顺序添加图片
        displayKeys.forEach(key => {
          const imageUrl = goodsData.display_pictures[key];
          if (imageUrl) {
            goodsImages.push(imageUrl);
          }
        });
      }
      // 如果没有足够的图片，确保至少有一个占位图
      if (goodsImages.length === 0) {
        goodsImages.push('/images/products.png');
      }
          
          // 只有当API返回了有效的items数据时才更新items
          let items = that.data.items; // 保留默认数据
          if (goodsData.items && goodsData.items.length > 0) {
            items = goodsData.items;
            // 初始化默认选中第一个颜色和对应的图片
            if (items.length > 0 && items[0].color) {
              // 获取主图作为默认颜色图片
              let defaultImage = '/images/products.png';
              if (goodsData.images && Array.isArray(goodsData.images)) {
                const mainImage = goodsData.images.find(img => img.is_main);
                if (mainImage && mainImage.url) {
                  defaultImage = mainImage.url;
                } else if (goodsData.images.length > 0) {
                  defaultImage = goodsData.images[0].url;
                }
              }
              that.setData({
                selectedColor: items[0].color,
                selectedColorImage: items[0].color_image || defaultImage
              });
            }
          }
          
          // 准备display_pictures数组，用于下方纵向展示
          let displayPicturesList = [];
          if (goodsData.display_pictures) {
            const displayKeys = Object.keys(goodsData.display_pictures).sort((a, b) => parseInt(a) - parseInt(b));
            displayPicturesList = displayKeys.map(key => goodsData.display_pictures[key]);
          }
          
          that.setData({
            goodsImages: goodsImages,
            goodsTitle: goodsData.name || '',
            goodsSubtitle: goodsData.subtitle || '',
            currentPrice: goodsData.price || 0,
            originalPrice: goodsData.original_price || 0,
            stock: goodsData.inventory || 100,
            serviceInfo: goodsData.service_info || '',
            loading: false,
            error: false,
            items: items,
            displayPicturesList: displayPicturesList // 保存display_pictures列表，用于下方纵向展示
          });
        } else {
          that.setData({
            loading: false,
            error: true
          });
          wx.showToast({
            title: '获取商品详情失败',
            icon: 'none'
          });
        }
      },
      function(err) {
        console.error('请求商品详情失败:', err);
        that.setData({
          loading: false,
          error: true
        });
        wx.showToast({
          title: '网络异常',
          icon: 'none'
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
   * 轮播图切换事件
   */
  swiperChange(e) {
    this.setData({
      currentSwiper: e.detail.current
    });
  },

  /**
   * 打开选择弹出层
   */
  openSelectModal(e) {
    // 检测用户是否登录
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || '';
    
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      // 跳转到登录页面，并传递当前商品ID
      const goodsId = this.options.style_code || this.options.id;
      wx.navigateTo({
        url: `/pages/accUser/index?returnPath=goods&id=${goodsId}`
      });
      return;
    }
    
    const actionType = e.currentTarget.dataset.type;
    this.setData({
      showSelectModal: true,
      actionType: actionType
    });
  },

  /**
   * 关闭选择弹出层
   */
  closeSelectModal() {
    this.setData({
      showSelectModal: false
    });
  },

  /**
   * 选择颜色
   */
  selectColor(e) {
    const color = e.currentTarget.dataset.color;
    const { items, goodsImages } = this.data;
    
    // 查找选中颜色对应的图片
    let selectedImage = '';
    const colorItem = items.find(item => item.color === color);
    if (colorItem && colorItem.color_image) {
      selectedImage = colorItem.color_image;
    } else if (goodsImages && goodsImages.length > 0) {
      // 如果没有颜色图片，保留当前第一张图片
      selectedImage = goodsImages[0];
    } else {
      // 如果都没有，使用默认占位图
      selectedImage = '/images/products.png';
    }
    
    // 更新选中颜色和图片
    this.setData({
      selectedColor: color,
      selectedSize: '', // 选择新颜色后重置尺码选择
      'goodsImages[0]': selectedImage, // 更新轮播图第一张图片为选中颜色的图片
      selectedColorImage: selectedImage // 更新选择界面显示的图片
    });
  },

  /**
   * 根据颜色获取对应的尺码列表
   */
  getSizesByColor(color) {
    const { items } = this.data;
    const colorItem = items.find(item => item.color === color);
    return colorItem ? colorItem.sizes || [] : [];
  },

  /**
   * 检查选中颜色是否有对应的尺码选项
   */
  hasSizesForColor(color) {
    const { items } = this.data;
    for (let i = 0; i < items.length; i++) {
      if (items[i].color === color && items[i].sizes && items[i].sizes.length > 0) {
        return true;
      }
    }
    return false;
  },

  /**
   * 选择尺码
   */
  selectSize(e) {
    const size = e.currentTarget.dataset.size;
    const { items, selectedColor } = this.data;
    
    // 找到对应的商品ID和库存
    let selectedCommodityId = '';
    let currentItemStock = 0;
    
    // 遍历items查找对应的颜色和尺码
    for (let i = 0; i < items.length; i++) {
      if (items[i].color === selectedColor && items[i].sizes) {
        for (let j = 0; j < items[i].sizes.length; j++) {
          if (items[i].sizes[j].size === size) {
            selectedCommodityId = items[i].sizes[j].commodity_id;
            currentItemStock = items[i].sizes[j].inventory || 0;
            break;
          }
        }
        break;
      }
    }
    
    this.setData({
      selectedSize: size,
      selectedCommodityId: selectedCommodityId,
      currentItemStock: currentItemStock
    });
  },

  /**
   * 确认选择
   */
  confirmSelect() {
    const { selectedColor, selectedSize, selectedCommodityId, actionType, quantity } = this.data;
    
    if (!selectedColor || !selectedSize || !selectedCommodityId) {
      wx.showToast({
        title: '请选择完整的商品规格',
        icon: 'none'
      });
      return;
    }
    
    this.setData({
      showSelectModal: false
    });
    
    if (actionType === 'cart') {
      this.addToCart(selectedCommodityId, quantity);
    } else if (actionType === 'buy') {
      this.buyNow(selectedCommodityId, quantity);
    }
  },
  
  /**
   * 增加购买数量
   */
  increaseQuantity() {
    const { quantity, currentItemStock } = this.data;
    const stock = currentItemStock > 0 ? currentItemStock : this.data.stock;
    if (quantity < stock) {
      this.setData({
        quantity: quantity + 1
      });
    } else {
      wx.showToast({
        title: '已达到最大库存',
        icon: 'none'
      });
    }
  },
  
  /**
   * 减少购买数量
   */
  decreaseQuantity() {
    const { quantity } = this.data;
    if (quantity > 1) {
      this.setData({
        quantity: quantity - 1
      });
    }
  },

  /**
   * 加入购物车
   */
  addToCart(commodityId, quantity = 1) {
    // 显示加载状态
    wx.showLoading({
      title: '添加中...',
    });
    
    // 实时检测用户是否登录
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || '';
    
    // 验证必要参数
    if (!userId) {
      wx.hideLoading();
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      // 跳转到登录页面，并传递当前商品ID
      const goodsId = this.options.style_code || this.options.id;
      wx.navigateTo({
        url: `/pages/accUser/index?returnPath=goods&id=${goodsId}`
      });
      return;
    }
    
    if (!commodityId) {
      wx.hideLoading();
      wx.showToast({
        title: '商品ID不存在',
        icon: 'none'
      });
      return;
    }
    
    // 构建请求数据
    const requestData = {
      user_id: userId,
      commodity_code: commodityId
    };
    
    // 数量大于1时添加quantity参数
    if (quantity > 1) {
      requestData.quantity = quantity;
    }
    
    // 发送POST请求到/cart/add_to_cart接口
    app.req.post('/cart/add_to_cart', requestData, 
      (res) => {
        wx.hideLoading();
        console.log('加入购物车响应:', res);
        // 根据响应处理结果
        if (res.code === 200) {
          wx.showToast({
            title: res.message || '已加入感兴趣的商品',
            icon: 'success'
          });
          
          // 可以根据需要处理返回的数据
          if (res.data) {
            console.log('商品代码:', res.data.commodity_code);
            console.log('数量:', res.data.quantity);
            console.log('购物车总数:', res.data.total_items);
          }
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
   * 立即购买
   */
  buyNow(commodityId, quantity = 1) {
    // 实时检测用户是否登录
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || '';
    
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      // 跳转到登录页面，并传递当前商品ID
      const goodsId = this.options.style_code || this.options.id;
      wx.navigateTo({
        url: `/pages/accUser/index?returnPath=goods&id=${goodsId}`
      });
      return;
    }
    
    const { goodsTitle, currentPrice, selectedColor, selectedSize } = this.data;
    
    // 准备要传递给下单页面的数据
    const orderData = {
      items: [{
        id: commodityId,
        name: goodsTitle,
        specification: `${selectedColor};${selectedSize}`,
        price: currentPrice,
        quantity: quantity,
        image: this.data.goodsImages[0] || '/images/products.png'
      }],
      totalPrice: currentPrice * quantity
    };
    
    // 跳转到购买订单页面并传递数据
    wx.navigateTo({
      url: '/pages/cart/buy_order/index',
      events: {},
      success: function(res) {
        // 通过 eventChannel 向被打开页面传送数据
        res.eventChannel.emit('selectedItems', orderData);
      }
    });
  },

  /**
   * 跳转到购物车页面
   */
  goToCart() {
    wx.switchTab({
      url: '/pages/cart/index/index'
    })
  },

  loadProductReviews(styleCode) {
    if (!styleCode) {
      return;
    }
    this.setData({ reviewLoading: true });

    app.req.post('/review/statistics', {
      style_code: styleCode
    }, (res) => {
      if (res && res.code === 200 && res.data && res.data.statistics) {
        const stats = res.data.statistics;
        this.setData({
          reviewStats: {
            total: stats.total || 0,
            averageRating: Number(stats.average_rating || 0).toFixed(2),
            goodRate: (Number(stats.good_rate || 0) * 100).toFixed(2)
          }
        });
      }
    }, (err) => {
      console.error('查询评价统计失败:', err);
    });

    app.req.post('/review/query_by_product', {
      style_code: styleCode,
      page: 1,
      page_size: 5
    }, (res) => {
      this.setData({ reviewLoading: false });
      const reviews = res && res.code === 200 && res.data && Array.isArray(res.data.data)
        ? res.data.data
        : [];
      this.setData({
        productReviews: reviews.map((review) => ({
          ...review,
          displayRating: '★★★★★'.slice(0, Number(review.rating || 0)),
          displayTags: this.parseReviewList(review.tags),
          displayReplies: Array.isArray(review.replies) ? review.replies : []
        }))
      });
    }, (err) => {
      console.error('查询商品评价失败:', err);
      this.setData({ reviewLoading: false });
    });
  },

  parseReviewList(value) {
    if (!value) {
      return [];
    }
    try {
      const parsed = JSON.parse(value);
      if (Array.isArray(parsed)) {
        return parsed.filter(Boolean).map(item => item.toString());
      }
    } catch (e) {
      return value.split(',').map(item => item.trim()).filter(Boolean);
    }
    return [];
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
   * 用户点击右上角分享
   */
  onShareAppMessage() {
    return {
      title: this.data.goodsTitle,
      path: '/pages/commodity/goods/index'
    };
  }
})
