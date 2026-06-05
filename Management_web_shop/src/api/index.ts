import http from './request'

export const getToken = () => {
  return http.getToken()
}

export interface BackendUserSession {
  id: number
  operator_no: string
  mobile: string
  nickname: string
  role: 'operation' | 'customer_service' | 'admin'
  status: 'pending' | 'active' | 'disabled'
  permissions: string[]
  remarks?: string
  token?: string
  refresh_token?: string
}

export interface BackendAuthResponse {
  code: number
  data: {
    user: BackendUserSession
  }
  msg: string
}

export const backendLogin = (params: { mobile: string; password: string }) => {
  return http.post<BackendAuthResponse>('/OperationUser/backend_login', params)
}

export const sendBackendRegisterCaptcha = (params: { mobile: string }) => {
  return http.post('/OperationUser/send_register_captcha', params)
}

export const backendRegisterByPhone = (params: { mobile: string; password: string; captcha: string }) => {
  return http.post<BackendAuthResponse>('/OperationUser/backend_register_by_phone', params)
}

export const queryBackendMe = async () => {
  await http.getToken()
  return http.post<BackendAuthResponse>('/OperationUser/backend_me')
}

export interface BackendUserQueryParams {
  mobile?: string
  status?: string
  role?: string
  page: number
  page_size: number
}

