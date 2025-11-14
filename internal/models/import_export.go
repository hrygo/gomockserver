package models

import "time"

// ExportType 导出类型
type ExportType string

const (
	ExportTypeRules       ExportType = "rules"       // 仅规则
	ExportTypeEnvironment ExportType = "environment" // 环境及规则
	ExportTypeProject     ExportType = "project"     // 项目、环境及规则
)

// ImportStrategy 导入策略
type ImportStrategy string

const (
	ImportStrategySkip      ImportStrategy = "skip"      // 跳过冲突
	ImportStrategyOverwrite ImportStrategy = "overwrite" // 覆盖冲突
	ImportStrategyAppend    ImportStrategy = "append"    // 追加新增（自动重命名）
)

// ExportData 导出数据结构
type ExportData struct {
	Version    string               `json:"version"`                 // 数据格式版本
	ExportTime time.Time            `json:"export_time"`             // 导出时间
	ExportType ExportType           `json:"export_type"`             // 导出类型
	Data       ExportDataContent    `json:"data"`                    // 导出内容
	Metadata   *ExportMetadata      `json:"metadata,omitempty"`      // 元数据（可选）
}

// ExportDataContent 导出内容
type ExportDataContent struct {
	Project      *ProjectExportData      `json:"project,omitempty"`      // 项目信息
	Environments []EnvironmentExportData `json:"environments,omitempty"` // 环境列表
	Rules        []RuleExportData        `json:"rules"`                  // 规则列表
}

// ProjectExportData 项目导出数据
type ProjectExportData struct {
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id"`
	Description string `json:"description,omitempty"`
}

// EnvironmentExportData 环境导出数据
type EnvironmentExportData struct {
	Name      string                 `json:"name"`
	BaseURL   string                 `json:"base_url,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// RuleExportData 规则导出数据
type RuleExportData struct {
	Name           string                 `json:"name"`
	EnvironmentID  string                 `json:"environment_id,omitempty"`  // 导出时可能包含环境ID引用
	EnvironmentName string                `json:"environment_name,omitempty"` // 导出时包含环境名称以便导入
	Protocol       ProtocolType           `json:"protocol"`
	MatchType      MatchType              `json:"match_type"`
	Priority       int                    `json:"priority"`
	Enabled        bool                   `json:"enabled"`
	MatchCondition map[string]interface{} `json:"match_condition"`
	Response       Response               `json:"response"`
	Tags           []string               `json:"tags,omitempty"`
	Description    string                 `json:"description,omitempty"`
}

// ExportMetadata 导出元数据
type ExportMetadata struct {
	ExportedBy string `json:"exported_by,omitempty"` // 导出者
	Comment    string `json:"comment,omitempty"`     // 备注说明
}

// ExportRequest 导出请求
type ExportRequest struct {
	ProjectID      string   `json:"project_id,omitempty"`      // 项目ID
	EnvironmentID  string   `json:"environment_id,omitempty"`  // 环境ID
	RuleIDs        []string `json:"rule_ids,omitempty"`        // 规则ID列表
	IncludeProject bool     `json:"include_project"`           // 是否包含项目信息
	IncludeEnvs    bool     `json:"include_environments"`      // 是否包含环境信息
	IncludeMetadata bool    `json:"include_metadata"`          // 是否包含元数据
}

// ImportRequest 导入请求
type ImportRequest struct {
	Data           ExportData     `json:"data"`                      // 导入数据
	TargetProjectID string        `json:"target_project_id,omitempty"` // 目标项目ID（可选，为空则创建新项目）
	TargetEnvID     string        `json:"target_environment_id,omitempty"` // 目标环境ID（可选）
	Strategy       ImportStrategy `json:"strategy"`                  // 导入策略
	CreateProject  bool           `json:"create_project"`            // 是否创建新项目
	CreateEnvs     bool           `json:"create_environments"`       // 是否创建新环境
}

// ImportResult 导入结果
type ImportResult struct {
	Success        bool              `json:"success"`
	ProjectID      string            `json:"project_id,omitempty"`
	EnvironmentIDs map[string]string `json:"environment_ids,omitempty"` // 环境名称 -> ID映射
	RuleIDs        []string          `json:"rule_ids,omitempty"`
	Skipped        int               `json:"skipped"`   // 跳过的规则数
	Created        int               `json:"created"`   // 新建的规则数
	Updated        int               `json:"updated"`   // 更新的规则数
	Errors         []ImportError     `json:"errors,omitempty"`
}

// ImportError 导入错误
type ImportError struct {
	RuleName string `json:"rule_name"`
	Error    string `json:"error"`
}

// CloneRuleRequest 克隆规则请求
type CloneRuleRequest struct {
	TargetProjectID     string `json:"target_project_id,omitempty"`     // 目标项目ID（可选）
	TargetEnvironmentID string `json:"target_environment_id"`           // 目标环境ID（必填）
	NewName             string `json:"new_name,omitempty"`              // 新规则名称（可选）
	NewPriority         *int   `json:"new_priority,omitempty"`          // 新优先级（可选）
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	RuleIDs   []string               `json:"rule_ids"`             // 规则ID列表
	Operation string                 `json:"operation"`            // 操作类型：enable, disable, delete, update
	Updates   map[string]interface{} `json:"updates,omitempty"`    // 更新字段（用于update操作）
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	Success      bool     `json:"success"`
	TotalCount   int      `json:"total_count"`   // 总数
	SuccessCount int      `json:"success_count"` // 成功数
	FailedCount  int      `json:"failed_count"`  // 失败数
	FailedIDs    []string `json:"failed_ids,omitempty"` // 失败的ID列表
	Errors       []string `json:"errors,omitempty"`
}
