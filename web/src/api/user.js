import request from '../utils/request'

// 获取用户列表
export function getUsers(params) {
  return request({
    url: '/users',
    method: 'get',
    params
  })
}

// 创建用户（注册）
export function createUser(data) {
  return request({
    url: '/auth/register',
    method: 'post',
    data
  })
}

// 更新用户
export function updateUser(id, data) {
  return request({
    url: `/users/${id}`,
    method: 'put',
    data
  })
}

// 删除用户
export function deleteUser(id) {
  return request({
    url: `/users/${id}`,
    method: 'delete'
  })
}

// 更新用户状态
export function updateUserStatus(id, status) {
  return request({
    url: `/users/${id}/status`,
    method: 'put',
    data: { status }
  })
}

// 获取当前用户信息
export function getCurrentUser() {
  return request({
    url: '/user/info',
    method: 'get'
  })
}

// 修改密码
export function changePassword(data) {
  return request({
    url: '/user/change-password',
    method: 'post',
    data
  })
}


// 获取角色列表
export function getRoles(params) {
  return request({
    url: '/roles',
    method: 'get',
    params
  })
}

// 创建角色
export function createRole(data) {
  return request({
    url: '/roles',
    method: 'post',
    data
  })
}

// 删除角色
export function deleteRole(id) {
  return request({
    url: `/roles/${id}`,
    method: 'delete'
  })
}

// 获取用户角色
export function getUserRoles(userId) {
  return request({
    url: `/users/${userId}/roles`,
    method: 'get'
  })
}

// 分配角色给用户
export function assignRole(userId, roleId) {
  return request({
    url: `/users/${userId}/roles`,
    method: 'post',
    data: { role_id: roleId }
  })
}

// 移除用户角色
export function removeUserRole(userId, roleId) {
  return request({
    url: `/users/${userId}/roles`,
    method: 'delete',
    data: { role_id: roleId }
  })
}