package service

import (
	"__APP_NAME__/internal/model"
)

// Example 的 CRUD 方法直接挂在 Service 聚合上（数据访问经 s.db）。
// 绑定方法（app.go）不直接碰 db，业务逻辑全在这一层。

// ListExamples 查询全部，返回列表。
// TODO: 业务逻辑 —— 按需加分页、keyword 过滤、排序等。
// 注意：Wails 绑定只把多返回值的第一个暴露给前端，分页总数等需封装成结构体。
func (s *Service) ListExamples() ([]model.Example, error) {
	var list []model.Example
	if err := s.db.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetExample 按 ID 查询单个。
func (s *Service) GetExample(id uint) (*model.Example, error) {
	var item model.Example
	if err := s.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// CreateExample 新增。
func (s *Service) CreateExample(item *model.Example) error {
	return s.db.Create(item).Error
}

// UpdateExample 更新（全字段覆盖）。
func (s *Service) UpdateExample(item *model.Example) error {
	return s.db.Save(item).Error
}

// DeleteExample 删除（软删除）。
func (s *Service) DeleteExample(id uint) error {
	return s.db.Delete(&model.Example{}, id).Error
}
