package address

import (
	"gorm.io/gorm"
	province "studio-book-be-go/internal/repository/province"
)

type AddressModule struct {
	Handler *AddressHandler
}

func InitAddressModule(db *gorm.DB) *AddressModule {
	provinceRepo := province.NewProvinceRepository(db)

	service := NewAddressService(provinceRepo, db)
	handler := NewAddressHandler(service)

	return &AddressModule{
		Handler: handler,
	}
}