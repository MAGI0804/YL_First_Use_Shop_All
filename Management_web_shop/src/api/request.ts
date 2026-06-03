import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:3088'

interface AccessTokenResponse {
  code: number
  msg?: string
  data?: {
    access_token?: string
  }
}

class HttpRequest {
  private instance: AxiosInstance
  private token: string = ''
  private backendToken: string = ''

  constructor() {
    this.instance = axios.create({
      baseURL: BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json'
      }
    })
    this.token = localStorage.getItem('access_token') || ''
    this.backendToken = localStorage.getItem('backend_token') || ''

    this.instance.interceptors.request.use(
      (config) => {
        if (this.token && !config.url?.includes('/access_token/get_token')) {
          config.params = {
            ...config.params,
            access_token: this.token
          }
        }
        this.backendToken = localStorage.getItem('backend_token') || ''
        if (this.backendToken && config.headers) {
          ;(config.headers as any).Authorization = `Bearer ${this.backendToken}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    this.instance.interceptors.response.use(
      (response: AxiosResponse) => {
        return response.data
      },
      async (error) => {
        const originalRequest = error.config
        const responseData = error.response?.data
        
        if (responseData && responseData.code === 201 && responseData.msg === 'token验证失败' && !originalRequest._retry) {
          originalRequest._retry = true
          try {
            await this.getToken()
            originalRequest.params = {
              ...originalRequest.params,
              access_token: this.token
            }
            return this.instance(originalRequest)
          } catch (tokenError) {
            console.error('重新获取token失败:', tokenError)
            return Promise.reject(tokenError)
          }
        }
        if (error.response?.status === 401 && !originalRequest?.url?.includes('/backend_login')) {
          localStorage.removeItem('backend_token')
          localStorage.removeItem('backend_refresh_token')
          localStorage.removeItem('backend_user')
        }
        
        return Promise.reject(error)
      }
    )
  }

  async getToken(): Promise<string> {
    try {
      const response = await this.instance.post<any, AccessTokenResponse>('/access_token/get_token')
      if (response.code === 200 && response.data && response.data.access_token) {
        this.token = response.data.access_token
        localStorage.setItem('access_token', this.token)
        return this.token
      } else {
        throw new Error(response.msg || '获取token失败')
      }
    } catch (error) {
      console.error('获取token失败:', error)
      throw error
    }
  }

  setBackendToken(token: string) {
    this.backendToken = token
    localStorage.setItem('backend_token', token)
  }

  clearBackendToken() {
    this.backendToken = ''
    localStorage.removeItem('backend_token')
    localStorage.removeItem('backend_refresh_token')
    localStorage.removeItem('backend_user')
  }

  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.get(url, config)
  }

  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.post(url, data, config)
  }

  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.put(url, data, config)
  }

  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.delete(url, config)
  }
}

const http = new HttpRequest()

export default http
export { BASE_URL }
