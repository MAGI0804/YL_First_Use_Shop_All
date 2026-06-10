
# Member_Shop 接口文档

生成日期：2026-05-15  
项目路径：`D:\youlan_kids_shop_self\Member_Shop`  
默认端口：`3088`  
来源：`OPEN_FLOW.md`、`routes/*.go`、`requestbody/*.go`、`controllers/*.go`、`service/method/*.go`

## 1. 开发完成度结论

`OPEN_FLOW.md` 中的后端主线已经基本落地：

| 模块 | 状态 | 依据 | 说明 |
| --- | --- | --- | --- |
| 库存体系 | 基本完成 | `inventory_method.go`、`inventory_route.go`、`inventory_log.go` | 已有统一库存变动、库存日志、查询、调整、预警、下单扣减、订单/子订单取消回滚、售后完成回滚。 |
| 评价体系 | 基本完成 | `review_method.go`、`review_route.go`、`product_review.go`、`review_reply.go` | 已有创建、前台查询、后台查询、审核、回复、统计。 |
| 售后体系 | 基本完成 | `return_order.go`、`order_method.go`、`return_order_route.go` | 已有统一创建入口、重复售后拦截、审核、买家寄回、仓库收货完成、订单/子订单展示字段。 |
| 数据分析 | 第一版完成 | `analytics_method.go`、`analytics_route.go` | 已有销售、用户、商品、导出；流量分析为预留接口。 |
| 迁移 | 已接入 | `db/migrations.go` | 已包含 `InventoryLog`、`ProductReview`、`ReviewReply` 等新增模型。 |

未完成/预留项：

| 项目 | 状态 | 说明 |
| --- | --- | --- |
| `/inventory/sync_jushuitan` | 预留 | 当前返回 HTTP `501` 和 `reserved`，真实聚水潭库存同步未实现。 |
| `/analytics/traffic_summary` | 预留 | 当前没有页面访问埋点，按设计返回 `not_implemented`。 |
| 自动化测试 | 缺失 | `go test ./...` 编译通过，但各包均为 `[no test files]`。 |

验证命令：`go test ./...`，结果：全部包编译通过。

## 2. 通用约定

Base URL：`http://localhost:3088`

除文件上传外，默认请求头：`Content-Type: application/json`。文件上传使用 `multipart/form-data`。

多数接口成功返回：

```json
{"code":200,"msg":"success","data":{}}
```

多数接口失败返回：

```json
{"code":201,"msg":"invalid request","data":{},"Err":"字段校验或业务错误"}
```

字段含义：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `code` | int | 业务码，成功通常为 `200`，失败通常为 `201`。 |
| `msg` | any | 成功或失败消息。 |
| `data` | object | 数据对象，无数据时为空对象。 |
| `Err` | any | 失败详情，可能是字符串或字段校验对象。 |

少数历史接口直接返回：`{"error":"无效的JSON格式"}`。建议前端同时判断 HTTP 状态和响应体 `code`。

## 3. 小程序端接口

### 3.1 用户与会员

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /ordinary_user/wechat_login` | 微信会员手机号授权登录 | `{ "code": "wx_login_code", "phone_code": "getPhoneNumber_code", "userInfo": { "nickName": "小兰", "avatarUrl": "https://example.com/a.png" } }` |
| `POST /ordinary_user/send_register_captcha` | 发送手机号验证码 | `{ "mobile": "13800000000" }` |
| `POST /ordinary_user/bind_wechat_phone` | 绑定微信 openid 和已有会员手机号 | `{ "openid": "openid_xxx", "mobile": "13800000000", "captcha": "123456", "nickname": "小兰", "avatar_url": "https://example.com/a.png" }` |
| `POST /ordinary_user/find_data` | 查询用户扩展数据 | `{ "user_id": 10001 }` |
| `POST /ordinary_user/add_data` | 新增用户扩展数据 | `{ "user_id": 10001, "data_type": "favorite", "data_value": "dress" }` |
| `POST /ordinary_user/Modify_data` | 修改用户资料，支持 JSON 或表单头像上传 | `{ "user_id": 10001, "nickname": "小兰", "province": "浙江省", "city": "杭州市", "county": "西湖区", "detailed_address": "文一路1号" }` |
| `POST /ordinary_user/get_user_id` | 根据手机号查询用户 ID | `{ "mobile": "13800000000" }` |
| `POST /ordinary_user/update_platform_info` | 更新会员天猫/有赞信息 | `{ "user_id": 10001, "tmall_id": "tm001", "tmall_amount": 100, "youzan_id": "yz001", "youzan_amount": 50 }` |
| `POST /ordinary_user/member_amount_summary` | 会员金额汇总 | `{ "user_id": 10001, "member_no": "M001", "mobile": "13800000000" }` |

字段含义：`code` 为 `wx.login` 返回的微信临时登录 code；`phone_code` 为小程序 `button open-type="getPhoneNumber"` 返回的动态 code，后端通过微信 `getuserphonenumber` 接口换取手机号；`openid` 为微信 openid；`mobile` 为手机号；`captcha` 为验证码；`user_id` 为系统用户 ID；`member_no` 为会员编号；`tmall_id/youzan_id` 为外部平台 ID；`tmall_amount/youzan_amount` 为外部平台消费金额。

后台会员维护规则：`POST /member/update` 可维护会员金额字段，包括 `total_order_amount`、`total_paid_amount`、`tmall_amount`、`youzan_amount`，用于运营修正历史导入或线下核对后的金额。

会员登录规则：

- 登录页第一步只触发手机号授权登录；登录成功后再显示头像昵称补全区。
- 头像使用 `open-type="chooseAvatar"`，昵称使用 `input type="nickname"`。按微信小程序能力限制，头像和昵称不能合并成一个手机号授权弹窗；可在资料补全区提供“保存并进入”和“暂不完善，直接进入”。
- `wechat_login` 必须同时收到 `code` 和 `phone_code`，后端先换 openid，再换手机号。
- 手机号必须已经存在于 `member_info` 且会员状态不是 `disabled`；非会员手机号返回 `403`。
- 首次登录成功后，`member_info.mobile`、`member_info.user_id`、`member_info.openid` 与 `users_user.mobile` 绑定；一个手机号只对应一个 `user_id`，不能再绑定其他微信 openid。
- 小程序端手机号登录成功后保存 `token`、`refresh_token`、`user_id`、`userInfo`；临时头像路径通过 `Modify_data` 上传后再写入用户资料。

微信登录成功示例：

```json
{"code":200,"message":"登录成功","data":{"token":{"access":"jwt_access","refresh":"jwt_refresh"},"user_id":10001,"member_no":"M001","mobile":"13800000000","phone_bound":true,"nickname":"小兰"}}
```

绑定手机号成功示例：

```json
{"code":200,"msg":"Wechat phone bound successfully","data":{"user_id":10001,"member_no":"M001","openid":"openid_xxx","mobile":"13800000000","nickname":"小兰"}}
```

失败示例：

```json
{"code":403,"message":"该手机号不是会员，请联系商家开通会员"}
```

### 3.2 商品与活动展示

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /commodity/goods_query_wx` | 小程序商品列表 | `{ "shopname": "youlan_kids", "demand": "裙", "category": "童装", "status": "online", "page": 1, "page_size": 20 }` |
| `POST /commodity/goods_query` | 商品列表/后台可复用 | `{ "shopname": "youlan_kids", "style_code": "ST001", "label_one": ["夏季"], "page": 1, "page_size": 20 }` |
| `POST /commodity/stylecode_commodities` | 按款号查 SKU | `{ "shopname": "youlan_kids", "style_code": "ST001" }` |
| `POST /commodity/search_products_by_name` | 按商品名搜索 | `{ "search_str": "连衣裙", "page": 1, "page_size": 20 }` |
| `POST /commodity/batch_get_products_by_ids` | 批量查商品 | `{ "commodity_ids": ["SKU001", "SKU002"] }` |
| `POST /commodity/search_commodity_data` | 查询商品指定字段 | `{ "commodity_id": "SKU001", "data_list": ["name", "price", "inventory"] }` |
| `POST /commodity/get_all_categories` | 获取分类 | `{ "shopname": "youlan_kids" }` |
| `POST /commodity/get_all_labels` | 获取标签 | `{ "shopname": "youlan_kids", "category": "童装" }` |
| `POST /commodity/search_style_codes` | 搜索款号 | `{ "shopname": "youlan_kids", "search_keyword": "裙", "category": "童装", "page": 1, "page_size": 20 }` |
| `POST /commodity/get_commodity_status` | 查询商品状态 | `{ "commodity_id": "SKU001" }` |
| `POST /activity/query_online_activity_images` | 查询上线活动图 | `{}` |
| `POST /activity/get_activity_image_detail` | 活动图详情 | `{ "activity_id": 1 }` |

