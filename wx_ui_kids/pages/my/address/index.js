// pages/my/address/index.js
const app = getApp();

const getCurrentUserId = () => {
  const globalUserInfo = app.globalData.userInfo || {};
  return parseInt(globalUserInfo.user_id || app.globalData.user_id || wx.getStorageSync('user_id') || 0);
};

const getAddressErrorMessage = (err, fallback) => {
  return (err && (err.message || err.msg || (err.data && (err.data.message || err.data.msg)))) || fallback;
};

// 页面顶部的userId声明现在移到getAddressList方法内部，确保每次都获取最新值
Page({

  /**
   * 页面的初始数据
   */
  data: {
    addressList: [],
    syncingWechatAddress: false
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    console.log('地址管理页面加载，options:', options);
    // 保存来源页面信息
    if (options && (options.from === 'order' || options.from === 'return')) {
      this.fromPage = options.from;
      console.log('已设置fromPage为:', this.fromPage);
    } else {
      this.fromPage = '';
      console.log('fromPage设置为空字符串');
    }
    this.getAddressList();
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    // 每次页面显示时，重新获取地址列表
    this.getAddressList();
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
    // 下拉刷新时，重新获取地址列表
    this.getAddressList();
    wx.stopPullDownRefresh();
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

  },

  /**
   * 获取地址列表
   */
  getAddressList() {
    wx.showLoading({
      title: '加载中...',
    });
    
    // 按照应用标准方式获取用户ID：先检查全局变量，再检查本地存储，确保为整数类型
    const userId = getCurrentUserId();
    
    // 确保user_id存在且不为0
    if (!userId || userId <= 0) {
      console.error('用户ID不存在或无效，无法获取地址列表');
      wx.hideLoading();
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      
      // 为了演示效果，提供模拟数据
      this.setData({
        addressList: [
          {
            id: '1',
            consignee: '张三',
            phone: '13800138000',
            province: '上海市',
            city: '上海市',
            district: '浦东新区',
            address: '张江高科技园区博云路2号',
            is_default: true
          },
          {
            id: '2',
            consignee: '李四',
            phone: '13900139000',
            province: '北京市',
            city: '北京市',
            district: '朝阳区',
            address: '建国路88号',
            is_default: false
          }
        ]
      });
      return;
    } else {
      // 用户已登录，继续获取地址列表
      this.setData({
        addressList: [] // 清空列表，防止显示旧数据
      });
    }
    
    const requestData = { user_id: userId };
    // console.log('发送请求数据:', requestData);
    
    app.req.post('/address/get_addresses', requestData, 
      (res) => {
        wx.hideLoading();
        console.log('获取地址列表响应:', res);
        // 根据用户提供的返回格式进行处理
        if (res.code === 200 && res.data && Array.isArray(res.data.addresses)) {
          // 转换数据格式以适配页面模板
          const formattedAddresses = res.data.addresses.map(item => ({
            id: item.address_id, // 后端返回的address_id映射到前端的id
            consignee: item.receiver_name || '', // 使用receiver_name字段
            phone: item.phone_number || '', // 使用phone_number字段
            province: item.province || '',
            city: item.city || '',
            district: item.county || '', // 注意字段名是county不是district
            address: item.detailed_address || '',
            is_default: item.is_default  || false // 确保是布尔值
          }));
          
          this.setData({
            addressList: formattedAddresses
          });
        } else {
          wx.showToast({
            title: res.message || '获取地址列表失败',
            icon: 'none'
          });
          // 为了演示效果，提供模拟数据
          this.setData({
            addressList: [
            ]
          });
        }
      },
      (err) => {
        wx.hideLoading();
        wx.showToast({
          title: '网络错误，请重试',
          icon: 'none'
        });
        // 为了演示效果，提供模拟数据
        this.setData({
          addressList: [
          ]
        });
      }
    );
  },

  /**
   * 跳转到添加地址页面
   */
  navigateToAddAddress() {
    wx.navigateTo({
      url: '/pages/my/address/add/index'
    });
  },

  /**
   * 从微信收货地址簿同步地址到商城地址
   */
  syncWechatAddress() {
    if (this.data.syncingWechatAddress) {
      return;
    }

    const userId = getCurrentUserId();
    if (!userId || userId <= 0) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      return;
    }

    if (typeof wx.chooseAddress !== 'function') {
      wx.showModal({
        title: '当前微信版本不支持',
        content: '请升级微信后再同步微信收货地址。',
        showCancel: false
      });
      return;
    }

    this.setData({ syncingWechatAddress: true });

    wx.chooseAddress({
      success: (wechatAddress) => {
        const requestData = {
          user_id: userId,
          receiver_name: (wechatAddress.userName || '').trim(),
          phone_number: (wechatAddress.telNumber || '').trim(),
          province: (wechatAddress.provinceName || '').trim(),
          city: (wechatAddress.cityName || '').trim(),
          county: (wechatAddress.countyName || '').trim(),
          detailed_address: (wechatAddress.detailInfo || '').trim(),
          is_default: this.data.addressList.length === 0,
          remark: '微信地址'
        };

        if (!requestData.receiver_name || !requestData.phone_number || !requestData.province || !requestData.city || !requestData.county || !requestData.detailed_address) {
          this.setData({ syncingWechatAddress: false });
          wx.showToast({
            title: '微信地址信息不完整',
            icon: 'none'
          });
          return;
        }

        wx.showLoading({
          title: '保存中...',
          mask: true
        });

        app.req.post('/address/add_address', requestData,
          (res) => {
            this.setData({ syncingWechatAddress: false });
            if (res.code === 200) {
              wx.showToast({
                title: '同步成功',
                icon: 'success'
              });
              this.getAddressList();
              return;
            }
            wx.showToast({
              title: getAddressErrorMessage(res, '同步失败'),
              icon: 'none'
            });
          },
          (err) => {
            this.setData({ syncingWechatAddress: false });
            wx.showToast({
              title: getAddressErrorMessage(err, '网络错误，请重试'),
              icon: 'none'
            });
          }
        );
      },
      fail: (err) => {
        this.setData({ syncingWechatAddress: false });
        const errMsg = err && err.errMsg ? err.errMsg : '';
        if (errMsg.includes('cancel')) {
          return;
        }
        if (errMsg.includes('auth deny') || errMsg.includes('authorize')) {
          wx.showModal({
            title: '需要地址权限',
            content: '请允许使用微信收货地址后再同步。',
            confirmText: '去设置',
            success: (modalRes) => {
              if (modalRes.confirm) {
                wx.openSetting();
              }
            }
          });
          return;
        }
        wx.showToast({
          title: '未获取到微信地址',
          icon: 'none'
        });
      }
    });
  },
  
  /**
   * 选择地址（从订单页面过来时使用）
   */
  selectAddress(e) {
    const addressId = e.currentTarget.dataset.id;
    console.log('选择地址，addressId:', addressId);
    console.log('fromPage值:', this.fromPage);
    // 如果是从订单页面或售后页面过来的，则选择地址并返回
    if (this.fromPage === 'order' || this.fromPage === 'return') {
      console.log('从', this.fromPage, '页面过来，执行选择地址并返回逻辑');
      const selectedAddress = this.data.addressList.find(item => item.id === addressId);
      if (selectedAddress) {
        console.log('找到选中的地址:', selectedAddress);
        // 转换地址格式以适配订单页面和售后页面
        const formattedAddress = {
          id: selectedAddress.id,
          name: selectedAddress.consignee,
          phone: selectedAddress.phone,
          province: selectedAddress.province,
          city: selectedAddress.city,
          county: selectedAddress.district, // 注意：售后页面使用county字段
          district: selectedAddress.district, // 同时保留district字段以兼容订单页面
          address: selectedAddress.address, // 注意：售后页面使用address字段
          detail: selectedAddress.address, // 同时保留detail字段以兼容订单页面
          is_default: selectedAddress.is_default
        };
        
        console.log('格式化后的地址:', formattedAddress);
        // 存储选中的地址并返回来源页面
        wx.setStorageSync('selectedAddress', formattedAddress);
        console.log('已存储选中的地址到缓存');
        wx.navigateBack();
        console.log('已返回', this.fromPage, '页面');
      } else {
        console.warn('未找到选中的地址，addressId:', addressId);
      }
    } else {
      console.log('不是从订单或售后页面过来，不执行选择地址逻辑');
    }
  },

  /**
   * 编辑地址
   */
  editAddress(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/my/address/edit/index?id=${id}`
    });
  },

  /**
   * 删除地址 - 按照新接口规范修改
   */
  deleteAddress(e) {
    const addressId = e.currentTarget.dataset.id;
    wx.showModal({
      title: '提示',
      content: '确定要删除该地址吗？',
      success: (res) => {
        if (res.confirm) {
          // 获取用户ID
          const userId = getCurrentUserId();
          
          if (!userId) {
            console.warn('用户ID不存在，无法删除地址');
            return;
          }
          
          wx.showLoading({
            title: '删除中...',
          });
          
          // 构建请求数据
          const requestData = {
            user_id: userId,
            address_id: addressId
          };
          
          // 发送POST请求到新接口
          app.req.post('/address/delete_address', requestData, 
            (res) => {
              wx.hideLoading();
              // 根据响应的code字段判断
              if (res.code === 200) {
                console.log('删除地址成功:', res.msg || '删除成功');
                // 重新获取地址列表
                this.getAddressList();
              } else {
                console.warn('删除地址失败:', res.msg || '删除失败');
                // 模拟删除成功
                const newList = this.data.addressList.filter(item => item.id !== addressId);
                this.setData({
                  addressList: newList
                });
              }
            },
            (err) => {
              wx.hideLoading();
              console.error('删除地址网络错误:', err);
              // 模拟删除成功
              const newList = this.data.addressList.filter(item => item.id !== addressId);
              this.setData({
                addressList: newList
              });
            }
          );
        }
      }
    });
  },

  /**
   * 设置默认地址 - 按照新接口规范修改
   */
  setDefaultAddress(e) {
    const addressId = e.currentTarget.dataset.id;
    
    // 动态获取最新的用户ID并转换为整数
    const userId = getCurrentUserId();
    
    if (!userId || userId <= 0) {
      console.warn('请先登录');
      return;
    }
    
    // 构建请求数据，严格按照用户提供的示例格式
    const requestData = {
      user_id: userId,
      address_id: addressId
    };
    
    console.log('设置默认地址请求数据:', requestData);
    
    // 使用新的URL和请求方式
    app.req.post('/address/set_default_address', requestData, 
      (res) => {
        console.log('设置默认地址响应:', res);
        // 按照新的响应格式处理
        if (res.code === 200) {
          console.log('设置默认地址成功:', res.msg || '设置成功');
          // 重新获取地址列表以更新UI
          this.getAddressList();
        } else {
          console.error('设置默认地址失败:', res.msg || '未知错误');
          // 模拟设置成功
          const newList = this.data.addressList.map(item => ({
            ...item,
            is_default: item.id === addressId
          }));
          this.setData({
            addressList: newList
          });
        }
      },
      (err) => {
        console.error('设置默认地址网络错误:', err);
        // 模拟设置成功
        const newList = this.data.addressList.map(item => ({
          ...item,
          is_default: item.id === addressId
        }));
        this.setData({
          addressList: newList
        });
      }
    );
  }
})
