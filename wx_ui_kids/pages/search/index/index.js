// pages/search/index.js
const app = getApp()

Page({

  /**
   * 页面的初始数据
   */
  data: {
    searchValue: '',
    hotSearches: ['羽绒服', '裤子', '厚睡衣', '软软壳', '成人装', '聪明屋穿', '内裤'],
    searchHistory: []
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {

  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 从本地存储获取搜索历史
    const history = wx.getStorageSync('searchHistory') || [];
    this.setData({
      searchHistory: history
    });
  },

  /**
   * 输入框内容变化时触发
   */
  onInputChange(e) {
    this.setData({
      searchValue: e.detail.value
    });
  },

  /**
   * 搜索按钮点击或键盘搜索键触发
   */
  onSearch() {
    const { searchValue } = this.data;
    if (!searchValue.trim()) return;

    // 保存搜索记录
    this.saveSearchHistory(searchValue);

    // 显示加载中
    wx.showLoading({
      title: '搜索中...',
    });

    // 获取app实例

    // 打印请求信息
      console.log('搜索请求URL:', '/commodity/search_style_codes');
      console.log('搜索请求体:', {"shopname": "youlan_kids", "search_keyword": searchValue, "page": 1, "page_size": 20});

      // 发送搜索请求
      app.req.post('/commodity/search_style_codes', {
        "shopname": "youlan_kids",
        "search_keyword": searchValue,
        "page": 1,
        "page_size": 20
      }, (apiRes) => {
        // 打印响应信息
        // console.log('搜索响应:', apiRes);
      // 隐藏加载
      wx.hideLoading();

      if (apiRes.code === 200) {
        // 跳转到搜索结果页面
        app.navigateTo({
          url: `/pages/search/display/index?keyword=${encodeURIComponent(searchValue)}`,
          success: function(navRes) {
            // 传递搜索结果数据
            navRes.eventChannel.emit('searchData', {
              data: (apiRes.data && apiRes.data.data) || [],
              total: (apiRes.data && apiRes.data.total) || 0,
              keyword: searchValue
            })
          }
        });
      } else {
        wx.showToast({
          title: '搜索失败，请重试',
          icon: 'none'
        });
      }
    }, (err) => {
      // 隐藏加载
      wx.hideLoading();
      wx.showToast({
        title: '网络错误，请重试',
        icon: 'none'
      });
    });
  },

  /**
   * 保存搜索历史
   */
  saveSearchHistory(keyword) {
    let history = wx.getStorageSync('searchHistory') || [];

    // 移除重复的记录
    history = history.filter(item => item !== keyword);

    // 添加新记录到开头
    history.unshift(keyword);

    // 只保留最近5条记录
    if (history.length > 5) {
      history = history.slice(0, 5);
    }

    // 保存到本地存储
    wx.setStorageSync('searchHistory', history);

    // 更新页面数据
    this.setData({
      searchHistory: history
    });
  },

  /**
   * 清除搜索内容
   */
  clearSearch() {
    this.setData({
      searchValue: ''
    });
  },

  /**
   * 取消搜索
   */
  onCancel() {
    app.navigateBack();
  },

  /**
   * 点击热门搜索或历史搜索标签
   */
  searchTag(e) {
    const keyword = e.currentTarget.dataset.keyword;
    this.setData({
      searchValue: keyword
    });
    this.onSearch();
  },

  /**
   * 清除搜索历史
   */
  clearHistory() {
    wx.showModal({
      title: '提示',
      content: '确定清除搜索历史？',
      success: (res) => {
        if (res.confirm) {
          wx.removeStorageSync('searchHistory');
          this.setData({
            searchHistory: []
          });
        }
      }
    });
  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {

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