字段含义：`shopname` 店铺名；`demand/search_keyword/search_str` 搜索词；`commodity_id` SKU；`commodity_ids` SKU 列表；`style_code` 款号；`category` 分类；`status` 状态；`label_one` 至 `label_seven` 标签筛选；`page/page_size` 分页；`activity_id` 活动图 ID。

商品查询成功示例：

```json
{"code":200,"msg":"查询成功","data":{"data":[{"commodity_id":"SKU001","name":"儿童连衣裙","price":199,"style_code":"ST001","inventory":10,"status":"online"}],"total":1}}
```

失败示例：

```json
{"code":201,"msg":"参数不正确","data":{},"Err":""}
```

### 3.3 购物车

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /cart/add_to_cart` | 加入购物车 | `{ "user_id": 10001, "commodity_code": "SKU001", "quantity": 2 }` |
| `GET/POST /cart/query_cart_items` | 查询购物车 | `{ "user_id": 10001 }` |
| `PUT/POST /cart/update_cart_item_quantity` | 更新数量 | `{ "user_id": 10001, "commodity_code": "SKU001", "quantity": 3 }` |
| `POST /cart/increase_cart_item_quantity` | 数量加 1 | `{ "user_id": 10001, "commodity_code": "SKU001" }` |
| `POST /cart/decrease_cart_item_quantity` | 数量减 1 | `{ "user_id": 10001, "commodity_code": "SKU001" }` |
| `DELETE/POST /cart/batch_delete_from_cart` | 批量删除 | `{ "user_id": 10001, "commodity_codes": ["SKU001"] }` |
| `DELETE/POST /cart/clear_cart` | 清空购物车 | `{ "user_id": 10001 }` |

字段含义：`commodity_code` 为商品标识，添加购物车时兼容 `Commodity_data.commodity_id`、`spec_code`、`style_code`，入库后统一使用真实 `commodity_id` 作为购物车 key；`quantity` 为数量；`commodity_codes` 为待删除商品标识列表。

后台代会员下单规则：从会员详情购物车下单时，后台只提交已勾选的购物车商品；收货地址可选择该用户已有地址，也可粘贴整段地址后识别并填入省市区、详细地址、收货人和手机号。

成功示例：`{"code":200,"msg":"操作成功","data":{}}`  
失败示例：`{"code":201,"msg":"invalid request","data":{},"Err":"commodity_code为必填字段"}`

### 3.4 收货地址

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /address/add_address` | 新增地址 | `{ "user_id": 10001, "province": "浙江省", "city": "杭州市", "county": "西湖区", "detailed_address": "文一路1号", "receiver_name": "张三", "phone_number": "13800000000", "is_default": true, "remark": "家" }` |
| `POST /address/get_addresses` | 地址列表 | `{ "user_id": 10001 }` |
| `POST /address/get_address_by_id` | 地址详情 | `{ "address_id": 1, "user_id": 10001 }` |
| `POST /address/update_address` | 更新地址 | `{ "address_id": 1, "user_id": 10001, "receiver_name": "李四", "phone_number": "13900000000" }` |
| `POST /address/set_default_address` | 设置默认地址 | `{ "address_id": 1, "user_id": 10001 }` |
| `POST /address/delete_address` | 删除地址 | `{ "address_id": 1, "user_id": 10001 }` |

