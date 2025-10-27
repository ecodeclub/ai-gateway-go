# 用户输入
{{.Input}}

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
