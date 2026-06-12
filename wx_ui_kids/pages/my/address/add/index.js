// pages/my/address/add/index.js
const app = getApp();
Page({

  /**
   * 页面的初始数据
   */
  data: {
    isEdit: false,
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
    // 检查是否为编辑模式
    if (options && options.id) {
      this.setData({
        isEdit: true,
        addressId: options.id
      });
      // 如果是编辑模式，获取地址详情
      this.getAddressDetail();
    }
  },

  /**
   * 获取地址详情
   */
  getAddressDetail() {
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);

    if (!userId) {
      console.warn('请先登录');
      return;
    }

    app.req.post('/address/get_address_by_id', {
      user_id: userId,
      address_id: this.data.addressId
    },
      (res) => {
        if (res.code === 200 && res.data && res.data.address) {
          const address = res.data.address;
          this.setData({
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
      });
        } else {
          console.error('获取地址详情失败:', res.message || '未知错误');
        }
      },
      (err) => {
        console.error('网络错误:', err);
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
    const { isEdit, addressId, addressData } = this.data;
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
    
    // 按照应用标准方式获取用户ID
    const globalUserInfo = app.globalData.userInfo || {};
    const userId = parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
    
    if (!userId) {
      console.warn('请先登录');
      return;
    }
    
    // 构建请求数据，严格按照用户提供的示例格式
    const requestData = {
      user_id: userId,
      receiver_name: consignee,
      phone_number: phone,
      province: province,
      city: city,
      county: district,
      detailed_address: address
    };
    
    // 如果选择设为默认地址，增加is_default参数
    if (addressData.is_default) {
      requestData.is_default = true;
    }
    
    // 如果是编辑模式，添加address_id参数
    if (isEdit && addressId) {
      requestData.address_id = addressId;
    }
    
    wx.showLoading({
      title: isEdit ? '更新中...' : '保存中...',
    });
    
    const url = isEdit ? '/address/update_address' : '/address/add_address';
    const method = 'post';
    
    app.req[method](url, requestData, 
      (res) => {
        if (res.code === 200) {
          console.log(isEdit ? '更新成功' : '添加成功');
          // 延迟返回上一页
          setTimeout(() => {
            app.navigateBack();
          }, 1500);
        } else {
          console.error(res.message || (isEdit ? '更新失败' : '添加失败'));
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