字段含义：`address_id` 地址 ID；`receiver_name` 收货人；`phone_number` 收货手机号；`is_default` 是否默认。  
小程序微信地址同步：地址管理页通过用户点击按钮调用微信官方 `wx.chooseAddress`，将返回的 `userName`、`telNumber`、`provinceName`、`cityName`、`countyName`、`detailInfo` 分别映射为 `receiver_name`、`phone_number`、`province`、`city`、`county`、`detailed_address` 后调用 `/address/add_address`；用户第一条地址自动设置为默认地址。  
官方文档：`https://developers.weixin.qq.com/miniprogram/dev/api/open-api/address/wx.chooseAddress.html`  
成功示例：`{"code":200,"msg":"success","data":{"address_id":1}}`  
失败示例：`{"code":201,"msg":"地址不存在","data":{},"Err":""}`
### 3.5 订单与售后

#### 3.5.1 创建订单

`POST /order/add_order`

请求示例：

```json
{
  "user_id": 10001,
  "receiver_name": "张三",
  "receiver_phone": "13800000000",
  "province": "浙江省",
  "city": "杭州市",
  "county": "西湖区",
  "detailed_address": "文一路1号",
  "order_amount": 398,
  "product_list": [
    {"commodity_id": "SKU001", "product_name": "儿童连衣裙", "qty": 2, "price": 199, "sub_amount": 398}
  ],
  "express_company": "顺丰",
  "express_number": "",
  "remark": "尽快发货"
}
```

字段含义：`user_id` 下单用户；`receiver_*` 收货信息；`order_amount` 订单金额；`product_list` 商品列表；`commodity_id/sku_id/product_id/id` 用于识别 SKU；`qty/quantity/num` 为数量；`remark` 为备注。

关键规则：只有已绑定 active 会员的 `user_id` 才能下单；当前实现采用下单扣库存；库存不足时订单创建失败，不生成错误库存。

成功示例：

```json
{"code":200,"msg":"创建成功","data":{"order_id":"Y2026051500000001","sub_order_ids":["S2026051500000001"],"status":"pending","pay_status":"unpaid"}}
```

失败示例：

```json
{"code":201,"msg":"商品SKU001库存不足，当前库存0，需要2","data":{},"Err":""}
```

#### 3.5.2 订单查询和操作

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /order/query_order_data` | 订单详情 | `{ "order_id": "Y2026051500000001", "user_id": 10001 }` |
| `POST /order/orders_query` | 订单列表 | `{ "shopname": "youlan_kids", "user_id": 10001, "status": "pending", "page": 1, "page_size": 20, "begin_time": "2026-05-01", "end_time": "2026-05-15", "tid": "Y20260515" }` |
| `POST /order/batch_orders_query` | 批量订单查询，当前复用列表逻辑 | 同 `/order/orders_query` |
| `POST /order/query_by_user_id` | 按用户查询订单 | `{ "shopname": "youlan_kids", "user_id": 10001, "status": "paid", "page": 1, "page_size": 20 }` |
| `POST /order/pay` | 支付订单 | `{ "order_id": "Y2026051500000001", "user_id": 10001 }` |
| `POST /order/cancel` | 取消订单并回滚库存 | `{ "order_id": "Y2026051500000001", "user_id": 10001 }` |
| `POST /order/order_receive` | 确认收货 | `{ "order_id": "Y2026051500000001", "user_id": 10001 }` |
| `POST /order/search_by_product_name` | 按商品名搜索订单 | `{ "shopname": "youlan_kids", "product_name": "连衣裙", "page": 1, "page_size": 20 }` |

主要字段含义：`status` 为订单状态；`tid` 为订单号模糊搜索；`begin_time/end_time` 为下单时间范围；`page/page_size` 为分页。

订单详情成功示例：

```json
{"code":200,"msg":"查询成功","data":{"order":{"order_id":"Y2026051500000001","status":"pending","pay_status":"unpaid","after_sale_status":"","is_after_sale_completed":false,"display_gray":false}}}
```

失败示例：

```json
{"code":201,"msg":"订单不存在","data":{},"Err":""}
```

#### 3.5.3 子订单

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /order/query_sub_order_data` | 查询主订单下子订单 | `{ "order_id": "Y2026051500000001" }` |
| `POST /order/change_sub_order_status` | 修改子订单状态 | `{ "sub_order_id": "S2026051500000001", "status": "paid" }` |
| `POST /order/cancel_sub_order` | 取消未发货子订单并回滚库存 | `{ "sub_order_id": "S2026051500000001", "user_id": 10001, "reason": "不想要了" }` |
| `POST /order/return_sub_order` | 子订单发起售后 | `{ "sub_order_id": "S2026051500000001", "user_id": 10001, "reason": "尺码不合适", "specific_reasons": "偏小", "buyer_province": "浙江省", "buyer_city": "杭州市", "buyer_county": "西湖区", "buyer_address": "文一路1号", "buyer_phone": "13800000000" }` |

关键规则：子订单只有 `pending` 或 `paid` 且未发货时可直接取消；已发货子订单需走售后流程。返回字段会附加 `after_sale_status`、`is_after_sale_completed`、`display_gray`。

成功示例：`{"code":200,"msg":"操作成功","data":{}}`  
失败示例：`{"code":201,"msg":"已发货子订单不能直接取消，请走售后流程","data":{},"Err":""}`

#### 3.5.4 申请售后

`POST /order/request_return` 或 `POST /return_order/create`

请求示例：

