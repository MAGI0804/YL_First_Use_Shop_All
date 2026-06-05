const API_HOSTS = {
  develop: 'https://snow-api.youlankids.com',
  trial: 'https://snow-api.youlankids.com',
  release: 'https://snow-api.youlankids.com'
}

const getEnvVersion = () => {
  try {
    return wx.getAccountInfoSync().miniProgram.envVersion || 'release'
  } catch (error) {
    return 'release'
  }
}

const getHost = () => {
  const overrideHost = wx.getStorageSync('api_host_override')
  if (overrideHost) {
    return overrideHost
  }
  return API_HOSTS[getEnvVersion()] || API_HOSTS.release
}

let tokenTask = null

const safeCallback = (callback, payload) => {
  if (typeof callback === 'function') {
    callback(payload)
  }
}

const goHome = () => {
  wx.switchTab({
    url: '/pages/index/index',
    fail() {
      wx.reLaunch({
        url: '/pages/index/index'
      })
    }
  })
}

/**
 * 获取access_token并存储到本地
 * 
 * @param {Function} success 成功回调
 * @param {Function} fail 失败回调
 */
const getAccessToken = (success, fail) => {
  // 先检查本地是否已有access_token，避免重复请求
  const existingToken = wx.getStorageSync('access_token');
  if (existingToken) {
    safeCallback(success, existingToken);
    return;
  }

  if (tokenTask) {
    tokenTask
      .then(token => safeCallback(success, token))
      .catch(err => safeCallback(fail, err))
    return
  }

  // 发送请求获取access_token
  tokenTask = new Promise((resolve, reject) => {
    wx.request({
      url: `${getHost()}/access_token/get_token`,
      method: 'POST',
      timeout: 10000,
      success: (res) => {
        if (res.data && res.data.code === 200 && res.data.data && res.data.data.access_token) {
          const accessToken = res.data.data.access_token;
          wx.setStorageSync('access_token', accessToken);
          resolve(accessToken);
        } else {
          const message = (res.data && (res.data.message || res.data.msg)) || '获取access_token失败';
          console.error('获取access_token失败:', message);
          reject(message);
        }
      },
      fail: (err) => {
        console.error('获取access_token请求失败:', err);
        reject('网络错误，无法获取access_token');
      },
      complete: () => {
        tokenTask = null
      }
    });
  })

  tokenTask
    .then(token => safeCallback(success, token))
    .catch(err => safeCallback(fail, err))
};

const http = (method, url, data, response, error, options = {}) => {
  // 获取存储的access_token
  const accessToken = wx.getStorageSync('access_token')
  
  // 如果没有access_token，则先获取
  if (!accessToken) {
    getAccessToken(
      (token) => {
        // 成功获取到token后，重新调用http函数
        http(method, url, data, response, error, options);
      },
      (errMsg) => {
        console.error('获取access_token失败:', errMsg);
        safeCallback(error, { message: errMsg });
      }
    );
    return;
  }
  
  // 处理URL，为所有请求在URL参数中添加access_token
  const separator = url.includes('?') ? '&' : '?';
  url = `${url}${separator}access_token=${encodeURIComponent(accessToken)}`;
  
  // wx.showLoading({
  //   title: '加载中...',
  //   mask: true
  // })
  
  // console.log('发起HTTP请求:', method, getHost() + url, data);
  
  // 使用传入的host或默认host
  const requestHost = options.host || getHost();
  
  wx.request({
    method: method,
    url: requestHost + url,
    timeout: options.timeout || 12000,
    header: {
      'content-type': 'application/json'
     },
    data: data,  // 直接使用原始数据，不添加access_token
    success: res => {
      // console.log('HTTP请求成功响应:', res);
      if (res.statusCode < 200 || res.statusCode >= 300) {
        safeCallback(error, {
          message: `请求失败(${res.statusCode})`,
          statusCode: res.statusCode,
          data: res.data
        })
        return
      }
      safeCallback(response, res.data || {})
    },
    fail: err => {
      console.error('HTTP请求失败:', err);
      safeCallback(error, err)
    },
    complete: info => {
      if(info.data && (info.data.code===401 || info.data.code===402)){
        if(url.includes('bonus/detail')){
          wx.hideLoading();
        }else{
          // 清除失效的token
          wx.removeStorageSync('access_token');
          goHome()
        }
      } else if(info.data && info.data.code===201){
        // 走注册流程
      } else if(info.data && info.data.code===200){
        // 成功
        wx.hideLoading();
      } else if(info.data && info.data.code===403){
        // 没有相关权限
        console.warn('没有相关权限:', info.data.message);
        wx.login({
          success(res){
            getApp().globalData.code = res.code
          }
        })
      }  else {
        if(info.data && info.data.message){
          // 不显示弹窗，仅在控制台输出信息
          console.log('提示信息:', info.data.message);
        }
      }
      if(url.includes('order/submit') && info.data && info.data.code===200){
        // do
      }else{
        wx.hideLoading();
      }
    }
  })
}

// 导出http和getAccessToken方法
export default {
  http,
  getAccessToken,
  getHost,
  API_HOSTS,
  // 添加req方法作为http方法的别名，保持向后兼容
  req: http,
  // 导出host变量供其他文件使用
  get host() {
    return getHost()
  }
}
