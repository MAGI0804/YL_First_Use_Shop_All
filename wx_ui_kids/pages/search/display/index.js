// pages/search/display/index.js
const app = getApp();

Page({

  /**
   * 页面的初始数据
   */
  data: {
    keyword: '', // 搜索关键词
    goodsList: [], // 商品列表
    totalCount: 0, // 结果总数
    loading: false, // 是否加载中
    page: 1, // 当前页码
    pageSize: 20, // 每页条数
    isEmpty: false // 是否为空状态
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    // 获取搜索关键词
    const { keyword } = options;
    if (keyword) {
      this.setData({
        keyword: decodeURIComponent(keyword),
        page: 1,
        goodsList: []
      });

      // 初始化eventChannel接收搜索结果数据
      const eventChannel = this.getOpenerEventChannel();
      if (eventChannel) {
        eventChannel.on('searchData', (data) => {
          // 直接使用传递过来的搜索结果数据
          this.setData({
            goodsList: data.data || [],
            totalCount: data.total || 0,
            isEmpty: (data.total || 0) === 0
          });
        });
      } else {
        // 如果没有eventChannel，就重新加载数据
        this.loadGoodsData();
      }
    }
  },

  /**
   * 加载商品数据
   */
  loadGoodsData() {
    const { keyword, page, pageSize } = this.data;


    // 显示加载中
    wx.showLoading({
      title: '加载中...',
    });

    this.setData({
      loading: true
    });

    // 发送真实API请求
    app.req.post('/commodity/search_style_codes', {
      shopname: 'youlan_kids',
      search_keyword: keyword,
      page: page,
      page_size: pageSize
    }, (res) => {
      // 隐藏加载
      wx.hideLoading();

      if (res.code === 200 && res.data) {
        // 更新数据
        this.setData({
          goodsList: page === 1 ? (res.data.data || []) : [...this.data.goodsList, ...(res.data.data || [])],
          totalCount: res.data.total || 0,
          loading: false,
          page: page + 1,
          isEmpty: (res.data.total || 0) === 0
        });
      } else {
        this.setData({
          loading: false,
          isEmpty: true
        });
      }

      // 停止下拉刷新
      wx.stopPullDownRefresh();
    }, (err) => {
      // 隐藏加载
      wx.hideLoading();
      this.setData({
        loading: false,
        isEmpty: true
      });
      wx.showToast({
        title: '网络错误，请重试',
        icon: 'none'
      });

      // 停止下拉刷新
      wx.stopPullDownRefresh();
    });
  },

  /**
   * 跳转到商品详情页
   */
  navigateToDetail(e) {
    const { id } = e.currentTarget.dataset;
    if (id) {
      console.log('跳转到商品详情页，商品款号:', id);
      wx.navigateTo({
        url: `/pages/commodity/goods/index?id=${id}`
      });
    } else {
      console.error('商品款号不存在，无法跳转');
      wx.showToast({
        title: '商品不存在',
        icon: 'none'
      });
    }
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
    // 重置页码并重新加载数据
    this.setData({
      page: 1
    });
    this.loadGoodsData();
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {
    // 加载更多数据
    if (!this.data.loading) {
      this.loadGoodsData();
    }
  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {

  }
})