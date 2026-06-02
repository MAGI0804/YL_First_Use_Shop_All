Component({
  properties: {
    // cancel | timeout | exchange | recycle | address | xingqiubi1 | laba | xingqiubi | wancheng | a- | point | ticket | cash | bonus | captcha | order | X-38 | wanchenghuishou | mianfeishangmen | yijianyuyue | duihuanxinpin | xiangyoujiantou | pants1 | hengY | mark | stepIn | nodata | logo1 | code | plant | logo | step2 | step3 | sheet | edit | addr | share | track | pants | l | step1 | time | pack | bird | copy | score | shoes | refrash | score2 | list | t-shirt | doc | checked | size | del | hat | vip | weight | recovery | mall | arr-right | search | user
    name: {
      type: String,
    },
    // string | string[]
    color: {
      type: null,
      observer: function(color) {
        this.setData({
          colors: this.fixColor(),
          isStr: typeof color === 'string',
        });
      }
    },
    size: {
      type: Number,
      value: 40,
      observer: function(size) {
        this.setData({
          svgSize: size / 750 * wx.getSystemInfoSync().windowWidth,
        });
      },
    },
  },
  data: {
    colors: '',
    svgSize: 40 / 750 * wx.getSystemInfoSync().windowWidth,
    quot: '"',
    isStr: true,
  },
  methods: {
    fixColor: function() {
      var color = this.data.color;
      var hex2rgb = this.hex2rgb;

      if (typeof color === 'string') {
        return color.indexOf('#') === 0 ? hex2rgb(color) : color;
      }

      return color.map(function (item) {
        return item.indexOf('#') === 0 ? hex2rgb(item) : item;
      });
    },
    hex2rgb: function(hex) {
      var rgb = [];

      hex = hex.substr(1);

      if (hex.length === 3) {
        hex = hex.replace(/(.)/g, '$1$1');
      }

      hex.replace(/../g, function(color) {
        rgb.push(parseInt(color, 0x10));
        return color;
      });

      return 'rgb(' + rgb.join(',') + ')';
    }
  }
});
