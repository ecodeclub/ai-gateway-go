# 用户输入
{{.Input}}

# 可用的变量数据（Function Call 返回的结果）
{{range $key, $value := .}}
{{if and (ne $key "Input") (ne $key "AskedQuestionIDs") (ne $key "InterviewHistory")}}
**变量 {{$key}}**:
{{$value}}
{{end}}
{{end}}

# 当前面试历史（用于提取已问题目ID）
{{if .InterviewHistory}}
历史记录：
{{.InterviewHistory}}

**重要**：从上面的历史记录中提取所有 `question_id`，用于排除已问过的题目。
{{else}}
（无历史记录，说明是首题）
{{end}}

# 已问题目ID列表（仅供参考，可能未更新）
{{if .AskedQuestionIDs}}
已问过的题目ID: [{{.AskedQuestionIDs}}]
{{else}}
已问过的题目ID: []（首题）
{{end}}

