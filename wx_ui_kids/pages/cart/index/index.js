// pages/cart/index/index.js
const app = getApp()

/**
 * 动态获取用户ID的辅助函数
 * 每次需要使用用户ID时都重新获取，确保使用最新的用户信息
 */
function getUserId() {
  const appInstance = getApp();
  const globalUserInfo = appInstance.globalData.userInfo || {};
  return globalUserInfo.user_id || appInstance.globalData.user_id || wx.getStorageSync('user_id') || '';
}

Page({

  /**
   * 页面的初始数据
   */
  data: {
    // 购物车商品列表，初始为空数组，等待真实API数据
    cartItems: [],
    loading: false,
    // 编辑状态
    isEditing: false,
    // 全选状态
    isAllSelected: false,
    // 已选商品数量
    selectedCount: 0,
    // 总价
    totalPrice: 0,
    // 格式化后的总价
    totalPriceFormatted: "0.00"
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 页面加载时不需要计算总价，因为还没有数据
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    this.loadCartItems()
  },

  /**
   * 加载购物车商品
   */
  loadCartItems() {
    // 获取最新的用户ID
    const userId = getUserId();
    
    // 检查用户ID是否存在
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      this.setData({
        cartItems: [],
        loading: false
      });
      return;
    }
    
    // 显示加载状态
    this.setData({
      loading: true,
      cartItems: [] // 清空购物车数据，等待真实数据返回
    })

    try {
      // 第一步：发送请求查询购物车内的商品ID列表
      app.req.post('/cart/query_cart_items', { user_id: userId },
        (res) => {
          console.log('购物车商品ID列表响应:', res);
          
          // 检查响应格式是否正确
          if (res.code === 200 && res.data && res.data.cart_items && Array.isArray(res.data.cart_items)) {
            const cartItemsList = res.data.cart_items;
            
            // 如果购物车为空
            if (cartItemsList.length === 0) {
              this.setData({
                cartItems: [],
                loading: false
              });
              this.calculateTotal();
              this.checkAllSelected();
              return;
            }
            
            // 创建临时的购物车商品列表
            const tempCartItems = cartItemsList.map(item => ({
              id: item.commodity_code || '',
              title: '加载中...',
              spec: '加载中...',
              price: 0.00,
              count: item.quantity || 1,
              image: '/images/products.png',
              selected: true
            }));

            // 提取商品ID列表用于下一步批量查询商品详情
            const commodityIds = cartItemsList.map(item => item.commodity_code);
            
            // 第二步：使用获取到的商品ID列表查询具体商品信息
            if (commodityIds.length > 0) {
              this.getProductsDetails(commodityIds, tempCartItems);
            } else {
              this.setData({
                cartItems: tempCartItems,
                loading: false
              });
              this.calculateTotal();
              this.checkAllSelected();
            }
          } else {
            // 响应格式不符合预期
            console.error('购物车ID列表响应格式不符合预期:', res);
            this.setData({
              loading: false,
              cartItems: []
            });
            this.calculateTotal();
            this.checkAllSelected();
          }
        },
        (err) => {
          // 查询购物车ID列表失败
          console.error('查询购物车ID列表失败:', err);
          this.setData({
            loading: false,
            cartItems: []
          });
          wx.showToast({
            title: '查询购物车失败',
            icon: 'none'
          });
          this.calculateTotal();
          this.checkAllSelected();
        }
      );
    } catch (error) {
      console.error('加载购物车过程中发生异常:', error);
      this.setData({
        loading: false,
        cartItems: []
      });
      this.calculateTotal();
      this.checkAllSelected();
    }
  },
  
  /**
   * 根据商品ID列表获取商品详情
   */
  getProductsDetails(commodityIds, tempCartItems) {
    // 显示加载状态
    this.setData({
      loading: true
    })

    // 构建请求数据，不进行额外转换，保持原始ID格式
    const requestData = {
      commodity_ids: commodityIds
    };
    
    console.log('请求商品详情的ID列表:', requestData.commodity_ids);
    
    // 第二步：使用商品ID列表查询具体商品信息
    app.req.post('/commodity/batch_get_products_by_ids', requestData, 
      (res) => {
        // 隐藏加载状态
        this.setData({
          loading: false
        })
        
        console.log('商品详情接口响应:', res);
        
        // 创建商品信息映射表
        const productMap = {};
        let cartItems = tempCartItems;
        
        // 处理成功响应
        if (res && res.code === 200 && res.data && res.data.data && Array.isArray(res.data.data)) {
          // 遍历商品详情数据，创建映射表
          res.data.data.forEach(product => {
            if (product && product.commodity_id) {
              productMap[product.commodity_id] = product;
            }
          });
          
          console.log('创建的商品映射表:', productMap);
          
          // 更新购物车中的商品信息
          cartItems = tempCartItems.map(item => {
            const realProduct = productMap[item.id];
            
            if (realProduct) {
              console.log('找到商品详情:', item.id, realProduct);
              // 确保价格是有效的数字
              const price = parseFloat(realProduct.price) || 0;
              
              // 构建规格信息
              const spec = `${realProduct.color || '颜色：随机'} ${realProduct.size || '尺码：均码'}`;
              
              return {
                id: item.id,
                title: realProduct.name || '商品名称',
                spec: spec,
                price: price,
                priceFormatted: price.toFixed(2),
                count: item.count || 1,
                image: realProduct.promo_image_url || realProduct.image || '/images/products.png',
                selected: item.selected,
                style_code: realProduct.style_code || '' // 保存style_code字段
              };
            }
            
            console.log('未找到商品详情，使用临时信息:', item.id);
            return {
              ...item,
              title: '商品信息加载失败',
              spec: '规格信息加载失败'
            };
          });
        } else {
          // 响应格式不符合预期或请求失败
          console.error('商品详情响应格式不符合预期:', res);
          
          // 请求商品详情失败，显示提示
          if (res && res.code !== 200) {
            wx.showToast({
              title: '获取商品信息失败',
              icon: 'none'
            });
          }
        }
        
        // 更新页面数据
        this.setData({
          cartItems: cartItems
        });
        
        this.calculateTotal();
        this.checkAllSelected();
      },
      (err) => {
        // 处理请求失败
        console.error('请求商品详情失败:', err);
        
        // 隐藏加载状态
        this.setData({
          loading: false
        })
        
        // 更新页面数据，即使商品详情获取失败，也显示购物车信息
        this.setData({
          cartItems: tempCartItems
        });
        
        this.calculateTotal();
        this.checkAllSelected();
        
        // 显示错误提示
        wx.showToast({
          title: '获取商品信息失败',
          icon: 'none'
        });
      }
    );
  },

  /**
   * 切换编辑状态
   */
  editCart() {
    this.setData({
      isEditing: !this.data.isEditing
    })
  },

  /**
   * 切换全选状态
   */
  toggleSelectAll() {
    const isAllSelected = !this.data.isAllSelected
    const cartItems = this.data.cartItems.map(item => ({
      ...item,
      selected: isAllSelected
    }))
    
    this.setData({
      isAllSelected,
      cartItems
    })
    
    this.calculateTotal()
  },

  /**
   * 切换单个商品选中状态
   */
  toggleSelectItem(e) {
    const id = e.currentTarget.dataset.id
    const cartItems = this.data.cartItems.map(item => {
      if (item.id === id) {
        return { ...item, selected: !item.selected }
      }
      return item
    })
    
    this.setData({
      cartItems
    })
    
    this.calculateTotal()
    this.checkAllSelected()
  },

  /**
   * 检查是否全选
   */
  checkAllSelected() {
    if (this.data.cartItems.length === 0) {
      this.setData({ isAllSelected: false })
      return
    }
    
    const isAllSelected = this.data.cartItems.every(item => item.selected)
    this.setData({ isAllSelected })
  },

  /**
   * 增加商品数量
   */
  increaseCount(e) {
    const id = e.currentTarget.dataset.id
    
    // 保存原始购物车数据，用于请求失败时恢复
    const originalCartItems = [...this.data.cartItems];
    
    // 先在本地更新数量，提升用户体验
    const cartItems = this.data.cartItems.map(item => {
      if (item.id === id) {
        return { ...item, count: item.count + 1 }
      }
      return item
    })
    
    // 立即更新UI
    this.setData({ cartItems })
    // 立即更新总价
    this.calculateTotal()
    
    // 获取最新的用户ID
    const userId = getUserId();
    
    // 检查用户ID是否存在
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      // 恢复原始数量
      this.setData({ cartItems: originalCartItems });
      this.calculateTotal();
      return;
    }
    
    // 发送请求到后端，更新商品数量
    app.req.post('/cart/increase_cart_item_quantity', {
      user_id: userId,
      commodity_code: id
    }, (res) => {
      // 处理成功响应
      if (res && res.code === 200 && res.data) {
        // 根据返回的quantity更新商品数量，确保与后端数据一致
        const updatedCartItems = this.data.cartItems.map(item => {
          if (item.id === id) {
            // 使用后端返回的数量，确保数据一致性
            return { ...item, count: res.data.quantity || item.count }
          }
          return item
        })
        
        this.setData({ cartItems: updatedCartItems })
        // 确保总价是最新的
        this.calculateTotal()
        
      } else {
        // 如果后端返回失败，恢复之前的数量
        this.setData({ cartItems: originalCartItems });
        this.calculateTotal();
        wx.showToast({
          title: '更新数量失败',
          icon: 'none'
        })
      }
    }, (err) => {
      // 处理请求失败
      console.error('增加商品数量失败:', err)
      // 请求失败，恢复之前的数量
      this.setData({ cartItems: originalCartItems });
      this.calculateTotal();
      wx.showToast({
        title: '更新数量失败',
        icon: 'none'
      })
    })
  },

  /**
   * 减少商品数量
   */
  decreaseCount(e) {
    const id = e.currentTarget.dataset.id
    
    // 保存原始购物车数据，用于请求失败时恢复
    const originalCartItems = [...this.data.cartItems];
    
    // 先在本地更新数量，提升用户体验
    const cartItems = this.data.cartItems.map(item => {
      if (item.id === id && item.count > 1) {
        return { ...item, count: item.count - 1 }
      }
      return item
    })
    
    // 立即更新UI
    this.setData({ cartItems })
    // 立即更新总价
    this.calculateTotal()
    
    // 获取最新的用户ID
    const userId = getUserId();
    
    // 检查用户ID是否存在
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      // 恢复原始数量
      this.setData({ cartItems: originalCartItems });
      this.calculateTotal();
      return;
    }
    
    // 发送请求到后端，更新商品数量
    app.req.post('/cart/decrease_cart_item_quantity', {
      user_id: userId,
      commodity_code: id
    }, (res) => {
      // 处理成功响应
      if (res && res.code === 200 && res.data) {
        // 根据返回的quantity更新商品数量，确保与后端数据一致
        const updatedCartItems = this.data.cartItems.map(item => {
          if (item.id === id) {
            // 使用后端返回的数量，确保数据一致性
            return { ...item, count: res.data.quantity || item.count }
          }
          return item
        })
        
        this.setData({ cartItems: updatedCartItems })
        // 确保总价是最新的
        this.calculateTotal()
        
      } else {
        // 如果后端返回失败，恢复之前的数量
        this.setData({ cartItems: originalCartItems });
        this.calculateTotal();
        wx.showToast({
          title: '更新数量失败',
          icon: 'none'
        })
      }
    }, (err) => {
      // 处理请求失败
      console.error('减少商品数量失败:', err)
      // 请求失败，恢复之前的数量
      this.setData({ cartItems: originalCartItems });
      this.calculateTotal();
      wx.showToast({
        title: '更新数量失败',
        icon: 'none'
      })
    })
  },

  /**
   * 计算总价和已选数量
   */
  calculateTotal() {
    // 过滤出已选中的商品
    const selectedItems = this.data.cartItems.filter(item => item.selected)
    
    // 计算总价，添加防御性检查确保价格是有效数字
    const totalPrice = selectedItems.reduce((sum, item) => {
      // 确保price和count都是有效数字
      const price = typeof item.price === 'number' && !isNaN(item.price) ? item.price : 0;
      const count = typeof item.count === 'number' && !isNaN(item.count) && item.count > 0 ? item.count : 0;
      return sum + price * count;
    }, 0)
    
    // 计算已选数量
    const selectedCount = selectedItems.reduce((sum, item) => {
      const count = typeof item.count === 'number' && !isNaN(item.count) && item.count > 0 ? item.count : 0;
      return sum + count;
    }, 0)
    
    // 添加格式化后的价格字段，用于WXML显示
    const totalPriceFormatted = totalPrice.toFixed(2)
    
    this.setData({
      totalPrice,
      totalPriceFormatted,
      selectedCount
    })
  },

  /**
   * 批量删除
   */
  batchDelete() {
    if (this.data.selectedCount === 0) {
      wx.showToast({
        title: '请先选择要删除的商品',
        icon: 'none'
      })
      return
    }
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除所选商品吗？',
      success: (res) => {
        if (res.confirm) {
          // 获取选中的商品ID列表
          const selectedItems = this.data.cartItems.filter(item => item.selected)
          const commodityCodes = selectedItems.map(item => item.id)
        
          
          // 获取最新的用户ID
          const userId = getUserId();
          
          // 检查用户ID是否存在
          if (!userId) {
            wx.showToast({
              title: '请先登录',
              icon: 'none'
            });
            return;
          }
          
          // 发送请求到后端，批量删除购物车商品
          app.req.post('/cart/batch_delete_from_cart', {
            user_id: userId,
            commodity_codes: commodityCodes
          }, (res) => {
            // 处理成功响应
            if (res && res.code === 200) {
              // 根据返回结果更新购物车数据
              const cartItems = this.data.cartItems.filter(item => !item.selected)
              this.setData({
                cartItems,
                isAllSelected: false
              })
              this.calculateTotal()
              
              // 显示成功提示
              wx.showToast({
                title: res.message || '删除成功',
                icon: 'success'
              })
            } else {
              // 显示失败提示
              wx.showToast({
                title: res.message || '删除失败',
                icon: 'none'
              })
            }
          }, (err) => {
            // 处理请求失败
            console.error('批量删除商品失败:', err)
            wx.showToast({
              title: '网络错误，删除失败',
              icon: 'none'
            })
          })
        }
      }
    })
  },

  /**
   * 跳转到商品详情页
   */
  goToProductDetail(e) {
    const { style_code, id } = e.currentTarget.dataset
    // 优先使用style_code跳转，如果没有则回退到id
    const targetId = style_code || id
    wx.navigateTo({
      url: `/pages/commodity/goods/index?id=${targetId}`
    })
  },

  /**
   * 批量下单
   */
  batchCheckout() {
    if (this.data.selectedCount === 0) {
      wx.showToast({
        title: '请先选择要购买的商品',
        icon: 'none'
      })
      return
    }
    
    // 确保总价是最新的
    this.calculateTotal()
    
    // 获取选中的商品
    const selectedItems = this.data.cartItems.filter(item => item.selected)
    
    // 将选中的商品信息传递给订单确认页面
    wx.navigateTo({
      url: '/pages/cart/buy_order/index',
      success: (res) => {
        res.eventChannel.emit('selectedItems', {
          items: selectedItems,
          totalPrice: this.data.totalPrice
        })
      }
    })
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