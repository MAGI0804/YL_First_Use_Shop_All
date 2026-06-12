// index.js
const app = getApp()
const IMAGE_PRELOAD_TIMEOUT = 8000

Page({
  data: {
    imageList: [],
    loading: true,
    error: false
  },
  requestId: 0,

  onLoad: function() {
    this.fetchActivityImages()
  },

  fetchActivityImages: function() {
    const requestId = ++this.requestId
    const that = this
    this.setData({
      loading: true,
      error: false
    })

    app.req.post(
      '/activity/query_online_activity_images',
      {
        shopname: 'youlan_kids'
      },
      function(res) {
        if (requestId !== that.requestId) return
        if (res.code === 200) {
          const dataArray = (res.data && Array.isArray(res.data.items)) ? res.data.items : 
                           (Array.isArray(res.data) ? res.data : []);
          
          const onlineImages = dataArray
            .filter(item => item.status === 'online')
            .map((item, index) => {
              return {
                ...item,
                id: item.id || index + 1,
                image: item.image || '/images/products.png'
              }
            })
            .sort((a, b) => {
              if (a.order !== null && a.order !== undefined && b.order !== null && b.order !== undefined) {
                return a.order - b.order;
              }
              if (a.order !== null && a.order !== undefined) {
                return -1;
              }
              if (b.order !== null && b.order !== undefined) {
                return 1;
              }
              return (a.id || 0) - (b.id || 0);
            })

          that.preloadActivityImages(onlineImages).then(() => {
            if (requestId !== that.requestId) return
            that.setData({
              imageList: onlineImages,
              loading: false,
              error: false
            })
          })
        } else {
          that.setData({
            loading: false,
            error: true
          })
          wx.showToast({
            title: '获取图片失败',
            icon: 'none'
          })
        }
      },
      function(err) {
        if (requestId !== that.requestId) return
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

  preloadActivityImages: function(images) {
    if (!Array.isArray(images) || images.length === 0) {
      return Promise.resolve()
    }
    return Promise.all(images.map(item => this.preloadImage(item.image))).then(() => {})
  },

  preloadImage: function(src) {
    return new Promise(resolve => {
      if (!src) {
        resolve()
        return
      }
      let settled = false
      let timer = null
      const done = () => {
        if (settled) return
        settled = true
        if (timer) {
          clearTimeout(timer)
        }
        resolve()
      }
      timer = setTimeout(done, IMAGE_PRELOAD_TIMEOUT)
      wx.getImageInfo({
        src,
        success: done,
        fail: done
      })
    })
  },

  onRetry: function() {
    this.fetchActivityImages()
  },
  onSearch: function() {
    app.navigateTo({
      url: '/pages/search/index/index'
    })
  },

  onImageTap: function(e) {
    const item = e.currentTarget.dataset.item
    if (item && item.has_activity_detail === true && item.id) {
      app.navigateTo({
        url: '/pages/activity/index/index?activity_id=' + item.id
      })
    }
  }
})
