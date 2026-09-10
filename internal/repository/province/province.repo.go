package ProvinceRepository

import "gorm.io/gorm"

type ProvinceRepository interface {
	Pagination(page int, limit int, search string) ([]Province, int64, error)
}

type provinceRepositoryImpl struct {
	db *gorm.DB
}

func NewProvinceRepository(db *gorm.DB) ProvinceRepository {
	return &provinceRepositoryImpl{
		db: db,
	}
}

func (r *provinceRepositoryImpl) Pagination(page int, limit int, search string) ([]Province, int64, error) {
	var provinces []Province
	var totalData int64

	query := r.db.Model(&Province{})

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%") 
	}

	if err := query.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Find(&provinces).Error
	if err != nil {
		return nil, 0, err
	}
	
	return provinces, totalData, nil
}
