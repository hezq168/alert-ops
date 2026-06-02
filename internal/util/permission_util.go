package util

import (
	"alert-ops/internal/model"
	dbRepo "alert-ops/internal/repo"
)

func AppendParentPermissions(perms []model.Permission) ([]model.Permission, error) {

	permissionMap := make(map[uint]model.Permission)

	// 已有权限
	for _, p := range perms {
		permissionMap[p.ID] = p
	}

	// 递归补父级
	for _, p := range perms {

		parentID := p.ParentID

		for parentID != 0 {

			// 已存在
			if existing, ok := permissionMap[parentID]; ok {

				// 继续向上找
				parentID = existing.ParentID

				continue
			}

			// 查父节点
			var parent model.Permission

			err := dbRepo.DB.
				Where("id = ? AND status = 1", parentID).
				First(&parent).Error

			if err != nil {
				break
			}

			permissionMap[parent.ID] = parent

			// 继续向上
			parentID = parent.ParentID
		}
	}

	// map 转 slice
	var result []model.Permission

	for _, p := range permissionMap {
		result = append(result, p)
	}

	return result, nil
}
