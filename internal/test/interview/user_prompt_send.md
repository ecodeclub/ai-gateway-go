# 系统自动触发：发送题目给用户

这是系统自动触发的操作，用于将最新获取的题目格式化后发送给用户。

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

