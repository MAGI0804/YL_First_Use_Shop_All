const app = getApp();
const MAX_REVIEW_IMAGES = 9;

function getUserId() {
  const globalUserInfo = app.globalData.userInfo || {};
  return parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
}

Page({
  data: {
    orderId: '',
    subOrderId: '',
    commodityId: '',
    styleCode: '',
    productName: '',
    mode: 'create',
    reviewId: 0,
    rating: 5,
    content: '',
    uploadedImages: [],
    uploadingImage: false,
    tagOptions: ['质量好', '尺码合适', '面料舒服', '颜色好看', '发货快'],
    selectedTags: [],
    submitting: false
  },

  onLoad(options) {
    const mode = options.mode === 'edit' ? 'edit' : 'create';
    let selectedTags = [];
    if (mode === 'edit' && options.tags) {
      try {
        const parsed = JSON.parse(decodeURIComponent(options.tags));
        if (Array.isArray(parsed)) {
          selectedTags = parsed.filter(Boolean).map((item) => String(item));
        }
      } catch (error) {
        selectedTags = [];
      }
    }
    let uploadedImages = [];
    if (mode === 'edit' && options.images) {
      try {
        const parsed = JSON.parse(decodeURIComponent(options.images));
        if (Array.isArray(parsed)) {
          uploadedImages = parsed.filter(Boolean).map((item) => String(item));
        }
      } catch (error) {
        uploadedImages = [];
      }
    }
    this.setData({
      orderId: options.orderId || '',
      subOrderId: options.subOrderId || '',
      commodityId: options.commodityId || '',
      styleCode: options.styleCode || '',
      productName: decodeURIComponent(options.productName || ''),
      mode,
      reviewId: Number(options.reviewId || 0),
      rating: mode === 'edit' ? Number(options.rating || 5) : 5,
      content: mode === 'edit' ? decodeURIComponent(options.content || '') : '',
      uploadedImages,
      selectedTags
    });
    if (mode === 'edit') {
      wx.setNavigationBarTitle({ title: '修改评价' });
    }
  },

  selectRating(e) {
    const rating = Number(e.currentTarget.dataset.rating || 5);
    this.setData({ rating });
  },

  inputContent(e) {
    this.setData({
      content: e.detail.value
    });
  },

  toggleTag(e) {
    const tag = e.currentTarget.dataset.tag;
    const selectedTags = [...this.data.selectedTags];
    const index = selectedTags.indexOf(tag);
    if (index >= 0) {
      selectedTags.splice(index, 1);
    } else {
      selectedTags.push(tag);
    }
    this.setData({ selectedTags });
  },

  chooseImages() {
    const remain = MAX_REVIEW_IMAGES - this.data.uploadedImages.length;
    if (remain <= 0 || this.data.uploadingImage) {
      return;
    }
    wx.chooseMedia({
      count: remain,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        const files = Array.isArray(res.tempFiles) ? res.tempFiles : [];
        if (files.length === 0) return;
        this.uploadReviewImages(files.map(item => item.tempFilePath).filter(Boolean));
      }
    });
  },

  uploadReviewImages(paths) {
    const userId = getUserId();
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      return;
    }
    this.setData({ uploadingImage: true });
    wx.showLoading({ title: '上传中...' });
    Promise.all(paths.map(path => this.uploadReviewImage(path, userId)))
      .then((urls) => {
        const uploadedImages = this.data.uploadedImages.concat(urls.filter(Boolean)).slice(0, MAX_REVIEW_IMAGES);
        this.setData({ uploadedImages });
      })
      .catch((err) => {
        console.error('上传评价图片失败:', err);
        wx.showToast({
          title: err && err.message ? err.message : '图片上传失败',
          icon: 'none'
        });
      })
      .finally(() => {
        wx.hideLoading();
        this.setData({ uploadingImage: false });
      });
  },

  uploadReviewImage(filePath, userId) {
    return new Promise((resolve, reject) => {
      app.req.getAccessToken((token) => {
        wx.uploadFile({
          url: `${app.req.getHost()}/review/upload_image?access_token=${encodeURIComponent(token)}`,
          filePath,
          name: 'image',
          formData: {
            user_id: String(userId)
          },
          success: (res) => {
            let body = {};
            try {
              body = JSON.parse(res.data || '{}');
            } catch (error) {
              reject(new Error('图片上传响应异常'));
              return;
            }
            if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 200 && body.data && body.data.url) {
              resolve(body.data.url);
              return;
            }
            reject(new Error(body.msg || body.message || '图片上传失败'));
          },
          fail: reject
        });
      }, (err) => {
        reject(new Error(err && err.message ? err.message : '获取上传凭证失败'));
      });
    });
  },

  removeImage(e) {
    const index = Number(e.currentTarget.dataset.index);
    const uploadedImages = this.data.uploadedImages.slice();
    if (index >= 0 && index < uploadedImages.length) {
      uploadedImages.splice(index, 1);
      this.setData({ uploadedImages });
    }
  },

  previewImage(e) {
    const current = e.currentTarget.dataset.url;
    wx.previewImage({
      current,
      urls: this.data.uploadedImages
    });
  },

  submitReview() {
    const userId = getUserId();
    const { orderId, subOrderId, commodityId, styleCode, mode, reviewId, rating, content, uploadedImages, selectedTags, submitting, uploadingImage } = this.data;
    if (submitting) {
      return;
    }
    if (uploadingImage) {
      wx.showToast({
        title: '图片上传中',
        icon: 'none'
      });
      return;
    }
    if (!userId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      return;
    }
    if (mode === 'create' && (!orderId || !subOrderId || !commodityId)) {
      wx.showToast({
        title: '评价信息缺失',
        icon: 'none'
      });
      return;
    }
    if (!content || !content.trim()) {
      wx.showToast({
        title: '请填写评价内容',
        icon: 'none'
      });
      return;
    }

    const payload = mode === 'edit'
      ? {
          review_id: reviewId,
          user_id: userId,
          rating,
          content: content.trim(),
          images: uploadedImages,
          tags: selectedTags
        }
      : {
          user_id: userId,
          order_id: orderId,
          sub_order_id: subOrderId,
          commodity_id: commodityId,
          style_code: styleCode,
          rating,
          content: content.trim(),
          images: uploadedImages,
          tags: selectedTags
        };
    if (mode === 'edit' && !reviewId) {
      wx.showToast({
        title: '评价信息缺失',
        icon: 'none'
      });
      return;
    }

    this.setData({ submitting: true });
    wx.showLoading({ title: '提交中...' });
    app.req.post(mode === 'edit' ? '/review/update' : '/review/create', payload, (res) => {
      wx.hideLoading();
      this.setData({ submitting: false });
      if (res && res.code === 200) {
        wx.showToast({
          title: mode === 'edit' ? '评价已修改' : '评价已提交',
          icon: 'success'
        });
        setTimeout(() => {
          wx.navigateBack();
        }, 1200);
      } else {
        wx.showToast({
          title: res && res.msg ? res.msg : '提交失败',
          icon: 'none'
        });
      }
    }, (err) => {
      console.error('提交评价失败:', err);
      wx.hideLoading();
      this.setData({ submitting: false });
      wx.showToast({
        title: '网络请求失败',
        icon: 'none'
      });
    });
  }
});
