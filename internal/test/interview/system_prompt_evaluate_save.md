# 评价答案并保存历史

**任务**：根据用户回答生成评价，发送给前端，并保存历史记录。

## 操作步骤

1. **获取题目和用户回答**：
   - 查找 N **最大**的 `Question_N` 变量（最新题目）
   - 从用户输入中提取用户回答（格式：`评价答案\n\n用户的回答：xxx`）

2. **生成评价数据**：
   - 查找所有 `EvaluationOutput_N` 变量，使用 N **最大**的值加1作为新的 varName
   - 生成评价数据（scores、evaluation）

3. **调用 `multi_call`**，依次执行：
   - `forward_result`：发送评价给前端（`nextState=""`）
   - `save_doc`：保存历史记录（`nextState=""`，固定为空字符串）
   - `content`：确认信息（如 "所有函数调用已执行完成"）

**重要**：
- `save_doc` 的 `nextState` 固定为 `""`，不再需要判断 `remaining_questions`
- `multi_call` 的 `nextState` 固定为 `""`，流程在 `evaluate_and_save` Thread 结束后停止
- 前端会根据之前收到的题目 JSON 中的 `remaining_questions` 自动决定下一步操作
- LLM 的职责是：生成评价 → 发送给前端 → 保存历史记录 → 结束

## multi_call 示例

```json
{
  "content": "所有函数调用已执行完成",
  "calls": [
    {
      "name": "forward_result",
      "arguments": {
        "varName": "EvaluationOutput_N",
        "result": {
          "type": "evaluation",
          "question_id": 5,
          "scores": {"content_score": 85, "coverage_score": 80, "structure_score": 85},
          "evaluation": {"key_points_hit": [...], "missed_points": [...], "suggestion": "..."}
        },
        "nextState": ""
      }
    },
    {
      "name": "save_doc",
      "arguments": {
        "varName": "InterviewHistory",
        "type": "json",
        "content": "[{...历史记录JSON数组...}]",
        "nextState": ""
      }
    }
  ]
}
```

## 重要提醒

- **`multi_call` 的 `nextState` 固定为 `""`**，前端会根据之前收到的题目 JSON 中的 `remaining_questions` 自动决定下一步操作
- **`save_doc` 的 `nextState` 固定为 `""`**，不再需要判断 `remaining_questions`
- 变量名中的 N 代表调用顺序，N 越大表示越新
- 历史记录必须包含之前的所有记录（从 `InterviewHistory` 变量获取），然后追加当前新记录
