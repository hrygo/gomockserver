package models

import "fmt"

// ErrorCode 错误码结构
type ErrorCode struct {
	Code      int    // 错误码
	Message   string // 错误信息（英文）
	ZhMessage string // 错误信息（中文）
}

// Error 实现 error 接口
func (e ErrorCode) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// ErrorResponse API错误响应结构
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ============================================
// 通用错误 1000-1999
// ============================================

var (
	// ErrInvalidParameter 无效参数
	ErrInvalidParameter = ErrorCode{1001, "Invalid parameter", "参数无效"}
	// ErrMissingParameter 缺少必填参数
	ErrMissingParameter = ErrorCode{1002, "Missing required parameter", "缺少必填参数"}
	// ErrInvalidFormat 格式无效
	ErrInvalidFormat = ErrorCode{1003, "Invalid format", "格式无效"}
	// ErrInvalidObjectID 无效的ObjectID
	ErrInvalidObjectID = ErrorCode{1004, "Invalid ObjectID format", "无效的ObjectID格式"}
	// ErrUnauthorized 未授权
	ErrUnauthorized = ErrorCode{1010, "Unauthorized", "未授权"}
	// ErrForbidden 禁止访问
	ErrForbidden = ErrorCode{1011, "Forbidden", "禁止访问"}
	// ErrRateLimited 请求过于频繁
	ErrRateLimited = ErrorCode{1020, "Too many requests", "请求过于频繁"}
)

// ============================================
// 项目相关 2000-2999
// ============================================

var (
	// ErrProjectNotFound 项目不存在
	ErrProjectNotFound = ErrorCode{2001, "Project not found", "项目不存在"}
	// ErrProjectExists 项目已存在
	ErrProjectExists = ErrorCode{2002, "Project already exists", "项目已存在"}
	// ErrProjectNameInvalid 项目名称无效
	ErrProjectNameInvalid = ErrorCode{2003, "Invalid project name", "项目名称无效"}
	// ErrProjectCreateFailed 创建项目失败
	ErrProjectCreateFailed = ErrorCode{2010, "Failed to create project", "创建项目失败"}
	// ErrProjectUpdateFailed 更新项目失败
	ErrProjectUpdateFailed = ErrorCode{2011, "Failed to update project", "更新项目失败"}
	// ErrProjectDeleteFailed 删除项目失败
	ErrProjectDeleteFailed = ErrorCode{2012, "Failed to delete project", "删除项目失败"}
)

// ============================================
// 环境相关 3000-3999
// ============================================

var (
	// ErrEnvironmentNotFound 环境不存在
	ErrEnvironmentNotFound = ErrorCode{3001, "Environment not found", "环境不存在"}
	// ErrEnvironmentExists 环境已存在
	ErrEnvironmentExists = ErrorCode{3002, "Environment already exists", "环境已存在"}
	// ErrEnvironmentNameInvalid 环境名称无效
	ErrEnvironmentNameInvalid = ErrorCode{3003, "Invalid environment name", "环境名称无效"}
	// ErrEnvironmentCreateFailed 创建环境失败
	ErrEnvironmentCreateFailed = ErrorCode{3010, "Failed to create environment", "创建环境失败"}
	// ErrEnvironmentUpdateFailed 更新环境失败
	ErrEnvironmentUpdateFailed = ErrorCode{3011, "Failed to update environment", "更新环境失败"}
	// ErrEnvironmentDeleteFailed 删除环境失败
	ErrEnvironmentDeleteFailed = ErrorCode{3012, "Failed to delete environment", "删除环境失败"}
)

// ============================================
// 规则相关 4000-4999
// ============================================

var (
	// ErrRuleNotFound 规则不存在
	ErrRuleNotFound = ErrorCode{4001, "Rule not found", "规则不存在"}
	// ErrRuleMatchFailed 未找到匹配的规则
	ErrRuleMatchFailed = ErrorCode{4002, "No matching rule found", "未找到匹配的规则"}
	// ErrRuleNameInvalid 规则名称无效
	ErrRuleNameInvalid = ErrorCode{4003, "Invalid rule name", "规则名称无效"}
	// ErrRulePriorityInvalid 规则优先级无效
	ErrRulePriorityInvalid = ErrorCode{4004, "Invalid rule priority", "规则优先级无效"}
	// ErrRuleConditionInvalid 规则匹配条件无效
	ErrRuleConditionInvalid = ErrorCode{4005, "Invalid rule match condition", "规则匹配条件无效"}
	// ErrRuleResponseInvalid 规则响应配置无效
	ErrRuleResponseInvalid = ErrorCode{4006, "Invalid rule response", "规则响应配置无效"}
	// ErrRuleCreateFailed 创建规则失败
	ErrRuleCreateFailed = ErrorCode{4010, "Failed to create rule", "创建规则失败"}
	// ErrRuleUpdateFailed 更新规则失败
	ErrRuleUpdateFailed = ErrorCode{4011, "Failed to update rule", "更新规则失败"}
	// ErrRuleDeleteFailed 删除规则失败
	ErrRuleDeleteFailed = ErrorCode{4012, "Failed to delete rule", "删除规则失败"}
	// ErrRuleImportFailed 导入规则失败
	ErrRuleImportFailed = ErrorCode{4020, "Failed to import rules", "导入规则失败"}
	// ErrRuleExportFailed 导出规则失败
	ErrRuleExportFailed = ErrorCode{4021, "Failed to export rules", "导出规则失败"}
	// ErrRuleCloneFailed 复制规则失败
	ErrRuleCloneFailed = ErrorCode{4022, "Failed to clone rule", "复制规则失败"}
)

