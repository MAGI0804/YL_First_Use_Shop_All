package requestbody

// InventoryQueryRequest 库存查询请求结构体
// 用于根据商品ID或款式编码查询库存信息
type InventoryQueryRequest struct {
	CommodityID string `json:"commodity_id" form:"commodity_id"` // 商品ID，精确查询
	StyleCode   string `json:"style_code" form:"style_code"`     // 款式编码，可查询该款式下所有商品的库存汇总
}

// InventoryAdjustRequest 库存调整请求结构体
// 用于手动调整商品库存数量，可增加或减少库存
type InventoryAdjustRequest struct {
	CommodityID   string `json:"commodity_id" binding:"required"`  // 商品ID，必填
	ChangeQty     int    `json:"change_qty" binding:"required"`    // 库存变动数量，正数增加，负数减少
	OperatorID    string `json:"operator_id"`                       // 操作人ID
	Remark        string `json:"remark"`                            // 备注说明
	WarehouseCode string `json:"warehouse_code"`                    // 仓库编码
}

// InventoryLogsRequest 库存变动日志查询请求结构体
// 用于查询库存变动的历史记录，支持多种筛选条件
type InventoryLogsRequest struct {
	CommodityID       string `json:"commodity_id"`        // 商品ID
	StyleCode         string `json:"style_code"`          // 款式编码
	ChangeType        string `json:"change_type"`         // 变动类型：order_deduct-订单扣减, order_cancel_restore-订单取消恢复, return_restore-售后恢复, manual_adjust-手动调整
	RelatedOrderID    string `json:"related_order_id"`    // 关联订单ID
	RelatedSubOrderID string `json:"related_sub_order_id"` // 关联子订单ID
	RelatedReturnID   string `json:"related_return_id"`    // 关联退货/售后ID
	Page              int    `json:"page"`                 // 页码，默认1
	PageSize          int    `json:"page_size"`            // 每页数量，默认20，最大100
}

// InventoryWarningsRequest 库存预警查询请求结构体
// 用于查询库存低于阈值的商品列表
type InventoryWarningsRequest struct {
	Threshold int `json:"threshold"` // 库存预警阈值，默认5
	Page      int `json:"page"`      // 页码，默认1
	PageSize  int `json:"page_size"`  // 每页数量，默认20
}

// InventorySyncJushuitanRequest 库存同步聚水潭请求结构体
// 用于将指定商品的库存同步到聚水潭系统
type InventorySyncJushuitanRequest struct {
	CommodityIDs []string `json:"commodity_ids"` // 要同步的商品ID列表
}
