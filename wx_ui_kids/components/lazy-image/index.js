Component({
  properties: {
    src: {
      type: String,
      value: ''
    },
    mode: {
      type: String,
      value: 'aspectFill'
    },
    placeholder: {
      type: String,
      value: '/images/products.png'
    }
  },
  data: {
    loaded: false,
    currentSrc: ''
  },
  lifetimes: {
    attached() {
      this.setData({
        currentSrc: this.data.placeholder
      });
      this.initObserver();
    },
    detached() {
      if (this.observer) {
        this.observer.disconnect();
      }
    }
  },
  methods: {
    initObserver() {
      const that = this;
      this.observer = wx.createIntersectionObserver({
        initialAlpha: 0.01,
        thresholds: [0.01],
        observeAll: false
      });
      
      this.observer.relativeToViewport({
        bottom: 100,
        top: 100
      }).observe('.lazy-image-container', (res) => {
        if (res.isIntersecting && !that.data.loaded) {
          that.loadImage();
        }
      });
    },
    loadImage() {
      const that = this;
      if (this.data.src) {
        wx.getImageInfo({
          src: this.data.src,
          success() {
            that.setData({
              loaded: true,
              currentSrc: that.data.src
            });
          },
          fail() {
            that.setData({
              currentSrc: that.data.placeholder
            });
          }
        });
      }
    },
    onImageLoad() {
      this.setData({
        loaded: true
      });
    },
    onImageError() {
      this.setData({
        currentSrc: this.data.placeholder
      });
    }
  }
});
