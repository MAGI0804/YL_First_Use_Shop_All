const app = getApp()

function getUserId() {
  const globalUserInfo = app.globalData.userInfo || {}
  return parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0)
}

function parseReviewList(value) {
  if (!value) return []
  if (Array.isArray(value)) return value.filter(Boolean).map(item => String(item))
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) return parsed.filter(Boolean).map(item => String(item))
  } catch (error) {
    // 兼容历史逗号分隔值
  }
  return String(value).split(',').map(item => item.trim()).filter(Boolean)
}

function statusText(status) {
  const map = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝',
    hidden: '已隐藏'
  }
  return map[status] || status || '-'
}

Page({
  data: {
    reviews: [],
    page: 1,
    pageSize: 10,
    total: 0,
    loading: false,
    hasMore: true
  },

  onShow() {
    this.refreshReviews()
  },

  onPullDownRefresh() {
    this.refreshReviews(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (!this.data.loading && this.data.hasMore) {
      this.loadReviews(false)
    }
  },

  refreshReviews(done) {
    this.setData({
      page: 1,
      reviews: [],
      total: 0,
      hasMore: true
    })
    this.loadReviews(true, done)
  },

  loadReviews(reset, done) {
    const userId = getUserId()
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      if (done) done()
      return
    }
    if (this.data.loading) {
      if (done) done()
      return
    }

    const nextPage = reset ? 1 : this.data.page
    this.setData({ loading: true })
    app.req.post('/review/query_mine', {
      user_id: userId,
      page: nextPage,
      page_size: this.data.pageSize
    }, (res) => {
      const rows = res && res.code === 200 && res.data && Array.isArray(res.data.data)
        ? res.data.data
        : []
      const normalizedRows = rows.map(item => ({
        ...item,
        statusText: statusText(item.status),
        displayTags: parseReviewList(item.tags),
        displayImages: parseReviewList(item.images),
        displayRating: '★★★★★'.slice(0, Number(item.rating || 0))
      }))
      const reviews = reset ? normalizedRows : this.data.reviews.concat(normalizedRows)
      const total = res && res.data ? Number(res.data.total || 0) : reviews.length
      this.setData({
        reviews,
        total,
        page: nextPage + 1,
        hasMore: reviews.length < total,
        loading: false
      })
      if (done) done()
    }, (err) => {
      console.error('查询我的评价失败:', err)
      this.setData({ loading: false })
      wx.showToast({
        title: '评价列表加载失败',
        icon: 'none'
      })
      if (done) done()
    })
  },

  editReview(e) {
    const review = this.data.reviews.find(item => item.id === Number(e.currentTarget.dataset.id))
    if (!review || review.status !== 'pending') return
    const tags = encodeURIComponent(JSON.stringify(review.displayTags || []))
    const images = encodeURIComponent(JSON.stringify(review.displayImages || []))
    app.navigateTo({
      url: `/pages/my/order/review/index?mode=edit&reviewId=${review.id}&rating=${review.rating || 5}&content=${encodeURIComponent(review.content || '')}&tags=${tags}&images=${images}&productName=${encodeURIComponent(review.commodity_id || '订单商品')}`
    })
  },

  deleteReview(e) {
    const reviewId = Number(e.currentTarget.dataset.id || 0)
    const userId = getUserId()
    if (!reviewId || !userId) return
    wx.showModal({
      title: '删除评价',
      content: '确定删除这条待审核评价吗？',
      confirmText: '删除',
      confirmColor: '#e64340',
      success: (modalRes) => {
        if (!modalRes.confirm) return
        wx.showLoading({ title: '删除中...' })
        app.req.post('/review/delete', {
          review_id: reviewId,
          user_id: userId
        }, (res) => {
          wx.hideLoading()
          if (res && res.code === 200) {
            wx.showToast({
              title: '已删除',
              icon: 'success'
            })
            this.refreshReviews()
          } else {
            wx.showToast({
              title: res && res.msg ? res.msg : '删除失败',
              icon: 'none'
            })
          }
        }, (err) => {
          console.error('删除评价失败:', err)
          wx.hideLoading()
          wx.showToast({
            title: '网络请求失败',
            icon: 'none'
          })
        })
      }
    })
  }
})
