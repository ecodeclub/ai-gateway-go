# 用户输入
{{.Input}}

# 可用的变量数据（Function Call 返回的结果）
{{range $key, $value := .}}
{{if and (ne $key "Input") (ne $key "AskedQuestionIDs") (ne $key "InterviewHistory")}}
**变量 {{$key}}**:
{{$value}}
{{end}}
{{end}}

# 已问题目ID列表（用于kbase_rag排除）
{{if .AskedQuestionIDs}}
已问过的题目ID: [{{.AskedQuestionIDs}}]
{{else}}
已问过的题目ID: []（首题）
{{end}}

# 当前面试历史（供生成总结时参考）
{{if .InterviewHistory}}
历史记录：
{{.InterviewHistory}}
{{else}}
（无历史记录）
{{end}}

