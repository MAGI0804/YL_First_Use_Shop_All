package requestbody

type AddGoodsRequestBody struct {
	CommodityID string  `form:"commodity_id" binding:"required"`
	Name        string  `form:"name" binding:"required"`
	Price       float64 `form:"price" binding:"required,gt=0"`
	Category    string  `form:"category" binding:"required"`
	StyleCode   string  `form:"style_code"`
	Size        string  `form:"size"`
	Notes       string  `form:"notes"`
}

type SearchStyleCodesRequestBody struct {
	Shopname      string      `json:"shopname" binding:"required"`
	SearchKeyword string      `json:"search_keyword"`
	Category      interface{} `json:"category"`
	Page          int         `json:"page" binding:"required,min=1"`
	PageSize      int         `json:"page_size" binding:"required,min=1"`
}

type DeleteGoodsRequestBody struct {
	CommodityID interface{} `json:"commodity_id" form:"commodity_id" binding:"required"`
}

type SearchCommodityDataRequestBody struct {
	CommodityID interface{} `json:"commodity_id" form:"commodity_id" binding:"required"`
	DataList    []string    `json:"data_list"`
}

type GoodsQueryRequestBody struct {
	Shopname   string      `json:"shopname" binding:"required"`
	Demand     string      `json:"demand"`
	StyleCode  string      `json:"style_code"`
	Category   interface{} `json:"category"`
	Status     string      `json:"status"`
	LabelOne   []string    `json:"label_one"`
	LabelTwo   []string    `json:"label_two"`
	LabelThree []string    `json:"label_three"`
	LabelFour  []string    `json:"label_four"`
	LabelSeven []string    `json:"label_seven"`
	BeginTime  string      `json:"begin_time"`
	EndTime    string      `json:"end_time"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

type ChangeCommodityDataRequestBody struct {
	CommodityID  interface{}            `json:"commodity_id" form:"commodity_id" binding:"required"`
	UpdateFields map[string]interface{} `json:"update_fields"`
}

type ChangeCommodityStatusRequestBody struct {
	CommodityID interface{} `json:"commodity_id" form:"commodity_id" binding:"required"`
}

type GetCommodityStatusRequestBody struct {
	CommodityID interface{} `json:"commodity_id" form:"commodity_id" binding:"required"`
}

type SearchProductsByNameRequestBody struct {
	SearchStr string `json:"search_str" binding:"required"`
	Page      int    `json:"page" binding:"required,min=1"`
	PageSize  int    `json:"page_size" binding:"required,min=1"`
}

type BatchGetProductsByIDsRequestBody struct {
	CommodityIDs []string `json:"commodity_ids" binding:"required"`
}

type ChangeStyleCodeStatusRequestBody struct {
	StyleCode string `json:"style_code" binding:"required"`
}

type Stylecode_commoditiesstruct struct {
	Shopname  string `json:"shopname" binding:"required"`
	StyleCode string `json:"style_code" binding:"required"`
}

// GetAllLabelsRequestBody 获取所有标签的请求体
type GetAllLabelsRequestBody struct {
	Shopname   string      `json:"shopname" binding:"required"`
	Category   interface{} `json:"category" binding:"omitempty"`
	LabelOne   []string    `json:"label_one" binding:"omitempty"`
	LabelTwo   []string    `json:"label_two" binding:"omitempty"`
	LabelThree []string    `json:"label_three" binding:"omitempty"`
	LabelFour  []string    `json:"label_four" binding:"omitempty"`
	LabelSeven []string    `json:"label_seven" binding:"omitempty"`
}

type UpdateStyleCodeInfoRequestBody struct {
	Shopname   string  `json:"shopname" binding:"required"`
	StyleCode  string  `json:"style_code" binding:"required"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Price      float64 `json:"price"`
	LabelOne   string  `json:"label_one"`
	LabelTwo   string  `json:"label_two"`
	LabelThree string  `json:"label_three"`
	LabelFour  string  `json:"label_four"`
	LabelFive  string  `json:"label_five"`
	LabelSix   string  `json:"label_six"`
	LabelSeven string  `json:"label_seven"`
}
