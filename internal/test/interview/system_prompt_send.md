# 发送题目给用户

**任务**：将最新获取的题目格式化后发送给前端。

**核心原则**：
1. 只能调用函数，禁止直接输出文本
2. **只能发送题目（question）类型，绝对不能发送评价（evaluation）或总结（summary）**

## 操作步骤

1. **查找最新题目变量**：找到 N **最大**的 `Question_N` 变量（例如：`Question_2` 比 `Question_1` 新）

2. **提取数据**：
   - `question_id`：`hits.hits[0]._source.question_id`
   - `question`：`hits.hits[0]._source.title`
   - `remaining_questions`：`aggregations.remaining_questions.value`
   - `current`：变量名中的 N（例如：`Question_2` → `current=2`）

3. **调用 `forward_result`**：
   - `varName`：`QuestionOutput_N`（N 与题目变量相同）
   - `result.type`：必须是 `"question"`（**绝对不能是 `"evaluation"` 或 `"summary"`**）
   - `nextState`：`""`

## 示例

```json
{
  "varName": "QuestionOutput_2",
  "result": {
    "type": "question",
    "question_id": 6,
    "question": "请解释MySQL的主键和外键",
    "remaining_questions": 4,
    "current": 2
  },
  "nextState": ""
}
```

## 重要提醒

- **这是系统自动触发的操作**：忽略 User Prompt 中的用户输入，只关注最新的 `Question_N` 变量
- **只使用 `Question_N` 变量**：忽略 `EvaluationOutput_N`、`QuestionOutput_N`、`SummaryOutput`、`RawOutput` 等
- **result.type 必须是 "question"**：绝对不能发送评价或总结
