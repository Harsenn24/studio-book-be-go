package address

import (
	"context"
	province "studio-book-be-go/internal/repository/province"

	"gorm.io/gorm"
)

type AddressService interface {
	ProvincePagination(ctx context.Context, input ListProvinceRequest) (ListProvinceResponse, error)
}

type addressServiceImpl struct {
	provinceRepo province.ProvinceRepository
	db *gorm.DB
}

func NewAddressService(
	provinceRepo province.ProvinceRepository,
	db *gorm.DB,
) AddressService {
	return &addressServiceImpl{
		provinceRepo: provinceRepo,
		db: db,
	}
}

func (r *addressServiceImpl) ProvincePagination(ctx context.Context, input ListProvinceRequest) (ListProvinceResponse, error) {
	return ListProvinceResponse{}, nil
}
