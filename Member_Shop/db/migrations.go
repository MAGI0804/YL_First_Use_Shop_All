package db

import (
	"Member_shop/models"
	"fmt"
	"log"
)

// RunMigrations keeps database tables aligned with the GORM models.
func RunMigrations() {
	log.Println("starting database migrations...")

	// User, Member, and BackendUser are three separate business models:
	// User stores WeChat identity, Member stores member data, BackendUser stores staff accounts.
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Member{},
		&models.BackendUser{},
		//&models.UserData{},
		&models.Address{},
		&models.Commodity{},
		&models.CommodityImage{},
		&models.CommoditySituation{},
		&models.StyleCodeSituation{},
		&models.StyleCodeData{},
		&models.Order{},
		&models.ReturnOrder{},
		&models.Cart{},
		&models.ActivityImage{},
		&models.Product{},
		&models.AccessToken{},
		&models.DjangoCustomerServiceUser{},
		&models.DjangoOperationUser{},
		&models.Message{},
		&models.SubOrder{},
		&models.InventoryLog{},
		&models.ProductReview{},
		&models.ReviewReply{},
		&models.JushuitanPushRawData{},
		&models.BackendUser{},
		&models.Member{},
	}

	for _, model := range modelsToMigrate {
		modelName := fmt.Sprintf("%T", model)
		if err := DB.AutoMigrate(model); err != nil {
			log.Printf("migrate %v failed: %v", modelName, err)
		} else {
			log.Printf("migrate %v succeeded", modelName)
		}
	}

	log.Println("database migrations completed")
}
