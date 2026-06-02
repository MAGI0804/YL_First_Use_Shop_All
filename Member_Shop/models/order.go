package models

import (
	"time"

	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	OrderID          string    `gorm:"column:order_id;primaryKey;size:20;comment:订单ID/订单号" json:"order_id"`                                                               //订单id
	UserID           int       `gorm:"column:user_id;not null;default:0;comment:用户ID" json:"user_id"`                                                                     //用户ID
	InternalId       int       `gorm:"column:internal_id;not null;default:0;comment:内部ID" json:"internal_id"`                                                             //内部ID
	ReceiverName     string    `gorm:"column:receiver_name;size:100;not null;comment:收货人姓名" json:"receiver_name"`                                                         //收货人姓名
	ReceiverPhone    string    `gorm:"column:receiver_phone;size:15;null;comment:收货人电话" json:"receiver_phone"`                                                            //收货人电话
	ExpressCompany   string    `gorm:"column:express_company;size:50;null;comment:物流公司" json:"express_company"`                                                           //物流公司
	ExpressNumber    string    `gorm:"column:express_number;size:50;null;comment:物流单号" json:"express_number"`                                                             //物流单号
	LogisticsProcess string    `gorm:"column:logistics_process;type:text;null;comment:物流过程/物流信息" json:"logistics_process"`                                                //物流过程/物流信息
	Province         string    `gorm:"column:province;size:50;not null;comment:收货省份" json:"province"`                                                                     //收货省份
	City             string    `gorm:"column:city;size:50;not null;comment:收货城市" json:"city"`                                                                             //收货城市
	County           string    `gorm:"column:county;size:50;not null;comment:收货区县" json:"county"`                                                                         //收货区县
	DetailedAddress  string    `gorm:"column:detailed_address;size:255;not null;comment:详细收货地址" json:"detailed_address"`                                                  //详细收货地址
	OrderAmount      float64   `gorm:"column:order_amount;type:decimal(10,2);not null;comment:订单金额" json:"order_amount"`                                                  //订单金额
	ProdoctNameList  string    `gorm:"column:prodoct_name_list;type:text;not null;comment:商品名称列表" json:"prodoct_name_list"`                                               //商品名称列表
	ProductList      string    `gorm:"column:product_list;type:text;not null;comment:商品列表JSON" json:"product_list"`                                                       //商品列表JSON
	Status           string    `gorm:"column:status;size:20;not null;default:'pending';comment:order lifecycle status: pending/shipped/delivered/canceled" json:"status"` // Lifecycle status. Payment happens after delivered.
	PayStatus        string    `gorm:"column:pay_status;size:20;not null;default:'unpaid';comment:payment status: unpaid/paid" json:"pay_status"`                         // Payment status. It is separate from shipping/signing status.
	OrderTime        time.Time `gorm:"column:order_time;autoCreateTime;comment:下单时间" json:"order_time"`                                                                   //下单时间
	Remarks          string    `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`                                                                           //备注
	PaymentMethod    string    `gorm:"column:payment_method;size:20;null;comment:支付方式" json:"payment_method"`
	DeliveryMethod   string    `gorm:"column:delivery_method;size:20;null;comment:配送方式" json:"delivery_method"`
	PaymentTime      time.Time `gorm:"column:payment_time;null;comment:支付时间" json:"payment_time"`                              //支付时间
	ShippedTime      time.Time `gorm:"column:shipped_time;null;comment:发货时间" json:"shipped_time"`                              //发货时间
	DeliveredTime    time.Time `gorm:"column:delivered_time;null;comment:签收时间" json:"delivered_time"`                          //签收时间
	CanceledTime     time.Time `gorm:"column:canceled_time;null;comment:取消时间" json:"canceled_time"`                            //取消时间
	ProcessingTime   time.Time `gorm:"column:processing_time;null;comment:处理时间" json:"processing_time"`                        //处理时间
	ProcessNum       string    `gorm:"column:process_num;null;comment:处理编号" json:"process_num"`                                //处理编号
	SubOrderIDs      string    `gorm:"column:sub_order_ids;type:text;null;comment:子订单ID列表JSON" json:"sub_order_ids"`           //子订单ID列表JSON
	JushuitanOrderID string    `gorm:"column:jushuitan_order_id;size:50;null;comment:聚水潭系统单号(o_id)" json:"jushuitan_order_id"` //聚水潭系统单号
	OrderFrom        string    `gorm:"column:order_from;size:50;null;comment:订单来源(order_from)" json:"order_from"`              //订单来源
	WmsCoID          string    `gorm:"column:wms_co_id;size:50;null;comment:仓库编码(wms_co_id)" json:"wms_co_id"`                 //仓库编码
	LCID             string    `gorm:"column:lcid;size:50;null;comment:物流公司编码(lc_id)" json:"lcid"`                             //物流公司编码
	IsSendAll        string    `gorm:"column:is_send_all;size:10;default:'no';comment:是否全部发完(is_send_all)" json:"is_send_all"` //是否全部发完(is_send_all)
}

// TableName 设置表名 - 与Django版本保持一致
func (Order) TableName() string {
	return "order_data"
}

// BeforeSave GORM钩子，确保gorm包被使用
func (o *Order) BeforeSave(*gorm.DB) error {
	return nil
}