export interface BackendUserQueryResponse {
  code: number
  data: {
    items: BackendUserSession[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export const queryBackendUsers = (params: BackendUserQueryParams) => {
  return http.post<BackendUserQueryResponse>('/OperationUser/backend_users', params)
}

export const inviteBackendUser = (params: { mobile: string; nickname: string; role?: string; permissions?: string[]; remarks?: string }) => {
  return http.post<BackendAuthResponse>('/OperationUser/backend_invite_user', params)
}

export const updateBackendUserStatus = (params: { id: number; status: 'pending' | 'active' | 'disabled' }) => {
  return http.post<BackendAuthResponse>('/OperationUser/backend_update_status', params)
}

export const updateBackendUser = (params: { id: number; nickname?: string; role?: string; status?: string; permissions?: string[]; remarks?: string }) => {
  return http.post<BackendAuthResponse>('/OperationUser/backend_update_user', params)
}

export const saveBackendSession = (session: BackendUserSession) => {
  if (session.token) {
    http.setBackendToken(session.token)
  }
  if (session.refresh_token) {
    localStorage.setItem('backend_refresh_token', session.refresh_token)
  }
  localStorage.setItem('backend_user', JSON.stringify({
    ...session,
    token: undefined,
    refresh_token: undefined
  }))
}

export const getStoredBackendUser = (): BackendUserSession | null => {
  const value = localStorage.getItem('backend_user')
  if (!value) return null
  try {
    return JSON.parse(value) as BackendUserSession
  } catch {
    return null
  }
}

export const clearBackendSession = () => {
  http.clearBackendToken()
}

export interface OrderQueryParams {
  page: number
  page_size: number
  shopname: string
  status?: string
  begin_time?: string
  end_time?: string
  tid?: string
  mobile?: string
  member_id?: number
  member_no?: string
  sub_order_id?: string
  pay_status?: string
}

export interface OrderItem {
  canceled_time: string
  city: string
  county: string
  delivered_time: string
  detailed_address: string
  express_company: string
  express_number: string
  logistics_process: any[]
  order_amount: number
  final_pay_amount?: number
  discount_amount?: number
  discount_reason?: string
  order_id: string
  order_time: string
  pay_status?: string
  process_num: string
  processing_time: string
  product_list: string[]
  province: string
  receiver_name: string
  receiver_phone: string
  remarks: string
  shipped_time: string
  status: string
  user_id: number
}

export interface OrderQueryResponse {
  code: number
  data: {
    code: number
    data: OrderItem[]
    page: number
    page_size: number
    total: number
  }
  msg: string
}

export interface ProductItem {
  category: string
  category_detail: string
  color: string
  commodity_id: string
  created_at: string
  height: string
  image: string
  name: string
  notes: string
  price: number
  size: string
  spec_code: string
  style_code: string
}

export interface BatchProductParams {
  commodity_ids: string[]
}

export interface BatchProductResponse {
  code: number
  data: {
    count: number
    data: ProductItem[]
  }
  msg: string
}

export const queryOrders = (params: OrderQueryParams) => {
  return http.post<OrderQueryResponse>('/order/orders_query', params)
}

export interface MemberItem {
  id: number
  member_no: string
  user_id: number
  openid: string
  mobile: string
  manual_unique_code: string
  nickname: string
  status: string
  source: string
  total_order_amount: number
  total_paid_amount: number
  tmall_id: string
  tmall_amount: number
  youzan_id: string
  youzan_amount: number
  remarks: string
  created_at: string
  updated_at: string
}

export interface MemberTagItem {
  id: number
  name: string
  color: string
  remarks: string
  created_by: number
  created_at: string
  updated_at: string
}

export interface MemberQueryParams {
  page: number
  page_size: number
  mobile?: string
  member_no?: string
  manual_unique_code?: string
  nickname?: string
  status?: string
  tag_id?: number
  tag_name?: string
}

export interface MemberQueryResponse {
  code: number
  data: {
    items: MemberItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export interface MemberDetailResponse {
  code: number
  data: {
    detail: {
      member: MemberItem
      tags: MemberTagItem[]
    }
  }
  msg: string
}

export interface MemberTagQueryResponse {
  code: number
  data: {
    items: MemberTagItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export interface OperationLogItem {
  id: number
  operator_id: number
  operator_no: string
  operator_mobile: string
  operator_role: string
  action: string
  module: string
  target_type: string
  target_id: string
  member_id: number
  user_id: number
  order_id: string
  before_data: string
  after_data: string
  client_ip: string
  remark: string
  created_at: string
}

export interface OperationLogQueryParams {
  page: number
  page_size: number
  operator_id?: number
  action?: string
  module?: string
  target_type?: string
  target_id?: string
  member_id?: number
  user_id?: number
  order_id?: string
  begin_time?: string
  end_time?: string
}

export interface OperationLogQueryResponse {
  code: number
  data: {
    items: OperationLogItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export const queryMembers = (params: MemberQueryParams) => {
  return http.post<MemberQueryResponse>('/member/list', params)
}

export const createMember = (params: Partial<MemberItem>) => {
  return http.post('/member/create', params)
}

export const updateMember = (params: Partial<MemberItem> & { id: number }) => {
  return http.post('/member/update', params)
}

export const queryMemberDetail = (params: { id?: number; member_no?: string; mobile?: string; user_id?: number }) => {
  return http.post<MemberDetailResponse>('/member/detail', params)
}

export const queryMemberTags = (params: { name?: string; page: number; page_size: number }) => {
  return http.post<MemberTagQueryResponse>('/member/tag/list', params)
}

export const createMemberTag = (params: { name: string; color?: string; remarks?: string }) => {
  return http.post('/member/tag/create', params)
}

export const setMemberTags = (params: { member_id: number; tag_ids: number[] }) => {
  return http.post('/member/tag/set_member_tags', params)
}

export const queryMemberCart = (params: { member_id?: number; member_no?: string; mobile?: string; user_id?: number }) => {
  return http.post('/member/cart/query', params)
}

export const addMemberCartItem = (params: { member_id?: number; user_id?: number; commodity_code: string; quantity: number }) => {
  return http.post('/member/cart/add', params)
}

export const updateMemberCartQuantity = (params: { member_id?: number; user_id?: number; commodity_code: string; quantity: number }) => {
  return http.post('/member/cart/update_quantity', params)
}

export const deleteMemberCartItems = (params: { member_id?: number; user_id?: number; commodity_codes?: string[] }) => {
  return http.post('/member/cart/delete', params)
}

export const backendCreateOrder = (params: any) => {
  return http.post('/order/backend_create_order', params)
}

export const queryOperationLogs = (params: OperationLogQueryParams) => {
  return http.post<OperationLogQueryResponse>('/operation_log/query', params)
}

export const batchGetProducts = (params: BatchProductParams) => {
  return http.post<BatchProductResponse>('/commodity/batch_get_products_by_ids', params)
}

export interface AnalyticsFilterParams {
  begin_time?: string
  end_time?: string
  shopname?: string
  category?: string
  style_code?: string
  low_inventory_threshold?: number
  slow_sales_threshold?: number
  limit?: number
}

export interface SalesDailyPoint {
  date: string
  order_count: number
  paid_order_count: number
  canceled_order_count: number
  sales_amount: number
  paid_amount: number
  original_order_amount: number
  discount_amount: number
  refund_amount: number
  average_order_value: number
}

export interface SalesSummaryData {
  order_count: number
  paid_order_count: number
  canceled_order_count: number
  sales_amount: number
  paid_amount: number
  original_order_amount: number
  discount_amount: number
  refund_amount: number
  average_order_value: number
  daily: SalesDailyPoint[]
}

export interface UserPreferencePoint {
  name: string
  user_count: number
  sales_qty: number
}

export interface UserSummaryData {
  new_user_count: number
  new_member_count: number
  order_user_count: number
  paid_user_count: number
  repurchase_user_count: number
  category_preferences: UserPreferencePoint[]
  style_preferences: UserPreferencePoint[]
}

export interface ProductSalesPoint {
  commodity_id: string
  name: string
  style_code: string
  category: string
  sales_qty: number
  sales_amount: number
  inventory: number
}

export interface StyleSalesPoint {
  style_code: string
  sales_qty: number
  sales_amount: number
  inventory: number
}

export interface ProductSummaryData {
  hot_skus: ProductSalesPoint[]
  hot_style_codes: StyleSalesPoint[]
  slow_moving_products: ProductSalesPoint[]
  inventory_turnover_rate: number
  low_inventory_count: number
  average_rating: number
  good_rate: number
}

export interface AnalyticsSummaryResponse<T> {
  code: number
  data: {
    summary: T
  }
  msg: string
}

export interface AnalyticsExportResponse {
  code: number
  data: {
    export: {
      sales: SalesSummaryData
      users: UserSummaryData
      products: ProductSummaryData
      traffic?: any
    }
  }
  msg: string
}

export const querySalesSummary = (params: AnalyticsFilterParams) => {
  return http.post<AnalyticsSummaryResponse<SalesSummaryData>>('/analytics/sales_summary', params)
}

export const queryUserSummary = (params: AnalyticsFilterParams) => {
  return http.post<AnalyticsSummaryResponse<UserSummaryData>>('/analytics/user_summary', params)
}

export const queryProductSummary = (params: AnalyticsFilterParams) => {
  return http.post<AnalyticsSummaryResponse<ProductSummaryData>>('/analytics/product_summary', params)
}

export const exportAnalytics = (params: AnalyticsFilterParams) => {
  return http.post<AnalyticsExportResponse>('/analytics/export', params)
}

export type DownloadTaskStatus = 'pending' | 'running' | 'success' | 'failed' | 'expired'
export type DownloadBusinessType = 'order' | 'product' | 'report' | 'inventory' | 'after_sale'

export interface DownloadTemplateItem {
  id: number
  template_code: string
  template_name: string
  business_type: DownloadBusinessType
  sql_content: string
  model_fields: string
  export_headers: string
  allowed_filters: string
  default_order_by: string
  file_format: string
  status: string
  created_at: string
  updated_at: string
}

export interface DownloadTaskItem {
  task_id: string
  template_code: string
  business_type: DownloadBusinessType
  task_name: string
  filters: string
  status: DownloadTaskStatus
  progress: number
  row_count: number
  file_path: string
  file_name: string
  file_size: number
  error_message: string
  download_count: number
  requested_by: number
  started_at?: string
  finished_at?: string
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface CreateDownloadTaskParams {
  template_code: string
  filters?: Record<string, any>
  file_format?: 'xlsx'
}

export interface DownloadTaskQueryParams {
  page: number
  page_size: number
  status?: string
  business_type?: string
  template_code?: string
}

export interface CreateDownloadTaskResponse {
  code: number
  data: {
    task: DownloadTaskItem
  }
  msg: string
}

export interface DownloadTasksResponse {
  code: number
  data: {
    list: DownloadTaskItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export interface DownloadTaskDetailResponse {
  code: number
  data: {
    task: DownloadTaskItem
  }
  msg: string
}

export interface DownloadTemplatesResponse {
  code: number
  data: {
    list: DownloadTemplateItem[]
  }
  msg: string
}

export const createDownloadTask = (params: CreateDownloadTaskParams) => {
  return http.post<CreateDownloadTaskResponse>('/download_center/tasks', params)
}

export const queryDownloadTasks = (params: DownloadTaskQueryParams) => {
  return http.get<DownloadTasksResponse>('/download_center/tasks', { params })
}

export const queryDownloadTaskDetail = (taskId: string) => {
  return http.get<DownloadTaskDetailResponse>(`/download_center/tasks/${taskId}`)
}

export const retryDownloadTask = (taskId: string) => {
  return http.post<DownloadTaskDetailResponse>(`/download_center/tasks/${taskId}/retry`)
}

export const queryDownloadTemplates = () => {
  return http.get<DownloadTemplatesResponse>('/download_center/templates')
}

export const downloadTaskFile = (taskId: string) => {
  return http.get<Blob>(`/download_center/tasks/${taskId}/file`, { responseType: 'blob' })
}

export interface ReviewReplyItem {
  id: number
  review_id: number
  operator_id: string
  content: string
  created_at: string
}

export interface ReviewItem {
  id: number
  user_id: number
  order_id: string
  sub_order_id: string
  commodity_id: string
  style_code: string
  rating: number
  content: string
  images: string
  tags: string
  status: string
  audit_remark: string
  created_at: string
  updated_at: string
  replies?: ReviewReplyItem[]
}

export interface ReviewBackendQueryParams {
  user_id?: number
  order_id?: string
  sub_order_id?: string
  commodity_id?: string
  style_code?: string
  status?: string
  page: number
  page_size: number
}

export interface ReviewQueryResponse {
  code: number
  data: {
    data: ReviewItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export interface ReviewAuditParams {
  review_id: number
  status: 'approved' | 'rejected' | 'hidden'
  audit_remark?: string
}

export interface ReviewReplyParams {
  review_id: number
  operator_id: string
  content: string
}

export interface ReviewStatisticsData {
  total: number
  average_rating: number
  good_rate: number
  rating_distribution: Record<string, number>
}

export interface ReviewStatisticsResponse {
  code: number
  data: {
    statistics: ReviewStatisticsData
  }
  msg: string
}

export const queryBackendReviews = (params: ReviewBackendQueryParams) => {
  return http.post<ReviewQueryResponse>('/review/query_backend', params)
}

export const auditReview = (params: ReviewAuditParams) => {
  return http.post('/review/audit', params)
}

export const replyReview = (params: ReviewReplyParams) => {
  return http.post('/review/reply', params)
}

export const queryReviewStatistics = (params: { commodity_id?: string; style_code?: string }) => {
  return http.post<ReviewStatisticsResponse>('/review/statistics', params)
}

export interface OrderDetailQueryParams {
  order_id: string
  inquired_list: string[]
  shopname?: string
}

export interface OrderDetailData {
  canceled_time: string
  city: string
  county: string
  delivered_time: string
  delivery_method: string
  detailed_address: string
  express_company: string
  express_number: string
  is_send_all: string
  jushuitan_order_id: string
  lcid: string
  logistics_process: any[]
  order_amount: number
  final_pay_amount?: number
  discount_amount?: number
  discount_reason?: string
  order_from: string
  order_id: string
  order_time: string
  payment_method: string
  payment_time: string
  pay_status?: string
  process_num: string
  processing_time: string
  product_list: string[]
  province: string
  receiver_name: string
  receiver_phone: string
  remarks: string
  shipped_time: string
  status: string
  sub_order_ids: string[]
  user_id: number
  wms_co_id: string
}

export interface OrderDetailResponse {
  code: number
  data: {
    code: number
    data: OrderDetailData
    message: string
    status: string
  }
  msg: string
}

export const queryOrderDetail = (params: OrderDetailQueryParams) => {
  return http.post<OrderDetailResponse>('/order/query_order_data', params)
}

export interface UpdatePaymentAmountParams {
  order_id: string
  final_pay_amount: number
  discount_reason?: string
  operator_id?: number
}

export const updatePaymentAmount = (params: UpdatePaymentAmountParams) => {
  return http.post('/order/update_payment_amount', params)
}

export interface ConfirmPaymentParams {
  order_id: string
  operator_id?: number
  payment_remark?: string
}

export const confirmOrderPayment = (params: ConfirmPaymentParams) => {
  return http.post('/order/confirm_payment', params)
}

export interface DeliverOrderParams {
  order_id: string
  user_id: number
  express_company: string
  express_number: string
}

export const deliverOrder = (params: DeliverOrderParams) => {
  return http.post('/order/backend_deliver', params)
}

export interface ReceiveOrderParams {
  order_id: string
  user_id: number
}

export const receiveOrder = (params: ReceiveOrderParams) => {
  return http.post('/order/backend_receive', params)
}

export interface GetAllLabelsParams {
  shopname: string
}

export interface GetAllLabelsResponse {
  code: number
  data: {
    label_four: string[]
    label_one: string[]
    label_seven: string[]
    label_three: string[]
    label_two: string[]
  }
  msg: string
}

export const getAllLabels = (params: GetAllLabelsParams) => {
  return http.post<GetAllLabelsResponse>('/commodity/get_all_labels', params)
}

export interface GetAllCategoriesParams {
  shopname: string
}

export interface GetAllCategoriesResponse {
  code: number
  data: {
    categories: string[]
  }
  msg: string
}

export const getAllCategories = (params: GetAllCategoriesParams) => {
  return http.post<GetAllCategoriesResponse>('/commodity/get_all_categories', params)
}

export interface GoodsQueryParams {
  shopname: string
  page: number
  page_size: number
  demand: string
  category: string
  label_one?: string[]
  label_two?: string[]
  label_three?: string[]
  label_four?: string[]
  label_seven?: string[]
  begin_time?: string
  end_time?: string
  status?: 'online' | 'offline'
  style_code?: string
}

export interface GoodsQueryItem {
  name: string
  price: number
  promo_image_url: string
  style_code: string
  created_at: string
  online_status: string
  online_time: string
}

export interface GoodsQueryResponse {
  code: number
  data: {
    data: GoodsQueryItem[]
    page: number
    page_size: number
    pages: number
    total: number
  }
  msg: string
}

export const goodsQuery = (params: GoodsQueryParams) => {
  return http.post<GoodsQueryResponse>('/commodity/goods_query', params)
}

export interface StyleCodeCommodityParams {
  shopname: string
  style_code: string
}

export interface StyleCodeCommodityImage {
  created_at: string
  id: any
  is_main: boolean
  url: string
}

export interface StyleCodeCommoditySize {
  commodity_id: string
  inventory: number
  size: string
}

export interface StyleCodeCommodityItem {
  color: string
  color_image: string
  sizes: StyleCodeCommoditySize[]
}

export interface StyleCodeCommodityLabels {
  label_five?: string
  label_four?: string
  label_one?: string
  label_seven?: string
  label_six?: string
  label_three?: string
  label_two?: string
  [key: string]: string | undefined
}

export interface StyleCodeCommodityData {
  category: string
  display_pictures: any
  images: StyleCodeCommodityImage[]
  inventory: number
  items: StyleCodeCommodityItem[]
  labels: StyleCodeCommodityLabels
  main_image: StyleCodeCommodityImage
  name: string
  other_images: any
  price: number
}

export interface StyleCodeCommodityResponse {
  code: number
  data: StyleCodeCommodityData
  msg: string
}

export const getStyleCodeCommodity = (params: StyleCodeCommodityParams) => {
  return http.post<StyleCodeCommodityResponse>('/commodity/stylecode_commodities', params)
}

export interface UpdateStyleCodeInfoParams {
  shopname: string
  style_code: string
  name?: string
  category?: string
  price?: number
  labels?: StyleCodeCommodityLabels
  image?: string
  display_pictures?: any
}

export interface UpdateStyleCodeInfoResponse {
  code: number
  data: any
  msg: string
}

export const updateStyleCodeInfo = (params: UpdateStyleCodeInfoParams) => {
  return http.post<UpdateStyleCodeInfoResponse>('/commodity/update_style_code_info', params)
}

export interface ActivityImageItem {
  category: string
  commodities: any
  created_at: string
  id: number
  image: string
  notes: string
  offline_time: string
  online_time: string
  order: number
  status: 'online' | 'offline' | 'pending'
  style_codes: string[] | null
  updated_at: string
  has_activity_detail: boolean
}

export interface ActivityImageParams {
  page?: number
  pageSize?: number
  status?: string
  start_time?: string
  end_time?: string
  shopname?: string
  has_activity_detail?: boolean
}

export interface ActivityImageResponse {
  code: number
  data: {
    items: ActivityImageItem[]
    page: number
    pageSize: number
    total: number
  }
  msg: string
}

export const batchQueryActivityImages = (params: ActivityImageParams) => {
  return http.post<ActivityImageResponse>('/activity/batch_query_activity_images', params)
}

export interface AddActivityImgResponse {
  code: number
  data: {
    id: number
    image: string
  }
  msg: string
}

export const addActivityImg = (formData: FormData) => {
  return http.post<AddActivityImgResponse>('/activity/add_activity_img', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export interface ActivityImageActionParams {
  activity_id: number
}

export interface ActivityImageActionResponse {
  code: number
  data: {
    id: number
    order: number
  }
  msg: string
}

export const activityImageOnline = (params: ActivityImageActionParams) => {
  return http.post<ActivityImageActionResponse>('/activity/activity_image_online', params)
}

export const activityImageOffline = (params: ActivityImageActionParams) => {
  return http.post<ActivityImageActionResponse>('/activity/activity_image_offline', params)
}

export interface BatchUpdateOrderItem {
  id: number
  order: number
}

export interface BatchUpdateOrderParams {
  images: BatchUpdateOrderItem[]
}

export interface BatchUpdateOrderResponse {
  code: number
  data: Record<string, any>
  msg: string
}

export const batchUpdateActivityImageOrder = (params: BatchUpdateOrderParams) => {
  return http.post<BatchUpdateOrderResponse>('/activity/batch_update_activity_image_order', params)
}

export interface UpdateActivityImageRelationsParams {
  activity_id: number
  category?: string
  style_codes?: string[]
}

export interface UpdateActivityImageRelationsResponse {
  code: number
  data: { id: number }
  msg: string
}

export const updateActivityImageRelations = (params: UpdateActivityImageRelationsParams) => {
  return http.post<UpdateActivityImageRelationsResponse>('/activity/update_activity_image_relations', params)
}

export interface GetActivityImageDetailParams {
  activity_id: number
}

export interface GetActivityImageDetailResponse {
  code: number
  data: {
    category: string
    commodities: any
    created_at: string
    id: number
    image: string
    notes: string
    offline_time: string
    online_time: string
    order: number
    promotional_pics: any
    status: string
    style_codes: string[] | null
    updated_at: string
    has_activity_detail: boolean
  }
  msg: string
}

export const getActivityImageDetail = (params: GetActivityImageDetailParams) => {
  return http.post<GetActivityImageDetailResponse>('/activity/get_activity_image_detail', params)
}

export interface AddPromotionalPicParams {
  activity_id: number
  image: File
}

export interface AddPromotionalPicResponse {
  code: number
  data: {
    image_url: string
    order: string
    upload_time: string
  }
  msg: string
}

export const addPromotionalPic = (formData: FormData) => {
  return http.post<AddPromotionalPicResponse>('/activity/add_promotional_pic', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export interface UpdatePromotionalPicOrderParams {
  activity_id: number
  old_order: number
  new_order: number
}

export interface UpdatePromotionalPicOrderResponse {
  code: number
  data: Record<string, any>
  msg: string
}

export const updatePromotionalPicOrder = (params: UpdatePromotionalPicOrderParams) => {
  return http.post<UpdatePromotionalPicOrderResponse>('/activity/update_promotional_pic_order', params)
}

export interface SetActivityDetailParams {
  activity_id: number
  has_activity_detail: boolean
}

export interface SetActivityDetailResponse {
  code: number
  data: Record<string, any>
  msg: string
}

export const setActivityDetail = (params: SetActivityDetailParams) => {
  return http.post<SetActivityDetailResponse>('/activity/set_has_activity_detail', params)
}

export interface ReturnOrderItem {
  id: number
  user_id: number
  return_id: string
  order_id: string
  sub_order_id: string
  product_list: string | any[]
  type: string
  status: string
  express_company: string
  express_number: string
  reason: string
  specific_reasons: string
  buyer_province: string
  buyer_city: string
  buyer_county: string
  buyer_address: string
  buyer_phone: string
  remarks: string
  request_time: string
  shipped_time: string
  completed_time: string
  canceled_time: string
}

export interface ReturnOrderQueryParams {
  return_order_id?: string
  order_id?: string
  user_id?: number
  status?: string
  page: number
  page_size: number
}

export interface ReturnOrderQueryResponse {
  code: number
  data: {
    return_orders: ReturnOrderItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export interface ReturnOrderApproveParams {
  return_order_id: string
  approve_status: 'approved' | 'rejected'
  user_id: number
  remark?: string
}

export interface ReturnOrderReceiveParams {
  return_order_id: string
  user_id: number
}

export interface ReturnReasonRankItem {
  reason: string
  count: number
}

export interface ReturnOrderStatisticsParams {
  begin_time?: string
  end_time?: string
}

export interface ReturnOrderStatisticsData {
  total_count: number
  pending_count: number
  completed_count: number
  after_sale_rate: number
  after_sale_amount: number
  reason_rank: ReturnReasonRankItem[]
  completed_orders: number
  after_sale_orders: number
}

export interface ReturnOrderStatisticsResponse {
  code: number
  data: {
    statistics: ReturnOrderStatisticsData
  }
  msg: string
}

export const queryReturnOrders = (params: ReturnOrderQueryParams) => {
  return http.post<ReturnOrderQueryResponse>('/return_order/query', params)
}

export const approveReturnOrder = (params: ReturnOrderApproveParams) => {
  return http.post('/return_order/approve', params)
}

export const receiveReturnOrder = (params: ReturnOrderReceiveParams) => {
  return http.post('/return_order/receive', params)
}

export const queryReturnOrderStatistics = (params: ReturnOrderStatisticsParams) => {
  return http.post<ReturnOrderStatisticsResponse>('/return_order/statistics', params)
}

export interface InventoryQueryParams {
  commodity_id?: string
  style_code?: string
}

export interface InventoryCommodity {
  commodity_id: string
  name: string
  style_code: string
  category: string
  price: number
  inventory: number
  size?: string
  color?: string
}

export interface InventoryQueryResponse {
  code: number
  data: {
    commodity?: InventoryCommodity
    commodities?: InventoryCommodity[]
    total_inventory?: number
    style_code?: string
  }
  msg: string
}

export interface InventoryWarningsParams {
  threshold?: number
  page: number
  page_size: number
}

export interface InventoryWarningsResponse {
  code: number
  data: {
    data: InventoryCommodity[]
    total: number
    threshold: number
    page: number
    page_size: number
  }
  msg: string
}

export interface InventoryLogItem {
  id: number
  commodity_id: string
  style_code: string
  warehouse_code: string
  before_qty: number
  change_qty: number
  after_qty: number
  change_type: string
  related_order_id: string
  related_sub_order_id: string
  related_return_id: string
  operator_id: string
  remark: string
  created_at: string
}

export interface InventoryLogsParams {
  commodity_id?: string
  style_code?: string
  change_type?: string
  related_order_id?: string
  related_sub_order_id?: string
  related_return_id?: string
  page: number
  page_size: number
}

export interface InventoryLogsResponse {
  code: number
  data: {
    data: InventoryLogItem[]
    total: number
    page: number
    page_size: number
  }
  msg: string
}

export interface InventoryAdjustParams {
  commodity_id: string
  change_qty: number
  operator_id?: string
  warehouse_code?: string
  remark?: string
}

export interface InventoryTransferParams {
  commodity_id: string
  qty: number
  source_warehouse_code: string
  target_warehouse_code: string
  operator_id?: string
  remark?: string
}

export interface InventoryStockCheckParams {
  commodity_id: string
  actual_qty: number
  warehouse_code?: string
  operator_id?: string
  remark?: string
}

export const queryInventory = (params: InventoryQueryParams) => {
  return http.post<InventoryQueryResponse>('/inventory/query', params)
}

export const queryInventoryWarnings = (params: InventoryWarningsParams) => {
  return http.post<InventoryWarningsResponse>('/inventory/warnings', params)
}

export const queryInventoryLogs = (params: InventoryLogsParams) => {
  return http.post<InventoryLogsResponse>('/inventory/logs', params)
}

export const adjustInventory = (params: InventoryAdjustParams) => {
  return http.post('/inventory/adjust', params)
}

export const transferInventory = (params: InventoryTransferParams) => {
  return http.post('/inventory/transfer', params)
}

export const stockCheckInventory = (params: InventoryStockCheckParams) => {
  return http.post('/inventory/stock_check', params)
}
