import http from './request'

export const getToken = () => {
  return http.getToken()
}

export const login = (username: string, password: string) => {
  return http.post('/login', { username, password })
}

export interface OrderQueryParams {
  page: number
  page_size: number
  shopname: string
  status?: string
  begin_time?: string
  end_time?: string
  tid?: string
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

export const batchGetProducts = (params: BatchProductParams) => {
  return http.post<BatchProductResponse>('/commodity/batch_get_products_by_ids', params)
}

export const getProductList = (params: any) => {
  return http.get('/product/list', { params })
}

export const getMemberList = (params: any) => {
  return http.get('/member/list', { params })
}

export const getDashboardData = (params: any) => {
  return http.get('/dashboard/data', { params })
}

export const getReportData = (params: any) => {
  return http.get('/report/data', { params })
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
  operator_id: number
}

export const updatePaymentAmount = (params: UpdatePaymentAmountParams) => {
  return http.post('/order/update_payment_amount', params)
}

export interface ConfirmPaymentParams {
  order_id: string
  operator_id: number
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
  return http.post('/order/deliver', params)
}

export interface ReceiveOrderParams {
  order_id: string
  user_id: number
}

export const receiveOrder = (params: ReceiveOrderParams) => {
  return http.post('/order/order_receive', params)
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