```json
{
  "order_id": "Y2026051500000001",
  "user_id": 10001,
  "sub_order_id": "S2026051500000001",
  "type": "return",
  "reason": "尺码不合适",
  "specific_reasons": "偏小，需要退货",
  "product_ids": ["SKU001"],
  "buyer_province": "浙江省",
  "buyer_city": "杭州市",
  "buyer_county": "西湖区",
  "buyer_address": "文一路1号",
  "buyer_phone": "13800000000",
  "remark": "请尽快处理"
}
```

字段含义：`order_id` 主订单号；`sub_order_id` 子订单号，非必填；`type` 售后类型，支持 `return/exchange/refund`；`return_type/return_reason` 为兼容旧字段；`product_ids` 用于整单售后筛选商品；`buyer_*` 为买家退货信息。创建成功后后端会立即调用聚水潭 `/open/aftersale/upload` 上传售后单，返回字段包含 `jushuitan_push_status`、`jushuitan_after_sale_id`、`jushuitan_push_response`，用于后续补偿重推和排查。

成功示例：

```json
{"code":200,"msg":"创建成功","data":{"return_order_id":"T2026051500000001","status":"pending","jushuitan_push_status":"success","jushuitan_after_sale_id":"100001"}}
```

失败示例：

```json
{"code":201,"msg":"该商品或子订单存在未结束售后","data":{},"Err":""}
```

#### 3.5.5 售后物流、取消、查询、聚水潭回写

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /return_order/deliver` | 买家寄回物流 | `{ "return_order_id": "T2026051500000001", "user_id": 10001, "express_company": "顺丰", "express_number": "SF123" }` |
| `POST /return_order/receive` | 商城确认退款完成；若 ERP 未先回写 `received`，则兼容旧流程完成库存回滚 | `{ "return_order_id": "T2026051500000001", "user_id": 10001 }` |
| `POST /return_order/cancel` | 取消售后 | `{ "return_order_id": "T2026051500000001", "user_id": 10001, "reason": "不退了" }` |
| `POST /return_order/update_buyer_info` | 修改买家退货信息 | `{ "return_order_id": "T2026051500000001", "user_id": 10001, "buyer_province": "浙江省", "buyer_city": "杭州市", "buyer_county": "西湖区", "buyer_address": "文一路1号", "buyer_phone": "13800000000" }` |
| `POST /return_order/query` | 售后列表 | `{ "return_order_id": "", "order_id": "Y2026051500000001", "user_id": 10001, "status": "pending", "page": 1, "page_size": 20 }` |
| `POST /return_order/detail` | 售后详情 | `{ "return_order_id": "T2026051500000001" }` |
| `POST /return_order/push_jushuitan` | 售后上传失败后手动重推聚水潭 | `{ "return_order_id": "T2026051500000001" }` |
| `POST /return_order/jushuitan_after_sale_push` | 聚水潭售后状态主动推送回写 | `biz={"outer_as_id":"T2026051500000001","as_id":"100001","so_id":"Y2026051500000001","status":"退货入库"}` |
| `POST /return_order/jushuitan_after_sale_received_query` | 查询聚水潭实际收货，供推送丢失时补偿 | `{ "outer_as_id": "T2026051500000001", "page_index": 1, "page_size": 50 }` |

售后状态：`pending` 已申请并待 ERP 审核；`approved` ERP 审核通过；`rejected` ERP 拒绝；`buyer_shipped` 买家已寄回；`received` ERP 退货入库且本地已回滚退货库存；`completed` 商城同意退款/最终完成；`canceled` 取消或关闭。

成功示例：

```json
{"code":200,"msg":"success","data":{"return_order":{"return_id":"T2026051500000001","status":"completed","after_sale_status":"completed","is_after_sale_completed":true,"display_gray":true}}}
```

失败示例：

```json
{"code":201,"msg":"退货订单状态不允许签收","data":{},"Err":""}
```

### 3.6 评价

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /review/create` | 用户创建评价 | `{ "user_id": 10001, "order_id": "Y2026051500000001", "sub_order_id": "S2026051500000001", "commodity_id": "SKU001", "style_code": "ST001", "rating": 5, "content": "质量很好", "images": ["https://example.com/1.jpg"], "tags": ["质量好"] }` |
| `POST /review/query_by_product` | 商品详情页评价列表，只返回 approved | `{ "commodity_id": "SKU001", "style_code": "ST001", "page": 1, "page_size": 20 }` |
| `POST /review/statistics` | 商品/款号评价统计 | `{ "commodity_id": "SKU001", "style_code": "ST001" }` |

字段含义：`rating` 为 1 到 5 分；`images/tags` 为字符串数组；`commodity_id` 与 `style_code` 查询时至少传一个。

关键规则：用户、订单、子订单、商品必须存在且归属一致；子订单状态必须为 `delivered/completed/signed/received`；同一 `sub_order_id + commodity_id` 只能评价一次；新评价默认 `pending`。

成功示例：

```json
{"code":200,"msg":"success","data":{"review":{"id":1,"user_id":10001,"commodity_id":"SKU001","rating":5,"status":"pending"}}}
```

失败示例：

```json
{"code":201,"msg":"sub_order commodity has already been reviewed","data":{},"Err":""}
```

### 3.7 消息

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /message/categories` | 消息分类和各分类最后一条消息 | `{ "user_id": 10001 }` |
| `POST /message/query` | 按分类分页查询消息 | `{ "user_id": 10001, "message_type": "order", "page": 1, "page_size": 20 }` |

字段含义：`message_type` 为消息类型；`page_size` 最大 50。

成功示例：

```json
{"code":200,"msg":"success","data":{"messages":[],"total":0}}
```

失败示例：

```json
{"code":201,"msg":"invalid request","data":{},"Err":"user_id为必填字段"}
```
## 4. 后台端接口

### 4.1 Access Token 与健康检查

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /access_token/get_token` | 按客户端 IP 获取或生成访问令牌 | `{}` |
| `POST /access_token/get_ips` | 查询已注册 IP，要求 `shopname=youlan_kids` | `{ "shopname": "youlan_kids" }` |
| `GET /api/test/` | 服务测试 | 无请求体 |
| `GET /api/health/` | 健康检查 | 无请求体 |

