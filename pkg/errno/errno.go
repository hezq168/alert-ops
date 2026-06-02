package errno

// 通用错误码
const (
	Success             = 200
	BadRequest          = 400
	Unauthorized        = 401
	Forbidden           = 403
	NotFound            = 404
	InternalServerError = 500
)

// 业务错误码 (1000-1999: 用户相关, 2000-2999: 角色相关, 3000-3999: K8S相关)
const (
	// 用户相关
	ErrUserNotFound    = 1001
	ErrUserExists      = 1002
	ErrInvalidPassword = 1003
	ErrUserDisabled    = 1004

	// 角色相关
	ErrRoleNotFound = 2001
	ErrRoleExists   = 2002
	ErrRoleAssigned = 2003

	// K8S相关
	ErrClusterNotFound = 3001
	ErrClusterConnect  = 3002
)

var errMsg = map[int]string{
	Success:             "success",
	BadRequest:          "请求参数错误",
	Unauthorized:        "未授权",
	Forbidden:           "禁止访问",
	NotFound:            "资源不存在",
	InternalServerError: "服务器内部错误",

	ErrUserNotFound:    "用户不存在",
	ErrUserExists:      "用户已存在",
	ErrInvalidPassword: "密码错误",
	ErrUserDisabled:    "账号已被禁用",

	ErrRoleNotFound: "角色不存在",
	ErrRoleExists:   "角色已存在",
	ErrRoleAssigned: "该用户已拥有此角色",

	ErrClusterNotFound: "集群不存在",
	ErrClusterConnect:  "集群连接失败",
}

func GetMessage(code int) string {
	if msg, ok := errMsg[code]; ok {
		return msg
	}
	return "未知错误"
}