// ============================================
// 批量操作相关 6000-6999
// ============================================

var (
	// ErrBatchOperationFailed 批量操作失败
	ErrBatchOperationFailed = ErrorCode{6001, "Batch operation failed", "批量操作失败"}
	// ErrBatchPartialSuccess 批量操作部分成功
	ErrBatchPartialSuccess = ErrorCode{6002, "Batch operation partially succeeded", "批量操作部分成功"}
	// ErrBatchInvalidInput 批量操作输入无效
	ErrBatchInvalidInput = ErrorCode{6003, "Invalid batch operation input", "批量操作输入无效"}
	// ErrBatchEmptyInput 批量操作输入为空
	ErrBatchEmptyInput = ErrorCode{6004, "Batch operation input is empty", "批量操作输入为空"}
)

// ============================================
// 导入导出相关 7000-7999
// ============================================

var (
	// ErrImportDataInvalid 导入数据无效
	ErrImportDataInvalid = ErrorCode{7001, "Import data is invalid", "导入数据无效"}
	// ErrExportFailed 导出操作失败
	ErrExportFailed = ErrorCode{7002, "Export operation failed", "导出操作失败"}
	// ErrImportConflict 导入数据冲突
	ErrImportConflict = ErrorCode{7003, "Import data conflicts with existing data", "导入数据与现有数据冲突"}
	// ErrUnsupportedVersion 不支持的数据版本
	ErrUnsupportedVersion = ErrorCode{7004, "Unsupported import data version", "不支持的导入数据版本"}
	// ErrImportStrategyInvalid 导入策略无效
	ErrImportStrategyInvalid = ErrorCode{7005, "Invalid import strategy", "导入策略无效"}
	// ErrExportTypeInvalid 导出类型无效
	ErrExportTypeInvalid = ErrorCode{7006, "Invalid export type", "导出类型无效"}
)

// ============================================
// 数据库相关 5000-5999
// ============================================

var (
	// ErrDatabaseConnection 数据库连接失败
	ErrDatabaseConnection = ErrorCode{5001, "Database connection failed", "数据库连接失败"}
	// ErrDatabaseQuery 数据库查询失败
	ErrDatabaseQuery = ErrorCode{5002, "Database query failed", "数据库查询失败"}
	// ErrDatabaseInsert 数据库插入失败
	ErrDatabaseInsert = ErrorCode{5003, "Database insert failed", "数据库插入失败"}
	// ErrDatabaseUpdate 数据库更新失败
	ErrDatabaseUpdate = ErrorCode{5004, "Database update failed", "数据库更新失败"}
	// ErrDatabaseDelete 数据库删除失败
	ErrDatabaseDelete = ErrorCode{5005, "Database delete failed", "数据库删除失败"}
	// ErrDatabaseTransaction 数据库事务失败
	ErrDatabaseTransaction = ErrorCode{5010, "Database transaction failed", "数据库事务失败"}
)

// ============================================
// 系统错误 9000-9999
// ============================================

var (
	// ErrInternalServer 服务器内部错误
	ErrInternalServer = ErrorCode{9001, "Internal server error", "服务器内部错误"}
	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable = ErrorCode{9002, "Service unavailable", "服务不可用"}
	// ErrTimeout 请求超时
	ErrTimeout = ErrorCode{9003, "Request timeout", "请求超时"}
	// ErrConfigError 配置错误
	ErrConfigError = ErrorCode{9010, "Configuration error", "配置错误"}
	// ErrUnknown 未知错误
	ErrUnknown = ErrorCode{9999, "Unknown error", "未知错误"}
)

// NewErrorResponse 创建错误响应
func NewErrorResponse(code ErrorCode, details string, requestID string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:      code.Code,
			Message:   code.Message,
			Details:   details,
			RequestID: requestID,
		},
	}
}

// NewErrorResponseZh 创建中文错误响应
func NewErrorResponseZh(code ErrorCode, details string, requestID string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:      code.Code,
			Message:   code.ZhMessage,
			Details:   details,
			RequestID: requestID,
		},
	}
}