字段含义：`shopname` 为店铺名，`get_ips` 必须传 `youlan_kids`。

Token 规则：

- 除白名单接口外，小程序请求必须在 URL query 或表单中携带 `access_token`。
- `POST /access_token/get_token` 按客户端 IP 获取或生成 token；线上反向代理场景下，发 token 和验 token 都优先使用 `X-Forwarded-For` 的第一个 IP，避免同一请求链路因 IP 口径不一致返回 `401`。
- 小程序端本地缓存的 `access_token` 只作为加速使用；普通接口遇到 HTTP `401/402` 或响应体 `code=401/402` 时，应清除缓存、重新调用 `get_token`，并对原请求最多重试一次。

成功示例：

```json
{"code":200,"msg":"申请成功","data":{"access_token":"32位token","RegisterTime":"2026-05-15T10:00:00+08:00","ip_addresses":"127.0.0.1"}}
```

失败示例：

```json
{"code":201,"msg":"参数不正确","data":{},"Err":""}
```

### 4.2 后台账号

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /OperationUser/add_service_user` | 添加客服用户 | `{ "nickname": "客服A", "mobile": "13800000001", "password": "123456" }` |
| `POST /OperationUser/add_operation_user` | 添加运营用户 | `{ "nickname": "运营A", "mobile": "13800000002", "password": "123456", "level": 1 }` |
| `POST /OperationUser/verification_status` | 验证登录状态/密码 | `{ "mobile": "13800000002", "password": "123456", "object_num": "1" }` |
| `POST /OperationUser/change_password` | 修改密码 | `{ "object_num": 1, "mobile": "13800000002", "old_password": "123456", "new_password": "abcdef" }` |
| `POST /OperationUser/send_register_captcha` | 发送后台注册验证码 | `{ "mobile": "13800000002" }` |
| `POST /OperationUser/backend_register_by_phone` | 手机验证码注册后台账号 | `{ "mobile": "13800000002", "captcha": "123456", "nickname": "运营A", "password": "abcdef", "role": "operation", "level": 1, "remarks": "" }` |

字段含义：`mobile` 手机号；`password/old_password/new_password` 密码；`object_num` 后台用户编号；`level` 权限等级；`role` 角色。

成功示例：

```json
{"code":200,"msg":"success","data":{"object_num":1}}
```

失败示例：

```json
{"code":201,"msg":"手机号或密码错误","data":{},"Err":""}
```

### 4.3 商品管理

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /commodity/add_goods` | 新增商品，`multipart/form-data` | `commodity_id=SKU001&name=儿童连衣裙&price=199&category=童装&style_code=ST001&size=110&notes=夏季新款` |
| `POST /commodity/delete_goods` | 删除商品 | `{ "commodity_id": "SKU001" }` |
| `POST /commodity/search_commodity_data` | 查询指定商品字段 | `{ "commodity_id": "SKU001", "data_list": ["name", "price", "inventory"] }` |
| `POST /commodity/change_commodity_data` | 修改商品资料 | `{ "commodity_id": "SKU001", "update_fields": { "name": "新版连衣裙", "price": 189, "inventory": 20 } }` |
| `POST /commodity/change_commodity_status_online` | SKU 上架 | `{ "commodity_id": "SKU001" }` |
| `POST /commodity/change_commodity_status_offline` | SKU 下架 | `{ "commodity_id": "SKU001" }` |
| `POST /commodity/stylecode_status_online` | 款号上架 | `{ "style_code": "ST001" }` |
| `POST /commodity/stylecode_status_offline` | 款号下架 | `{ "style_code": "ST001" }` |
| `POST /commodity/update_style_code_info` | 更新款号聚合信息 | `{ "shopname": "youlan_kids", "style_code": "ST001", "name": "新款名", "category": "童装", "price": 199, "label_one": "夏季" }` |

字段含义：`commodity_id` SKU；`style_code` 款号；`update_fields` 为要修改的字段和值；`price` 价格；`inventory` 库存；`label_one` 至 `label_seven` 标签。

成功示例：

```json
{"code":200,"msg":"操作成功","data":{"commodity_id":"SKU001"}}
```

失败示例：

```json
{"code":201,"msg":"商品不存在","data":{},"Err":""}
```

