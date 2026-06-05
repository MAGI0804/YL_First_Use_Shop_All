package db

import (
	"Member_shop/models"
	"log"
)

func seedDefaultDownloadTemplates() {
	templates := []models.DownloadTemplate{
		{
			TemplateCode:   "order_export",
			TemplateName:   "订单导出",
			BusinessType:   "order",
			SQLContent:     "SELECT order_id, user_id, receiver_name, receiver_phone, province, city, county, detailed_address, order_amount, final_pay_amount, discount_amount, status, pay_status, express_company, express_number, order_time, payment_time, shipped_time, delivered_time, jushuitan_order_id FROM order_data",
			ModelFields:    `[{"field":"order_id","model":"Order","model_field":"OrderID","db_column":"order_id","type":"string"},{"field":"user_id","model":"Order","model_field":"UserID","db_column":"user_id","type":"int"},{"field":"receiver_name","model":"Order","model_field":"ReceiverName","db_column":"receiver_name","type":"string"},{"field":"receiver_phone","model":"Order","model_field":"ReceiverPhone","db_column":"receiver_phone","type":"string"},{"field":"final_pay_amount","model":"Order","model_field":"FinalPayAmount","db_column":"final_pay_amount","type":"decimal"},{"field":"status","model":"Order","model_field":"Status","db_column":"status","type":"string"},{"field":"pay_status","model":"Order","model_field":"PayStatus","db_column":"pay_status","type":"string"},{"field":"order_time","model":"Order","model_field":"OrderTime","db_column":"order_time","type":"datetime"}]`,
			ExportHeaders:  `[{"field":"order_id","header":"订单号","width":24},{"field":"user_id","header":"用户ID","width":12},{"field":"receiver_name","header":"收货人","width":16},{"field":"receiver_phone","header":"收货电话","width":18},{"field":"province","header":"省份","width":12},{"field":"city","header":"城市","width":12},{"field":"county","header":"区县","width":12},{"field":"detailed_address","header":"详细地址","width":32},{"field":"order_amount","header":"订单金额","width":14,"format":"money"},{"field":"final_pay_amount","header":"实付金额","width":14,"format":"money"},{"field":"discount_amount","header":"优惠金额","width":14,"format":"money"},{"field":"status","header":"订单状态","width":14},{"field":"pay_status","header":"支付状态","width":14},{"field":"express_company","header":"物流公司","width":16},{"field":"express_number","header":"物流单号","width":22},{"field":"order_time","header":"下单时间","width":20,"format":"datetime"},{"field":"payment_time","header":"支付时间","width":20,"format":"datetime"},{"field":"shipped_time","header":"发货时间","width":20,"format":"datetime"},{"field":"delivered_time","header":"签收时间","width":20,"format":"datetime"},{"field":"jushuitan_order_id","header":"聚水潭订单号","width":22}]`,
			AllowedFilters: `[{"field":"begin_time","operator":">=","db_column":"order_time","type":"datetime"},{"field":"end_time","operator":"<=","db_column":"order_time","type":"datetime"},{"field":"status","operator":"=","db_column":"status","type":"string"},{"field":"pay_status","operator":"=","db_column":"pay_status","type":"string"},{"field":"order_from","operator":"=","db_column":"order_from","type":"string"},{"field":"order_id","operator":"=","db_column":"order_id","type":"string"},{"field":"receiver_phone","operator":"=","db_column":"receiver_phone","type":"string"}]`,
			DefaultOrderBy: "order_time DESC",
			FileFormat:     "xlsx",
			Status:         "enabled",
		},
		{
			TemplateCode:   "product_export",
			TemplateName:   "商品导出",
			BusinessType:   "product",
			SQLContent:     "SELECT commodity_id, name, style_code, category, category_detail, price, size, color, height, spec_code, inventory, created_at, notes FROM Commodity_data",
			ModelFields:    `[{"field":"commodity_id","model":"Commodity","model_field":"CommodityID","db_column":"commodity_id","type":"string"},{"field":"name","model":"Commodity","model_field":"Name","db_column":"name","type":"string"},{"field":"style_code","model":"Commodity","model_field":"StyleCode","db_column":"style_code","type":"string"},{"field":"category","model":"Commodity","model_field":"Category","db_column":"category","type":"string"},{"field":"inventory","model":"Commodity","model_field":"Inventory","db_column":"inventory","type":"int"}]`,
			ExportHeaders:  `[{"field":"commodity_id","header":"商品ID","width":24},{"field":"name","header":"商品名称","width":28},{"field":"style_code","header":"款号","width":16},{"field":"category","header":"分类","width":16},{"field":"category_detail","header":"详细分类","width":16},{"field":"price","header":"价格","width":12,"format":"money"},{"field":"size","header":"尺码","width":12},{"field":"color","header":"颜色","width":12},{"field":"height","header":"身高","width":12},{"field":"spec_code","header":"规格码","width":18},{"field":"inventory","header":"库存","width":12},{"field":"created_at","header":"创建时间","width":20,"format":"datetime"},{"field":"notes","header":"备注","width":24}]`,
			AllowedFilters: `[{"field":"category","operator":"=","db_column":"category","type":"string"},{"field":"category_detail","operator":"=","db_column":"category_detail","type":"string"},{"field":"style_code","operator":"=","db_column":"style_code","type":"string"},{"field":"commodity_id","operator":"=","db_column":"commodity_id","type":"string"},{"field":"name","operator":"LIKE","db_column":"name","type":"string"},{"field":"low_inventory","operator":"<=","db_column":"inventory","type":"int"},{"field":"begin_time","operator":">=","db_column":"created_at","type":"datetime"},{"field":"end_time","operator":"<=","db_column":"created_at","type":"datetime"}]`,
			DefaultOrderBy: "created_at DESC",
			FileFormat:     "xlsx",
			Status:         "enabled",
		},
		{
			TemplateCode:   "analytics_sales_export",
			TemplateName:   "报表导出",
			BusinessType:   "report",
			SQLContent:     "SELECT order_id, order_time, order_amount, final_pay_amount, discount_amount, status, pay_status FROM order_data",
			ModelFields:    `[{"field":"order_id","model":"Order","model_field":"OrderID","db_column":"order_id","type":"string"},{"field":"order_time","model":"Order","model_field":"OrderTime","db_column":"order_time","type":"datetime"},{"field":"final_pay_amount","model":"Order","model_field":"FinalPayAmount","db_column":"final_pay_amount","type":"decimal"}]`,
			ExportHeaders:  `[{"field":"date","header":"日期","width":14},{"field":"order_count","header":"订单数","width":12},{"field":"paid_order_count","header":"已支付订单","width":14},{"field":"sales_amount","header":"销售额","width":14,"format":"money"},{"field":"discount_amount","header":"优惠金额","width":14,"format":"money"},{"field":"refund_amount","header":"退款金额","width":14,"format":"money"},{"field":"average_order_value","header":"客单价","width":14,"format":"money"}]`,
			AllowedFilters: `[{"field":"begin_time","operator":">=","db_column":"order_time","type":"datetime"},{"field":"end_time","operator":"<=","db_column":"order_time","type":"datetime"},{"field":"category","operator":"=","db_column":"category","type":"string"},{"field":"style_code","operator":"=","db_column":"style_code","type":"string"},{"field":"low_inventory_threshold","operator":"<=","db_column":"inventory","type":"int"},{"field":"slow_sales_threshold","operator":"<=","db_column":"sales_qty","type":"int"},{"field":"limit","operator":"LIMIT","db_column":"limit","type":"int"}]`,
			DefaultOrderBy: "order_time DESC",
			FileFormat:     "xlsx",
			Status:         "enabled",
		},
		{
			TemplateCode:   "inventory_export",
			TemplateName:   "库存导出",
			BusinessType:   "inventory",
			SQLContent:     "SELECT commodity_id, name, style_code, category, size, color, inventory, created_at FROM Commodity_data",
			ModelFields:    `[{"field":"commodity_id","model":"Commodity","model_field":"CommodityID","db_column":"commodity_id","type":"string"},{"field":"inventory","model":"Commodity","model_field":"Inventory","db_column":"inventory","type":"int"}]`,
			ExportHeaders:  `[{"field":"commodity_id","header":"商品ID","width":24},{"field":"name","header":"商品名称","width":28},{"field":"style_code","header":"款号","width":16},{"field":"category","header":"分类","width":16},{"field":"size","header":"尺码","width":12},{"field":"color","header":"颜色","width":12},{"field":"inventory","header":"当前库存","width":12},{"field":"created_at","header":"创建时间","width":20,"format":"datetime"}]`,
			AllowedFilters: `[{"field":"category","operator":"=","db_column":"category","type":"string"},{"field":"style_code","operator":"=","db_column":"style_code","type":"string"},{"field":"commodity_id","operator":"=","db_column":"commodity_id","type":"string"},{"field":"low_inventory_threshold","operator":"<=","db_column":"inventory","type":"int"},{"field":"begin_time","operator":">=","db_column":"created_at","type":"datetime"},{"field":"end_time","operator":"<=","db_column":"created_at","type":"datetime"}]`,
			DefaultOrderBy: "created_at DESC",
			FileFormat:     "xlsx",
			Status:         "enabled",
		},
		{
			TemplateCode:   "after_sale_export",
			TemplateName:   "售后导出",
			BusinessType:   "after_sale",
			SQLContent:     "SELECT return_id, order_id, sub_order_id, type, status, user_id, sub_order_product_info, product_list, express_company, express_number, request_time, shipped_time, completed_time, jushuitan_after_sale_id, jushuitan_push_status FROM return_order_data",
			ModelFields:    `[{"field":"return_id","model":"ReturnOrder","model_field":"ReturnID","db_column":"return_id","type":"string"},{"field":"order_id","model":"ReturnOrder","model_field":"OrderID","db_column":"order_id","type":"string"},{"field":"status","model":"ReturnOrder","model_field":"Status","db_column":"status","type":"string"},{"field":"request_time","model":"ReturnOrder","model_field":"RequestTime","db_column":"request_time","type":"datetime"}]`,
			ExportHeaders:  `[{"field":"return_id","header":"售后单号","width":24},{"field":"order_id","header":"订单号","width":24},{"field":"sub_order_id","header":"子订单号","width":24},{"field":"type","header":"售后类型","width":14},{"field":"status","header":"售后状态","width":14},{"field":"user_id","header":"用户ID","width":12},{"field":"sub_order_product_info","header":"子订单商品信息","width":32},{"field":"product_list","header":"商品列表","width":32},{"field":"express_company","header":"退货物流公司","width":16},{"field":"express_number","header":"退货物流单号","width":22},{"field":"request_time","header":"申请时间","width":20,"format":"datetime"},{"field":"shipped_time","header":"发货时间","width":20,"format":"datetime"},{"field":"completed_time","header":"完成时间","width":20,"format":"datetime"},{"field":"jushuitan_after_sale_id","header":"聚水潭售后单号","width":22},{"field":"jushuitan_push_status","header":"聚水潭推送状态","width":18}]`,
			AllowedFilters: `[{"field":"begin_time","operator":">=","db_column":"request_time","type":"datetime"},{"field":"end_time","operator":"<=","db_column":"request_time","type":"datetime"},{"field":"status","operator":"=","db_column":"status","type":"string"},{"field":"type","operator":"=","db_column":"type","type":"string"},{"field":"order_id","operator":"=","db_column":"order_id","type":"string"},{"field":"return_id","operator":"=","db_column":"return_id","type":"string"}]`,
			DefaultOrderBy: "request_time DESC",
			FileFormat:     "xlsx",
			Status:         "enabled",
		},
	}

	for _, template := range templates {
		var existing models.DownloadTemplate
		err := DB.Where("template_code = ?", template.TemplateCode).First(&existing).Error
		if err == nil {
			if err := DB.Model(&existing).Updates(map[string]any{
				"template_name":    template.TemplateName,
				"business_type":    template.BusinessType,
				"sql_content":      template.SQLContent,
				"model_fields":     template.ModelFields,
				"export_headers":   template.ExportHeaders,
				"allowed_filters":  template.AllowedFilters,
				"default_order_by": template.DefaultOrderBy,
				"file_format":      template.FileFormat,
			}).Error; err != nil {
				log.Printf("update download template %s failed: %v", template.TemplateCode, err)
			}
			continue
		}
		if err := DB.Create(&template).Error; err != nil {
			log.Printf("seed download template %s failed: %v", template.TemplateCode, err)
		}
	}
}
