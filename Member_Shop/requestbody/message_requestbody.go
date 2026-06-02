package requestbody

// MessageCategoryRequest is the request for querying message categories
type MessageCategoryRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

// MessageQueryRequest is the request for querying messages by category and user ID
type MessageQueryRequest struct {
	UserID      int    `json:"user_id" binding:"required"`
	MessageType string `json:"message_type" binding:"required"`
	Page        int    `json:"page" binding:"required,min=1"`
	PageSize    int    `json:"page_size" binding:"required,min=1,max=50"`
}

// MessageCreateRequest 自定义添加消息请求体
type MessageCreateRequest struct {
	UserID         int    `json:"user_id" binding:"required"`
	MessageType    string `json:"message_type" binding:"required"`
	MessageTitleOne string `json:"message_title_one" binding:"required"`
	MessageTitleTwo string `json:"message_title_two" binding:"required"`
	MessageBody     string `json:"message_body" binding:"required"`
	RelatedNum      string `json:"related_num"`
	DisplayImg      string `json:"display_img"`
}
