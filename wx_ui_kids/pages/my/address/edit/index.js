// pages/my/address/edit/index.js
const app = getApp();
// 页面顶部不再声明固定的userId，改为在需要时动态获取最新值

Page({

  /**
   * 页面的初始数据
   */
  data: {
    addressId: '',
    addressData: {
      consignee: '',
      phone: '',
      province: '',
      city: '',
      district: '',
      address: '',
      is_default: false
    },
    regionArray: [], // 用于地区选择器的数组
    phoneError: false, // 手机号错误状态
    phoneErrorMessage: '' // 手机号错误提示信息
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    console.log('页面加载，接收到的参数:', options);
    // 获取地址ID参数
    if (options && options.id) {
      console.log('检测到地址ID:', options.id);
      this.setData({
        addressId: options.id
      });
      // 获取地址详情
      this.getAddressDetail();
    } else {
      console.log('未检测到地址ID参数');
    }
  },

  /**
   * 获取地址详情
   */
  getAddressDetail() {
    console.log('开始获取地址详情，addressId:', this.data.addressId);
    // 动态获取最新的用户ID并转换为整数
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    const requestData = {
            user_id: userId,
            address_id: this.data.addressId
          }
    console.log('请求数据:', requestData);
    
    // 使用POST请求，并将参数放在请求体内，统一使用add_address路径
    app.req.post('/address/get_address_by_id', requestData, 
      (res) => {
        // console.log('获取地址详情响应:', res);
        if (res.code === 200 && res.data && res.data.address) {
          const address = res.data.address;
          // console.log('获取到的地址数据:', address);
          
          // 构建要设置的数据
          const dataToSet = {
            addressData: {
              consignee: address.receiver_name || '',
              phone: address.phone_number || '',
              province: address.province || '',
              city: address.city || '',
              district: address.county || '',
              address: address.detailed_address || '',
              is_default: !!address.is_default
            },
            // 设置地区选择器的初始值
            regionArray: [
              address.province || '',
              address.city || '',
              address.county || ''
            ].filter(Boolean) // 过滤空值
          };
          
          // console.log('准备设置的数据:', dataToSet);
          this.setData(dataToSet);
          // console.log('地址数据设置成功');
          
        } else {
          console.error('获取地址详情失败:', res.message || '未知错误');
          // 为了演示效果，提供模拟数据
        }
      },
      (err) => {
        console.error('网络错误:', err);
        // 网络错误时也提供模拟数据
      }
    );
  },

  /**
   * 收货人输入处理
   */
  onConsigneeInput(e) {
    this.setData({
      'addressData.consignee': e.detail.value
    });
  },

  /**
   * 手机号输入处理 - 添加实时验证
   */
  onPhoneInput(e) {
    const phone = e.detail.value;
    
    // 清除之前的错误状态
    this.setData({
      'addressData.phone': phone,
      phoneError: false,
      phoneErrorMessage: ''
    });
    
    // 如果输入框不为空，进行实时验证
    if (phone) {
      this.validatePhone(phone);
    }
  },
  
  /**
   * 验证手机号格式
   */
  validatePhone(phone) {
    // 1. 基本长度验证
    if (phone.length !== 11) {
      this.setData({
        phoneError: true,
        phoneErrorMessage: '请输入11位手机号码'
      });
      return false;
    }
    
    // 2. 手机号码格式验证
    const phoneRegex = /^1[3-9]\d{9}$/;
    if (!phoneRegex.test(phone)) {
      this.setData({
        phoneError: true,
        phoneErrorMessage: '请输入有效的手机号码'
      });
      return false;
    }
    
    // 3. 清除错误状态
    this.setData({
      phoneError: false,
      phoneErrorMessage: ''
    });
    return true;
  },

  /**
   * 详细地址输入处理
   */
  onAddressInput(e) {
    this.setData({
      'addressData.address': e.detail.value
    });
  },

  /**
   * 默认地址开关处理
   */
  onDefaultChange(e) {
    this.setData({
      'addressData.is_default': e.detail.value
    });
  },

  /**
   * 地区选择变化处理
   */
  onRegionChange(e) {
    const [province, city, district] = e.detail.value;
    this.setData({
      'addressData.province': province,
      'addressData.city': city,
      'addressData.district': district
    });
  },

  /**
   * 保存地址
   */
  saveAddress() {
    const { addressId, addressData } = this.data;
    const { consignee, phone, province, city, district, address } = addressData;
    
    // 表单验证
    if (!consignee) {
      console.warn('请输入收货人姓名');
      return;
    }
    
    // 调用统一的手机号验证方法
    if (!phone || !this.validatePhone(phone)) {
      return;
    }
    
    if (!province || !city || !district) {
      console.warn('请选择所在地区');
      return;
    }
    
    if (!address) {
      console.warn('请输入详细地址');
      return;
    }
    
    // 按照应用标准方式获取用户ID，并确保为整数类型
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    if (!userId) {
      console.warn('请先登录');
      return;
    }
    
    // 构建请求数据，严格按照用户提供的示例格式
    const requestData = {
      user_id: userId,
      address_id: addressId,
      receiver_name: consignee,
      phone_number: phone,
      province: province,
      city: city,
      county: district,
      detailed_address: address,
      is_default: addressData.is_default
    };
    
    wx.showLoading({
      title: '更新中...',
    });
    
    const url = '/address/update_address';
    const method = 'post';
    
    app.req[method](url, requestData, 
      (res) => {
        if (res.code === 200) {
          console.log('更新成功');
          // 延迟返回上一页
          setTimeout(() => {
            app.navigateBack();
          }, 1500);
        } else {
          console.error(res.message || '更新失败');
        }
      },
      (err) => {
        console.error('网络错误:', err);
      }
    );
  },

  /**
   * 返回上一页
   */
  navigateBack() {
    app.navigateBack();
  }
})
