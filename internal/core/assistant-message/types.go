package assistantmessage

// ContentType represents the type of content in an assistant message
type ContentType string

const (
	// TextContentType represents text content
	TextContentType ContentType = "text"
	// ToolUseContentType represents a tool use
	ToolUseContentType ContentType = "tool_use"
)

// Content represents the content of an assistant message
type Content struct {
	Type    ContentType `json:"type"`
	Content string      `json:"content,omitempty"`
	Partial bool        `json:"partial"`
}

// TextContent represents text content in an assistant message
type TextContent struct {
	Content
}

// NewTextContent creates a new text content
func NewTextContent(content string, partial bool) TextContent {
	return TextContent{
		Content: Content{
			Type:    TextContentType,
			Content: content,
			Partial: partial,
		},
	}
}

// ToolUseName represents the name of a tool
type ToolUseName string

// ToolParamName represents the name of a tool parameter
type ToolParamName string

// Tool use names
const (
	ExecuteCommandToolName          ToolUseName = "execute_command"
	ReadFileToolName                ToolUseName = "read_file"
	WriteToFileToolName             ToolUseName = "write_to_file"
	ReplaceInFileToolName           ToolUseName = "replace_in_file"
	SearchFilesToolName             ToolUseName = "search_files"
	ListFilesToolName               ToolUseName = "list_files"
	ListCodeDefinitionNamesToolName ToolUseName = "list_code_definition_names"
	BrowserActionToolName           ToolUseName = "browser_action"
	UseMcpToolToolName              ToolUseName = "use_mcp_tool"
	AccessMcpResourceToolName       ToolUseName = "access_mcp_resource"
	AskFollowupQuestionToolName     ToolUseName = "ask_followup_question"
	PlanModeResponseToolName        ToolUseName = "plan_mode_response"
	AttemptCompletionToolName       ToolUseName = "attempt_completion"
)

// Tool parameter names
const (
	CommandParam          ToolParamName = "command"
	RequiresApprovalParam ToolParamName = "requires_approval"
	PathParam             ToolParamName = "path"
	ContentParam          ToolParamName = "content"
	DiffParam             ToolParamName = "diff"
	RegexParam            ToolParamName = "regex"
	FilePatternParam      ToolParamName = "file_pattern"
	RecursiveParam        ToolParamName = "recursive"
	ActionParam           ToolParamName = "action"
	URLParam              ToolParamName = "url"
	CoordinateParam       ToolParamName = "coordinate"
	TextParam             ToolParamName = "text"
	ServerNameParam       ToolParamName = "server_name"
	ToolNameParam         ToolParamName = "tool_name"
	ArgumentsParam        ToolParamName = "arguments"
	URIParam              ToolParamName = "uri"
	QuestionParam         ToolParamName = "question"
	ResponseParam         ToolParamName = "response"
	ResultParam           ToolParamName = "result"
)

// ToolUse represents a tool use in an assistant message
type ToolUse struct {
	Content
	Name   ToolUseName              `json:"name"`
	Params map[ToolParamName]string `json:"params"`
}

// NewToolUse creates a new tool use
func NewToolUse(name ToolUseName, partial bool) ToolUse {
	return ToolUse{
		Content: Content{
			Type:    ToolUseContentType,
			Partial: partial,
		},
		Name:   name,
		Params: make(map[ToolParamName]string),
	}
}

// AllToolUseNames returns all tool use names
func AllToolUseNames() []ToolUseName {
	return []ToolUseName{
		ExecuteCommandToolName,
		ReadFileToolName,
		WriteToFileToolName,
		ReplaceInFileToolName,
		SearchFilesToolName,
		ListFilesToolName,
		ListCodeDefinitionNamesToolName,
		BrowserActionToolName,
		UseMcpToolToolName,
		AccessMcpResourceToolName,
		AskFollowupQuestionToolName,
		PlanModeResponseToolName,
		AttemptCompletionToolName,
	}
}

// AllToolParamNames returns all tool parameter names
func AllToolParamNames() []ToolParamName {
	return []ToolParamName{
		CommandParam,
		RequiresApprovalParam,
		PathParam,
		ContentParam,
		DiffParam,
		RegexParam,
		FilePatternParam,
		RecursiveParam,
		ActionParam,
		URLParam,
		CoordinateParam,
		TextParam,
		ServerNameParam,
		ToolNameParam,
		ArgumentsParam,
		URIParam,
		QuestionParam,
		ResponseParam,
		ResultParam,
	}
}
