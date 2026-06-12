// pages/activity/index/index.js
const app = getApp()

Page({
  data: {
    activityId: null,
    activityImages: [],
    productList: [],
    loading: true
  },

  onLoad(options) {
    if (options.activity_id) {
      this.setData({
        activityId: options.activity_id
      })
      this.getActivityDetail(options.activity_id)
    } else {
      this.setData({
        loading: false
      })
      wx.showToast({
        title: '活动ID不存在',
        icon: 'none'
      })
    }
  },

  onShow() {

  },

  getActivityDetail(activityId) {
    this.setData({
      loading: true
    })

    app.req.post('/activity/get_activity_image_detail', { activity_id: parseInt(activityId) },
      (res) => {
        // console.log('活动详情响应:', res)

        if (res.code === 200 && res.data) {
          const data = res.data
          const activityImages = []
          const promotionalPics = data.promotional_pics

          if (promotionalPics) {
            const keys = Object.keys(promotionalPics).sort((a, b) => a - b)
            keys.forEach(key => {
              if (promotionalPics[key] && promotionalPics[key].image_url) {
                activityImages.push(promotionalPics[key].image_url)
              }
            })
          }

          if (data.commodities && Array.isArray(data.commodities) && data.commodities.length > 0) {
            const productList = data.commodities.map(item => ({
              id: item.id || item.commodity_id || '',
              style_code: item.style_code || '',
              image: item.image || item.promo_image_url || '',
              name: item.name || '',
              price: item.price || '0.00'
            }))

            this.setData({
              activityImages,
              productList,
              loading: false
            })
          } else if (data.style_codes && Array.isArray(data.style_codes) && data.style_codes.length > 0) {
            this.getProductDetailsByStyleCodes(data.style_codes, activityImages)
          } else {
            // console.log('没有商品数据')
            this.setData({
              activityImages,
              productList: [],
              loading: false
            })
          }
        } else {
          this.setData({
            loading: false
          })
          wx.showToast({
            title: res.msg || '获取活动详情失败',
            icon: 'none'
          })
        }
      },
      (err) => {
        // console.error('获取活动详情失败:', err)
        this.setData({
          loading: false
        })
        wx.showToast({
          title: '网络错误',
          icon: 'none'
        })
      }
    )
  },

  getProductDetailsByStyleCodes(styleCodes, activityImages) {
    const productList = []
    let completedCount = 0

    styleCodes.forEach((styleCode, index) => {
      app.req.post('/commodity/stylecode_commodities', 
        {
          style_code: styleCode,
          shopname: 'youlan_kids'
        },
        (res) => {
          // console.log('商品详情响应 ' + styleCode + ':', res)

          if (res && res.code === 200 && res.data) {
            const goodsData = res.data
            let image = '/images/products.png'
            
            if (goodsData.images && Array.isArray(goodsData.images) && goodsData.images.length > 0) {
              const mainImage = goodsData.images.find(img => img.is_main)
              if (mainImage && mainImage.url) {
                image = mainImage.url
              } else {
                image = goodsData.images[0].url
              }
            }

            productList.push({
              id: goodsData.id || styleCode,
              style_code: styleCode,
              image: image,
              name: goodsData.name || '',
              price: goodsData.price || '0.00'
            })
          }

          completedCount++
          if (completedCount === styleCodes.length) {
            // console.log('处理后的商品列表:', productList)
            this.setData({
              activityImages,
              productList,
              loading: false
            })

            if (productList.length === 0) {
              wx.showToast({
                title: '暂无商品',
                icon: 'none'
              })
            }
          }
        },
        (err) => {
          // console.error('获取商品详情失败 ' + styleCode + ':', err)
          completedCount++
          if (completedCount === styleCodes.length) {
            this.setData({
              activityImages,
              productList,
              loading: false
            })
          }
        }
      )
    })
  },

  onPullDownRefresh() {
    if (this.data.activityId) {
      this.getActivityDetail(this.data.activityId)
    }
    setTimeout(() => {
      wx.stopPullDownRefresh()
    }, 500)
  },

  onReachBottom() {

  },

  onShareAppMessage() {
    return {
      title: '精彩活动来袭',
      path: `/pages/activity/index/index?activity_id=${this.data.activityId}`
    }
  },

  goProductDetail(e) {
    const { id, style_code } = e.currentTarget.dataset
    const targetId = style_code || id
    if (targetId) {
      app.navigateTo({
        url: `/pages/commodity/goods/index?id=${targetId}`
      })
    }
  }
})
