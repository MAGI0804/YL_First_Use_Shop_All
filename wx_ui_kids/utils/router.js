const TAB_ROUTES = [
  '/pages/index/index',
  '/pages/commodity/KIDS/index/index',
  '/pages/commodity/Adult/index/index',
  '/pages/cart/index/index',
  '/pages/my/index/index'
]

let lockedUrl = ''
let lockTimer = null

const normalizeUrl = (url = '') => {
  if (!url) return ''
  return url.startsWith('/') ? url : `/${url}`
}

const getRouteOnly = (url = '') => normalizeUrl(url).split('?')[0]

const isTabRoute = (url = '') => TAB_ROUTES.includes(getRouteOnly(url))

const lock = (url, duration = 500) => {
  if (lockedUrl === url) return false
  lockedUrl = url
  if (lockTimer) clearTimeout(lockTimer)
  lockTimer = setTimeout(() => {
    lockedUrl = ''
    lockTimer = null
  }, duration)
  return true
}

const unlock = (url) => {
  if (lockedUrl === url) {
    lockedUrl = ''
  }
}

const navigateTo = (options) => {
  const navOptions = typeof options === 'string' ? { url: options } : { ...(options || {}) }
  const url = normalizeUrl(navOptions.url)
  if (!url || !lock(url, navOptions.lockDuration)) {
    return
  }

  const complete = navOptions.complete
  const fail = navOptions.fail
  navOptions.url = url
  navOptions.complete = (res) => {
    unlock(url)
    if (typeof complete === 'function') complete(res)
  }

  if (isTabRoute(url)) {
    wx.switchTab({
      url: getRouteOnly(url),
      success: navOptions.success,
      fail: (err) => {
        if (typeof fail === 'function') fail(err)
        wx.reLaunch({ url: getRouteOnly(url) })
      },
      complete: navOptions.complete
    })
    return
  }

  wx.navigateTo({
    ...navOptions,
    fail: (err) => {
      const message = err && err.errMsg ? err.errMsg : ''
      if (message.includes('limit')) {
        wx.redirectTo({
          url,
          fail: (redirectErr) => {
            if (typeof fail === 'function') fail(redirectErr)
          }
        })
        return
      }
      if (typeof fail === 'function') fail(err)
    }
  })
}

const switchTab = (options) => {
  const navOptions = typeof options === 'string' ? { url: options } : { ...(options || {}) }
  navOptions.url = getRouteOnly(navOptions.url)
  navigateTo(navOptions)
}

const redirectTo = (options) => {
  const navOptions = typeof options === 'string' ? { url: options } : { ...(options || {}) }
  const url = normalizeUrl(navOptions.url)
  if (!url || !lock(url, navOptions.lockDuration)) {
    return
  }

  const complete = navOptions.complete
  const fail = navOptions.fail
  navOptions.url = url
  navOptions.complete = (res) => {
    unlock(url)
    if (typeof complete === 'function') complete(res)
  }

  if (isTabRoute(url)) {
    wx.switchTab({
      url: getRouteOnly(url),
      success: navOptions.success,
      fail: (err) => {
        if (typeof fail === 'function') fail(err)
        wx.reLaunch({ url: getRouteOnly(url) })
      },
      complete: navOptions.complete
    })
    return
  }

  wx.redirectTo({
    ...navOptions,
    fail: (err) => {
      unlock(url)
      navigateTo({
        ...navOptions,
        fail: fail || (() => {})
      })
    }
  })
}

const navigateBack = (options = {}) => {
  const pages = getCurrentPages()
  if (pages.length > 1) {
    wx.navigateBack(options)
    return
  }
  switchTab('/pages/index/index')
}

export default {
  navigateTo,
  switchTab,
  redirectTo,
  navigateBack,
  isTabRoute
}