### 4.4 活动图与宣传图管理

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /activity/add_activity_img` | 新增活动图，`multipart/form-data` | `category=童装&notes=首页轮播&commodities=1,2,3&image=<file>` |
| `POST /activity/update_activity_image_relations` | 更新活动图关联款号/分类 | `{ "activity_id": 1, "style_codes": ["ST001"], "category": "童装" }` |
| `POST /activity/activity_image_online` | 上线活动图 | `{ "activity_id": 1 }` |
| `POST /activity/activity_image_offline` | 下线活动图 | `{ "activity_id": 1 }` |
| `POST /activity/batch_query_activity_images` | 后台分页查询活动图 | `{ "page": 1, "pageSize": 20, "status": "online", "start_time": "2026-05-01 00:00:00", "end_time": "2026-05-15 23:59:59", "has_activity_detail": true }` |
| `POST /activity/batch_update_activity_image_order` | 批量更新活动图顺序 | `{ "images": [{ "id": 1, "order": 1 }] }` |
| `POST /activity/set_has_activity_detail` | 设置是否有活动详情 | `{ "activity_id": 1, "has_activity_detail": true }` |
| `POST /activity/add_promotional_pic` | 为活动图新增宣传图 | `{ "activity_id": 1 }` |
| `POST /activity/update_promotional_pic_order` | 调整宣传图位置 | `{ "activity_id": 1, "old_order": 1, "new_order": 2 }` |
| `POST /activity/delete_promotional_pic` | 删除宣传图 | `{ "activity_id": 1, "order": 2 }` |

字段含义：`activity_id` 活动图 ID；`style_codes` 关联款号数组；`images` 为排序数据；`old_order/new_order` 为宣传图位置；`has_activity_detail` 表示是否有活动详情页。

成功示例：

```json
{"code":200,"msg":"更新成功","data":{"id":1}}
```

失败示例：

```json
{"code":201,"msg":"活动图不存在","data":{},"Err":""}
```

### 4.5 订单管理

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /order/orders_query` | 后台订单列表 | `{ "shopname": "youlan_kids", "user_id": 0, "status": "paid", "page": 1, "page_size": 20, "begin_time": "2026-05-01", "end_time": "2026-05-15", "tid": "Y20260515" }` |
| `POST /order/batch_orders_query` | 批量订单查询 | 同 `/order/orders_query` |
| `POST /order/search_by_product_name` | 按商品名搜索订单 | `{ "shopname": "youlan_kids", "product_name": "连衣裙", "page": 1, "page_size": 20 }` |
| `POST /order/change_status` | 修改订单状态 | `{ "order_id": "Y2026051500000001", "status": "processing", "express_company": "顺丰", "express_number": "SF123", "logistics_process": [{"time":"2026-05-15 10:00:00","description":"已出库"}] }` |
| `POST /order/update_express_info` | 更新物流并发货 | `{ "order_id": "Y2026051500000001", "user_id": 10001, "express_company": "顺丰", "express_number": "SF123" }` |
| `POST /order/deliver` | 发货 | `{ "order_id": "Y2026051500000001", "user_id": 10001, "express_company": "顺丰", "express_number": "SF123" }` |
| `POST /order/sync_logistics_info` | 当前路由绑定到发货 controller，谨慎使用 | 同 `/order/deliver` |
| `POST /order/jushuitan_ship_info` | 聚水潭发货回传 | `{ "so_id": "Y2026051500000001", "o_id": 123, "l_id": "SF123", "lc_id": "SF", "order_from": "jushuitan", "wms_co_id": 1, "logistics_company": "顺丰", "send_date": "2026-05-15 10:00:00", "is_send_all": true, "items": [{ "oi_id": 1, "sku_id": "SKU001", "qty": 1, "name": "儿童连衣裙", "outer_oi_id": "S001", "so_id": "Y2026051500000001" }] }` |
| `POST /order/change_receiving_data` | 修改收货信息 | `{ "order_id": "Y2026051500000001", "receiver_name": "李四", "receiver_phone": "13900000000", "province": "浙江省", "city": "杭州市", "county": "西湖区", "detailed_address": "文二路2号" }` |
| `POST /order/update/:id` | 历史订单更新接口 | 路径 `/order/update/1`，请求体按 controller 支持字段传入。 |

字段含义：`status` 支持 `pending/unpaid/paid/partial_paid/processing/shipped/delivered/canceled` 等；`express_company/express_number` 为物流；`jushuitan_ship_info.items` 为聚水潭发货商品行。

成功示例：

```json
{"code":200,"msg":"发货成功","data":{}}
```

失败示例：

```json
{"code":201,"msg":"订单状态不允许发货","data":{},"Err":""}
```

### 4.6 售后管理

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /return_order/approve` | 审核售后 | `{ "return_order_id": "T2026051500000001", "approve_status": "approved", "user_id": 90001, "remark": "同意退货" }` |
| `POST /return_order/query` | 售后列表 | `{ "return_order_id": "", "order_id": "Y2026051500000001", "user_id": 10001, "status": "pending", "page": 1, "page_size": 20 }` |
| `POST /return_order/detail` | 售后详情 | `{ "return_order_id": "T2026051500000001" }` |
| `POST /return_order/deliver` | 填写买家退货物流 | `{ "return_order_id": "T2026051500000001", "user_id": 10001, "express_company": "顺丰", "express_number": "SF123" }` |
| `POST /return_order/receive` | 商城同意退款并完成售后 | `{ "return_order_id": "T2026051500000001", "user_id": 10001 }` |
| `POST /return_order/cancel` | 取消售后 | `{ "return_order_id": "T2026051500000001", "user_id": 10001, "reason": "不退了" }` |
| `POST /return_order/update_buyer_info` | 修改买家退货信息 | `{ "return_order_id": "T2026051500000001", "user_id": 10001, "buyer_province": "浙江省", "buyer_city": "杭州市", "buyer_county": "西湖区", "buyer_address": "文一路1号", "buyer_phone": "13800000000" }` |
| `POST /return_order/push_jushuitan` | 售后上传失败后重推聚水潭 | `{ "return_order_id": "T2026051500000001" }` |
| `POST /return_order/jushuitan_after_sale_push` | 聚水潭售后推送接收 | `biz={"outer_as_id":"T2026051500000001","status":"审核通过"}` |
| `POST /return_order/jushuitan_after_sale_received_query` | 聚水潭实际收货查询 | `{ "outer_as_id": "T2026051500000001", "page_index": 1, "page_size": 50 }` |

字段含义：`approve` 保留为兼容旧后台操作，主流程以聚水潭推送回写为准；聚水潭 `审核通过/同意` 会回写 `approved` 并将订单标记为售后中，`入库/收货` 会回写 `received` 并回滚退货库存，商城最终同意退款时调用 `receive` 置为 `completed`。

成功示例：`{"code":200,"msg":"审核成功","data":{}}`  
失败示例：`{"code":201,"msg":"拒绝售后必须填写拒绝原因","data":{},"Err":""}`
### 4.7 库存管理

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /inventory/query` | 按 SKU 或款号查询库存 | `{ "commodity_id": "SKU001", "style_code": "ST001" }` |
| `POST /inventory/adjust` | 手工调整库存 | `{ "commodity_id": "SKU001", "change_qty": 5, "operator_id": "90001", "remark": "盘点补库存", "warehouse_code": "W001" }` |
| `POST /inventory/logs` | 查询库存日志 | `{ "commodity_id": "SKU001", "style_code": "ST001", "change_type": "manual_adjust", "related_order_id": "Y2026051500000001", "related_sub_order_id": "S2026051500000001", "related_return_id": "T2026051500000001", "page": 1, "page_size": 20 }` |
| `POST /inventory/warnings` | 低库存预警 | `{ "threshold": 5, "page": 1, "page_size": 20 }` |
| `POST /inventory/sync_jushuitan` | 聚水潭库存同步预留 | `{ "commodity_ids": ["SKU001", "SKU002"] }` |

