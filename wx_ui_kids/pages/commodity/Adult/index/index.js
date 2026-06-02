// pages/commodity/index/index.js
const app = getApp()

const getProductImage = (item) => {
  if (!item) return '/images/products.png'
  if (item.promo_image_url) return item.promo_image_url
  if (item.promo_image) return item.promo_image
  if (item.image_url) return item.image_url
  if (item.image) return item.image
  if (item.main_image && item.main_image.url) return item.main_image.url
  if (Array.isArray(item.images) && item.images.length > 0) {
    const mainImage = item.images.find(image => image && image.is_main && image.url)
    return (mainImage && mainImage.url) || (item.images[0] && item.images[0].url) || '/images/products.png'
  }
  return '/images/products.png'
}

const normalizeProducts = (products) => {
  return (products || []).map(item => ({
    ...item,
    imageUrl: getProductImage(item)
  }))
}

Page({
  /**
   * 页面的初始数据
   */
  data: {
    categories: [], // 存储所有类目
    selectedCategory: '', // 选中的类目
    products: [], // 商品列表
    loading: true, // 加载状态
    loadingMore: false, // 底部加载更多状态
    error: false, // 错误状态
    page: 1, // 当前页码
    pageSize: 10, // 每页数量
    total: 0, // 总数量
    hasMore: true, // 是否有更多数据
    refresherTriggered: false, // 下拉刷新状态
    // 标签数据
    labels: {
      label_one: [],
      label_two: [],
      label_three: [],
      label_four: [],
      label_seven: [],
    },
    // 选中的label_two
    selectedLabelTwo: [], // 多选
    menuItems: [], // 处理后的菜单数据
    
    // 筛选相关数据
    showFilterModal: false, // 是否显示筛选模态框
    selectedLabelOne: [], // 选中的label_one（年份）
    selectedLabelFour: [], // 选中的label_four（服装类型）
    selectedLabelSeven: [], // 选中的label_seven（具体类别）
    currentFilters: null, // 当前生效的筛选条件
    filterItems: {} // 处理后的筛选数据
  },
  navigating: false,

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
    // 只获取类目信息，不获取标签
    this.fetchAllCategories()
  },

  /**
   * 检查数组是否包含指定元素
   */
  arrayContains(array, item) {
    return array && array.indexOf(item) >= 0;
  },

  /**
   * 获取处理后的菜单数据
   */
  getProcessedMenuItems() {
    const menuItems = [];
    const labelTwoList = this.data.labels.label_two || [];
    const selectedList = this.data.selectedLabelTwo || [];
    
    labelTwoList.forEach((item, index) => {
      menuItems.push({
        label: item,
        isSelected: selectedList.indexOf(item) >= 0,
        index: index
      });
    });
    

    return menuItems;
  },

  /**
   * 获取处理后的筛选数据
   */
  getProcessedFilterItems() {
    const filterItems = {
      labelOne: [],
      labelFour: [],
      labelSeven: []
    };
    
    // 处理label_one（年份）
    if (this.data.labels.label_one) {
      filterItems.labelOne = this.data.labels.label_one.map(item => ({
        label: item,
        isSelected: this.data.selectedLabelOne.indexOf(item) >= 0
      }));
    }
    
    // 处理label_four（服装类型）
    if (this.data.labels.label_four) {
      filterItems.labelFour = this.data.labels.label_four.map(item => ({
        label: item,
        isSelected: this.data.selectedLabelFour.indexOf(item) >= 0
      }));
    }
    
    // 处理label_seven（具体类别）
    if (this.data.labels.label_seven) {
      filterItems.labelSeven = this.data.labels.label_seven.map(item => ({
        label: item,
        isSelected: this.data.selectedLabelSeven.indexOf(item) >= 0
      }));
    }
    
    return filterItems;
  },

  /**
   * 打开筛选模态框
   */
  openFilterModal() {
    this.setData({
      showFilterModal: true
    });
  },

  /**
   * 关闭筛选模态框
   */
  closeFilterModal() {
    this.setData({
      showFilterModal: false
    });
  },

  /**
   * 切换label_one（年份）选中状态
   */
  toggleLabelOne(e) {
    const item = e.currentTarget.dataset.item;
    let selectedLabelOne = [...this.data.selectedLabelOne];
    
    const index = selectedLabelOne.indexOf(item);
    if (index === -1) {
      selectedLabelOne.push(item);
    } else {
      selectedLabelOne.splice(index, 1);
    }
    
    this.setData({
      selectedLabelOne: selectedLabelOne
    }, () => {
      // 更新筛选数据
      const filterItems = this.getProcessedFilterItems();
      this.setData({
        filterItems: filterItems
      });
    });
  },

  /**
   * 切换label_four（服装类型）选中状态
   */
  toggleLabelFour(e) {
    const item = e.currentTarget.dataset.item;
    let selectedLabelFour = [...this.data.selectedLabelFour];
    
    const index = selectedLabelFour.indexOf(item);
    if (index === -1) {
      selectedLabelFour.push(item);
    } else {
      selectedLabelFour.splice(index, 1);
    }
    
    this.setData({
      selectedLabelFour: selectedLabelFour
    }, () => {
      // 更新筛选数据
      const filterItems = this.getProcessedFilterItems();
      this.setData({
        filterItems: filterItems
      });
    });
  },

  /**
   * 切换label_seven（具体类别）选中状态
   */
  toggleLabelSeven(e) {
    const item = e.currentTarget.dataset.item;
    let selectedLabelSeven = [...this.data.selectedLabelSeven];
    
    const index = selectedLabelSeven.indexOf(item);
    if (index === -1) {
      selectedLabelSeven.push(item);
    } else {
      selectedLabelSeven.splice(index, 1);
    }
    
    this.setData({
      selectedLabelSeven: selectedLabelSeven
    }, () => {
      // 更新筛选数据
      const filterItems = this.getProcessedFilterItems();
      this.setData({
        filterItems: filterItems
      });
    });
  },

  /**
   * 重置所有筛选条件
   */
  resetFilters() {
    this.setData({
      selectedLabelOne: [],
      selectedLabelFour: [],
      selectedLabelSeven: [],
      currentFilters: null // 清空当前筛选条件
    }, () => {
      // 更新筛选数据
      const filterItems = this.getProcessedFilterItems();
      this.setData({
        filterItems: filterItems
      });
      
      // 重新加载原始商品数据（不带筛选条件）
      this.reloadOriginalProducts();
    });
  },

  /**
   * 重新加载原始商品数据（不带筛选条件）
   */
  reloadOriginalProducts() {
    this.setData({
      loading: true,
      products: [],
      page: 1,
      hasMore: true
    });
    
    // 重新加载当前分类的商品
    this.fetchProductsByCategory(this.data.selectedCategory);
  },

  /**
   * 确认筛选条件并发送请求
   */
  confirmFilters() {
    const { selectedLabelOne, selectedLabelFour, selectedLabelSeven } = this.data;
    
    // 关闭模态框
    this.setData({
      showFilterModal: false
    });
    
    // 检查是否有选择任何筛选条件
    const hasFilters = selectedLabelOne.length > 0 || selectedLabelFour.length > 0 || selectedLabelSeven.length > 0;
    
    if (!hasFilters) {
      // 如果没有筛选条件，直接重置并加载原始数据
      this.setData({
        currentFilters: null,
        selectedLabelTwo: [] // 清空季节筛选
      }, () => {
        this.reloadOriginalProducts();
      });
      return;
    }
    
    // 显示加载提示
    wx.showLoading({
      title: '筛选中...',
      mask: true
    });
    
    // 保存当前筛选条件，并重置分页
    this.setData({
      currentFilters: {
        labelOne: selectedLabelOne,
        labelFour: selectedLabelFour,
        labelSeven: selectedLabelSeven
      },
      page: 1, // 重置分页
      selectedLabelTwo: [] // 清空季节筛选
    }, () => {
      // 更新菜单数据
      const menuItems = this.getProcessedMenuItems();
      this.setData({
        menuItems: menuItems
      });
    });
    
    // 发送筛选请求
    this.fetchProductsByFilters({
      labelOne: selectedLabelOne,
      labelFour: selectedLabelFour,
      labelSeven: selectedLabelSeven
    });
  },

  /**
   * 根据筛选条件获取商品
   */
  fetchProductsByFilters(filters) {
    const { pageSize } = this.data;
    
    this.setData({
      loading: true,
      products: [],
      page: 1, // 重置为第一页
      hasMore: true
    });

    // 构建请求数据，在原有参数基础上添加筛选条件
    const requestData = {
      shopname: 'youlan_kids',
      category: this.data.selectedCategory, // 使用当前选中的分类
      page: 1, // 始终从第一页开始
      page_size: pageSize,
      demand: 'style_code',
      label_three: ["成人"],
      label_two: this.data.selectedLabelTwo, // 保留季节筛选
      label_one: filters.labelOne, // 年份筛选
      label_four: filters.labelFour, // 服装类型筛选
      label_seven: filters.labelSeven // 具体类别筛选
    };

    console.log('筛选请求数据:', requestData);
    
    return new Promise((resolve, reject) => {
      app.req.post(
        '/commodity/goods_query_wx',
        requestData,
        (res) => {
          // 隐藏加载提示
          wx.hideLoading();
          
          if (res.code === 200) {
            const newProducts = normalizeProducts((res.data && res.data.data) ? res.data.data : []);
            const totalCount = res.data.total || newProducts.length; // 使用total或当前数量
            
            this.setData({
              products: newProducts,
              loading: false,
              hasMore: newProducts.length >= pageSize,
              total: totalCount
            });
            
            wx.showToast({
              title: `找到 ${totalCount} 件商品`,
              icon: 'success',
              duration: 2000
            });
            
            resolve(res);
          } else {
            wx.showToast({
              title: '获取商品失败: ' + res.message,
              icon: 'none'
            });
            reject(res);
          }
        },
        (err) => {
          // 隐藏加载提示
          wx.hideLoading();
          
          this.setData({
            loading: false
          });
          console.error('筛选请求失败:', err);
          wx.showToast({
            title: '网络异常',
            icon: 'none'
          });
          reject(err);
        }
      );
    });
  },

  /**
   * 根据筛选条件获取商品（带页码参数，用于加载更多）
   */
  fetchProductsByFiltersWithPage(filters, pageNum, options = {}) {
    const { pageSize } = this.data;
    const isRefresh = options.refresh === true;
    
    // 构建请求数据，在原有参数基础上添加筛选条件
    const requestData = {
      shopname: 'youlan_kids',
      category: this.data.selectedCategory, // 使用当前选中的分类
      page: pageNum,
      page_size: pageSize,
      demand: 'style_code',
      label_three: ["成人"],
      label_two: this.data.selectedLabelTwo, // 保留季节筛选
      label_one: filters.labelOne, // 年份筛选
      label_four: filters.labelFour, // 服装类型筛选
      label_seven: filters.labelSeven // 具体类别筛选
    };

    console.log('筛选加载更多请求数据:', requestData);
    
    return new Promise((resolve, reject) => {
      app.req.post(
        '/commodity/goods_query_wx',
        requestData,
        (res) => {
          if (res.code === 200) {
            const newProducts = normalizeProducts((res.data && res.data.data) ? res.data.data : []);
            const totalCount = res.data.total || newProducts.length;
            const allProducts = pageNum === 1 ? newProducts : [...this.data.products, ...newProducts]; // 累积商品数据
            
            this.setData({
              products: allProducts,
              page: pageNum,
              loadingMore: false,
              refresherTriggered: false,
              hasMore: newProducts.length >= pageSize,
              total: totalCount
            });
            
            resolve(res);
          } else {
            this.setData({
              loadingMore: false,
              refresherTriggered: false
            });
            wx.showToast({
              title: '加载更多失败: ' + res.message,
              icon: 'none'
            });
            reject(res);
          }
        },
        (err) => {
          this.setData({
            loadingMore: false,
            refresherTriggered: false
          });
          console.error('筛选加载更多请求失败:', err);
          wx.showToast({
            title: '网络异常',
            icon: 'none'
          });
          reject(err);
        }
      );
    });
  },

  /**
   * 获取指定类别的标签
   */
  fetchAllLabels(category = '全部') {
    const that = this
    this.setData({
      error: false
    })

    return new Promise((resolve, reject) => {
      app.req.post(
        '/commodity/get_all_labels',
        {
          shopname: 'youlan_kids',
          label_three: ["成人"],
          category: category // 添加类别参数
        },
        function(res) {
          if (res.code === 200) {
            // 记录所有的标签
            that.setData({
              labels: {
                label_one: res.data.label_one || [],
                label_two: res.data.label_two || [],
                label_three: res.data.label_three || [],
                label_four: res.data.label_four || [],
                label_seven: res.data.label_seven || []
              }
            })
            
            const menuItems = that.getProcessedMenuItems();
            const filterItems = that.getProcessedFilterItems();
            that.setData({
              menuItems: menuItems,
              filterItems: filterItems
            });
            
            resolve();
          } else {
            that.setData({
              labels: {
                label_one: [],
                label_two: [],
                label_three: [],
                label_four: [],
                label_seven: []
              },
              menuItems: [],
              filterItems: {}
            })
            wx.showToast({
              title: '获取标签失败: ' + res.msg,
              icon: 'none'
            })
            resolve();
          }
        },
        function(err) {
          console.error('请求标签失败:', err)
          that.setData({
            menuItems: [],
            filterItems: {}
          })
          wx.showToast({
            title: '网络异常',
            icon: 'none'
          })
          resolve();
        }
      );
    });
  },

  /**
   * 获取所有类目
   */
  fetchAllCategories() {
    const that = this
    this.setData({
      loading: true,
      error: false
    })

    app.req.post(
      '/commodity/get_all_categories',
      {
        shopname: 'youlan_kids',
        "label_three": ["成人"]
      },
      function(res) {
        if (res.code === 200) {
          // 将获取到的类目数据转换为合适的格式
          const categories = (res.data && Array.isArray(res.data.categories)) ? 
            res.data.categories.map((item, index) => ({
              id: index + 1,
              name: item
            })) : []

          that.setData({
            categories: categories,
            loading: false
          })

          // 如果有类目，默认选中第一个
          if (categories.length > 0) {
            const firstCategory = categories[0].name;
            that.setData({
              selectedCategory: firstCategory
            })
            
            // 先获取第一个类别的标签，然后加载商品
            that.fetchAllLabels(firstCategory).then(() => {
              that.fetchProductsByCategory(firstCategory);
            });
          } else {
            that.setData({
              products: [],
              hasMore: false
            })
          }
        } else {
          that.setData({
            loading: false,
            error: true
          })
          wx.showToast({
            title: '获取类目失败: ' + res.message,
            icon: 'none'
          })
        }
      },
      function(err) {
        console.error('请求失败:', err)
        that.setData({
          loading: false,
          error: true
        })
        wx.showToast({
          title: '网络异常',
          icon: 'none'
        })
      }
    )
  },

  /**
   * 点击切换类目
   */
  onCategoryClick(e) {
    const category = e.currentTarget.dataset.category
    if (category !== this.data.selectedCategory) {
      this.setData({
        selectedCategory: category,
        products: [],
        page: 1,
        hasMore: true,
        // 重置筛选状态
        currentFilters: null,
        selectedLabelOne: [],
        selectedLabelFour: [],
        selectedLabelSeven: [],
        selectedLabelTwo: []
      }, () => {
        // 更新菜单和筛选数据
        const menuItems = this.getProcessedMenuItems();
        const filterItems = this.getProcessedFilterItems();
        this.setData({
          menuItems: menuItems,
          filterItems: filterItems
        });
      })
      
      // 先获取当前类别的标签，再加载商品
      this.fetchAllLabels(category).then(() => {
        this.fetchProductsByCategory(category);
      });
    }
  },

  /**
   * 跳转到商品详情页
   */
  navigateToGoodsDetail(e) {
    if (this.navigating) return
    const id = e.currentTarget.dataset.id
    if (!id) return
    this.navigating = true
    wx.navigateTo({
      url: `/pages/commodity/goods/index?id=${id}`,
      complete: () => {
        setTimeout(() => {
          this.navigating = false
        }, 500)
      }
    })
  },

  /**
   * 点击选择label_two
   */
  onLabelTwoClick(e) {
    const label = e.currentTarget.dataset.label
    let selectedLabelTwo = [...this.data.selectedLabelTwo]
    
    // 实现多选逻辑
    const index = selectedLabelTwo.indexOf(label)
    if (index === -1) {
      // 未选中，添加到数组
      selectedLabelTwo.push(label)
    } else {
      // 已选中，从数组中移除
      selectedLabelTwo.splice(index, 1)
    }
    

    
    this.setData({
      selectedLabelTwo: selectedLabelTwo,
      products: [],
      page: 1,
      hasMore: true
    }, () => {
      // 更新菜单数据
      const menuItems = this.getProcessedMenuItems();
      this.setData({
        menuItems: menuItems
      });
    })
    
    // 重新加载商品数据
    this.fetchProductsByCategory(this.data.selectedCategory)
  },

  /**
   * 检查label_two是否选中
   */
  isLabelTwoSelected(label) {
    return this.data.selectedLabelTwo.indexOf(label) !== -1
  },

  /**
   * 根据类目获取商品
   */
  fetchProductsByCategory(category, pageNum, options = {}) {
    const that = this
    const isRefresh = options.refresh === true
    const requestPage = pageNum || this.data.page
    // 根据当前页码决定使用哪个加载状态
    const isInitialLoad = requestPage === 1 && !isRefresh
    this.setData({
      ...(isRefresh ? {} : { [isInitialLoad ? 'loading' : 'loadingMore']: true }),
      error: false
    })

    // 使用传入的分类参数，而不是硬编码的"全部"
    const requestCategory = category || '全部';
    
    const requestData = {
      shopname: 'youlan_kids',
      category: requestCategory,
      page: requestPage,
      page_size: this.data.pageSize,
      demand:'style_code',
      "label_three": ["成人"],
      "label_two": this.data.selectedLabelTwo
    }
    
    console.log('分类加载请求数据:', requestData);

    return new Promise((resolve, reject) => {
      app.req.post(
        '/commodity/goods_query_wx',
        requestData,
        function(res) {
          if (res.code === 200) {
            const newProducts = normalizeProducts((res.data && res.data.data) ? res.data.data : [])
            const allProducts = requestPage === 1 ? newProducts : [...that.data.products, ...newProducts]

            that.setData({
              products: allProducts,
              page: requestPage,
              loading: false,
              loadingMore: false,
              refresherTriggered: false,
              total: res.data.total || 0,
              hasMore: allProducts.length < (res.data.total || 0)
            })
            resolve()
          } else {
            that.setData({
              loading: false,
              loadingMore: false,
              refresherTriggered: false,
              error: true
            })
            wx.showToast({
              title: '获取商品失败',
              icon: 'none'
            })
            resolve()
          }
        },
        function(err) {
          console.error('请求失败:', err)
          that.setData({
            loading: false,
            loadingMore: false,
            refresherTriggered: false,
            error: true
          })
          wx.showToast({
            title: '网络异常',
            icon: 'none'
          })
          resolve()
        }
      )
    })
  },

  /**
   * 加载更多商品
   */
  loadMoreProducts() {
    if (this.data.hasMore && !this.data.loading && !this.data.loadingMore) {
      const newPage = this.data.page + 1;
      
      this.setData({
        loadingMore: true
      })
      
      // 检查是否有生效的筛选条件
      if (this.data.currentFilters) {
        // 使用筛选API加载更多
        this.fetchProductsByFiltersWithPage(this.data.currentFilters, newPage);
      } else {
        // 使用原有的商品加载API
        this.fetchProductsByCategory(this.data.selectedCategory, newPage);
      }
    }
  },

  refreshProducts() {
    if (this.data.refresherTriggered || this.data.loadingMore) return

    this.setData({
      refresherTriggered: true,
      page: 1,
      hasMore: true
    }, () => {
      if (this.data.currentFilters) {
        this.fetchProductsByFiltersWithPage(this.data.currentFilters, 1, { refresh: true })
      } else {
        this.fetchProductsByCategory(this.data.selectedCategory, 1, { refresh: true })
      }
    })
  },

  /**
   * 重新加载类目数据
   */
  reloadCategories() {
    this.fetchAllCategories()
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {

  },

  /**
   * 下拉刷新事件处理
   */
  onPullDownRefresh() {
    this.refreshProducts()
  },

  /**
   * 滚动到底部事件处理
   */
  onScrollToLower() {
    this.loadMoreProducts()
  }
})
