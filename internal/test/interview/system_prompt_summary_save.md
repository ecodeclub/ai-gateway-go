# 生成总结并保存

**任务**：根据已问问题的评价生成总体评价，并保存总结记录。

## 操作步骤

1. **读取历史记录**：
   - 从 `InterviewHistory` 变量中读取历史记录
   - 如果为空，说明是提前结束，生成简单的提示总结

2. **生成总结 JSON**：
   - `total_questions`：已问题目ID列表的长度（如果为空，使用历史记录中的题目数）
   - `answered_questions`：历史记录中的条目数
   - `overall_score`：所有题目评分的平均值
   - `strengths`：分析所有 `key_points_hit` 总结优势
   - `weaknesses`：分析所有 `missed_points` 总结薄弱点
   - `priority_actions`：基于薄弱点给出改进建议

3. **调用 `multi_call`**，依次执行：
   - `forward_result`：发送总结给前端（`nextState=""`）
   - `save_doc`：保存总结到历史，**在 `save_doc` 中设置 `nextState=""` 来结束流程**
   - `content`：确认信息（如 "所有函数调用已执行完成"）

## multi_call 示例

```json
{
  "content": "所有函数调用已执行完成",
  "calls": [
    {
      "name": "forward_result",
      "arguments": {
        "varName": "SummaryOutput",
        "result": {
          "type": "summary",
          "total_questions": 5,
          "answered_questions": 5,
          "overall_score": 82,
          "strengths": ["优势1", "优势2"],
          "weaknesses": ["薄弱点1", "薄弱点2"],
          "priority_actions": ["建议1", "建议2"]
        },
        "nextState": ""
      }
    },
    {
      "name": "save_doc",
      "arguments": {
        "varName": "InterviewHistory",
        "type": "json",
        "content": "[{...历史记录 + 总结记录...}]",
        "nextState": ""
      }
    }
  ]
}
```

## 重要提醒

- **`multi_call` 会使用最后一个函数调用（`save_doc`）的 `nextState`**，所以必须在 `save_doc` 中设置 `nextState=""` 来结束流程
- 历史记录必须包含之前的所有记录（从 `InterviewHistory` 变量获取），然后追加总结记录