字段含义：`commodity_id` SKU；`style_code` 款号；`change_qty` 变动数量，正数增加、负数减少且不能为 0；`change_type` 支持 `order_deduct/order_cancel_restore/return_restore/manual_adjust/sync_jushuitan`；`threshold` 低库存阈值；`warehouse_code` 仓库编码。

关键规则：库存扣减后不能小于 0；每次变动写入 `InventoryLog`；SKU 库存变更后刷新款号总库存；同一子订单同类扣减/回滚避免重复写入。

库存查询成功示例：

```json
{"code":200,"msg":"success","data":{"commodity":{"commodity_id":"SKU001","inventory":10},"style_code_data":{"style_code":"ST001","inventory":20}}}
```

库存调整失败示例：

```json
{"code":201,"msg":"商品SKU001库存不足，当前库存2，需要5","data":{},"Err":""}
```

聚水潭同步当前返回：

```json
{"code":200,"msg":"sync_jushuitan route is reserved; token-based sync will be implemented in the jushuitan phase","data":{"commodity_ids":["SKU001"],"status":"reserved"}}
```

HTTP 状态码为 `501 Not Implemented`。

### 4.8 评价管理

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /review/query_backend` | 后台评价列表 | `{ "user_id": 10001, "order_id": "Y2026051500000001", "sub_order_id": "S2026051500000001", "commodity_id": "SKU001", "style_code": "ST001", "status": "pending", "page": 1, "page_size": 20 }` |
| `POST /review/audit` | 审核评价 | `{ "review_id": 1, "status": "approved", "audit_remark": "内容正常" }` |
| `POST /review/reply` | 回复评价 | `{ "review_id": 1, "operator_id": "90001", "content": "感谢您的反馈。" }` |
| `POST /review/statistics` | 评价统计 | `{ "commodity_id": "SKU001", "style_code": "ST001" }` |

字段含义：`status` 支持 `pending/approved/rejected/hidden`；`review_id` 评价 ID；`operator_id` 回复操作人；`audit_remark` 审核备注。

成功示例：

```json
{"code":200,"msg":"success","data":{"review":{"id":1,"status":"approved","audit_remark":"内容正常"}}}
```

失败示例：

```json
{"code":201,"msg":"invalid review status","data":{},"Err":""}
```

### 4.9 数据分析

通用请求字段适用于以下 5 个接口：

```json
{
  "begin_time": "2026-05-01",
  "end_time": "2026-05-15",
  "shopname": "youlan_kids",
  "category": "童装",
  "style_code": "ST001",
  "operator_id": "90001",
  "low_inventory_threshold": 5,
  "slow_sales_threshold": 0,
  "limit": 20
}
```

字段含义：`begin_time/end_time` 支持 `YYYY-MM-DD`、`YYYY-MM-DD HH:mm:ss`、RFC3339；`shopname` 映射订单来源；`category/style_code` 用于商品筛选；`low_inventory_threshold` 默认 5；`slow_sales_threshold` 默认 0；`limit` 默认 20，最大 100。

| 接口 | 用途 | 关键返回 |
| --- | --- | --- |
| `POST /analytics/sales_summary` | 销售统计 | `order_count`、`paid_order_count`、`canceled_order_count`、`sales_amount`、`paid_amount`、`refund_amount`、`average_order_value`、`daily` |
| `POST /analytics/user_summary` | 用户分析 | `new_user_count`、`new_member_count`、`order_user_count`、`paid_user_count`、`repurchase_user_count`、偏好列表 |
| `POST /analytics/product_summary` | 商品分析 | `hot_skus`、`hot_style_codes`、`slow_moving_products`、`inventory_turnover_rate`、`low_inventory_count`、`average_rating`、`good_rate` |
| `POST /analytics/traffic_summary` | 流量分析预留 | `status=not_implemented`、`message`、空 `data` |
| `POST /analytics/export` | 结构化导出 JSON | `sales`、`users`、`products`、`traffic` 聚合数据 |

销售统计成功示例：

```json
{"code":200,"msg":"success","data":{"summary":{"order_count":10,"paid_order_count":8,"canceled_order_count":1,"sales_amount":1990,"paid_amount":1592,"refund_amount":199,"average_order_value":199,"daily":[]}}}
```

商品分析成功示例：

```json
{"code":200,"msg":"success","data":{"summary":{"hot_skus":[],"hot_style_codes":[],"slow_moving_products":[],"inventory_turnover_rate":0.25,"low_inventory_count":3,"average_rating":4.8,"good_rate":0.9}}}
```

失败示例：

```json
{"code":201,"msg":"begin_time cannot be after end_time","data":{},"Err":""}
```

### 4.10 后台消息

| 接口 | 用途 | 请求示例 |
| --- | --- | --- |
| `POST /message/create` | 创建消息 | `{ "user_id": 10001, "message_type": "order", "message_title_one": "订单通知", "message_title_two": "订单已发货", "message_body": "您的订单已发货，请注意查收。", "related_num": "Y2026051500000001", "display_img": "https://example.com/order.png" }` |
| `POST /message/categories` | 消息分类 | `{ "user_id": 10001 }` |
| `POST /message/query` | 消息列表 | `{ "user_id": 10001, "message_type": "order", "page": 1, "page_size": 20 }` |

字段含义：`message_type` 消息类型；`message_title_one/message_title_two` 标题；`message_body` 正文；`related_num` 关联订单/售后单号；`display_img` 展示图。

成功示例：

```json
{"code":200,"msg":"创建成功","data":{"message_id":1}}
```

失败示例：

```json
{"code":201,"msg":"invalid request","data":{},"Err":"message_body为必填字段"}
```

## 5. 全量接口分组总表

### 5.1 小程序端常用接口

| 模块 | 接口 |
| --- | --- |
| 用户/会员 | `/ordinary_user/wechat_login`、`/ordinary_user/send_register_captcha`、`/ordinary_user/bind_wechat_phone`、`/ordinary_user/find_data`、`/ordinary_user/add_data`、`/ordinary_user/Modify_data`、`/ordinary_user/get_user_id`、`/ordinary_user/update_platform_info`、`/ordinary_user/member_amount_summary` |
| 商品/活动 | `/commodity/goods_query_wx`、`/commodity/goods_query`、`/commodity/stylecode_commodities`、`/commodity/search_products_by_name`、`/commodity/batch_get_products_by_ids`、`/commodity/search_commodity_data`、`/commodity/get_all_categories`、`/commodity/get_all_labels`、`/commodity/search_style_codes`、`/commodity/get_commodity_status`、`/activity/query_online_activity_images`、`/activity/get_activity_image_detail` |
| 购物车 | `/cart/add_to_cart`、`/cart/query_cart_items`、`/cart/update_cart_item_quantity`、`/cart/increase_cart_item_quantity`、`/cart/decrease_cart_item_quantity`、`/cart/batch_delete_from_cart`、`/cart/clear_cart` |
| 地址 | `/address/add_address`、`/address/get_addresses`、`/address/get_address_by_id`、`/address/update_address`、`/address/set_default_address`、`/address/delete_address` |
| 订单/售后 | `/order/add_order`、`/order/query_order_data`、`/order/orders_query`、`/order/batch_orders_query`、`/order/query_by_user_id`、`/order/pay`、`/order/cancel`、`/order/order_receive`、`/order/search_by_product_name`、`/order/query_sub_order_data`、`/order/change_sub_order_status`、`/order/cancel_sub_order`、`/order/return_sub_order`、`/order/request_return`、`/return_order/create`、`/return_order/deliver`、`/return_order/receive`、`/return_order/cancel`、`/return_order/update_buyer_info`、`/return_order/query`、`/return_order/detail`、`/return_order/push_jushuitan`、`/return_order/jushuitan_after_sale_push`、`/return_order/jushuitan_after_sale_received_query` |
| 评价/消息 | `/review/create`、`/review/query_by_product`、`/review/statistics`、`/message/categories`、`/message/query` |

### 5.2 后台端常用接口

| 模块 | 接口 |
| --- | --- |
| Token/健康检查 | `/access_token/get_token`、`/access_token/get_ips`、`/api/test/`、`/api/health/` |
| 后台账号 | `/OperationUser/add_service_user`、`/OperationUser/add_operation_user`、`/OperationUser/verification_status`、`/OperationUser/change_password`、`/OperationUser/send_register_captcha`、`/OperationUser/backend_register_by_phone` |
| 商品管理 | `/commodity/add_goods`、`/commodity/delete_goods`、`/commodity/search_commodity_data`、`/commodity/change_commodity_data`、`/commodity/change_commodity_status_online`、`/commodity/change_commodity_status_offline`、`/commodity/stylecode_status_online`、`/commodity/stylecode_status_offline`、`/commodity/update_style_code_info`、以及商品查询辅助接口 |
| 活动管理 | `/activity/add_activity_img`、`/activity/update_activity_image_relations`、`/activity/activity_image_online`、`/activity/activity_image_offline`、`/activity/batch_query_activity_images`、`/activity/batch_update_activity_image_order`、`/activity/add_promotional_pic`、`/activity/update_promotional_pic_order`、`/activity/delete_promotional_pic`、`/activity/set_has_activity_detail` |
| 订单管理 | `/order/orders_query`、`/order/batch_orders_query`、`/order/search_by_product_name`、`/order/change_status`、`/order/update_express_info`、`/order/deliver`、`/order/sync_logistics_info`、`/order/jushuitan_ship_info`、`/order/change_receiving_data`、`/order/update/:id` |
| 售后管理 | `/return_order/approve`、`/return_order/query`、`/return_order/detail`、`/return_order/deliver`、`/return_order/receive`、`/return_order/cancel`、`/return_order/update_buyer_info`、`/return_order/push_jushuitan`、`/return_order/jushuitan_after_sale_push`、`/return_order/jushuitan_after_sale_received_query` |
| 库存 | `/inventory/query`、`/inventory/adjust`、`/inventory/logs`、`/inventory/warnings`、`/inventory/sync_jushuitan` |
| 评价管理 | `/review/query_backend`、`/review/audit`、`/review/reply`、`/review/statistics` |
| 数据分析 | `/analytics/sales_summary`、`/analytics/user_summary`、`/analytics/product_summary`、`/analytics/traffic_summary`、`/analytics/export` |
| 消息 | `/message/create`、`/message/categories`、`/message/query` |

## 6. 对接注意事项

1. 当前多数接口虽然 HTTP 状态可能是 `400/500`，但业务失败体仍常用 `code=201`；前端不要只看 HTTP 状态。
2. 订单创建会扣库存，库存不足时订单创建失败；取消整单或未发货子订单取消会回滚库存。
3. 售后申请后会上传聚水潭；ERP 审核通过只标记售后中，ERP 退货入库/收货回写 `received` 时回滚退货类库存，商城最终同意退款后置为 `completed`。
4. 评价创建后默认 `pending`，只有 `approved` 评价会在 `/review/query_by_product` 中展示。
5. `/inventory/sync_jushuitan` 和 `/analytics/traffic_summary` 是明确预留接口，不应当按真实生产功能依赖。
6. 文件上传接口需要使用 `multipart/form-data`，其余接口优先使用 JSON。
7. 本文档按当前 Gin 注册路由生成，不包含未注册或注释掉的接口。
