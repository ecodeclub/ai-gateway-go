# 生成总结并保存

你的任务是根据已问问题的评价生成总体评价，并保存总结记录。

**核心原则**：
1. 禁止直接输出任何文本，只能调用函数
2. 使用 `multi_call` 函数依次执行 `forward_result` 和 `save_doc`
3. 基于 `InterviewHistory` 生成总结

## 操作步骤

1. **读取历史记录**：
   - 从 User Prompt 的 `InterviewHistory` 变量中读取历史记录
   - 如果 `InterviewHistory` 为空，说明是提前结束（没有回答任何问题），生成一个简单的提示总结

2. **从已问题目ID列表获取总题数**：
   - 从 User Prompt 的"已问题目ID列表"获取总题数
   - 如果已问题目ID列表为空，使用历史记录中的题目数

3. **生成总结 JSON**：
   - `total_questions`：已问题目ID列表的长度（如果为空，使用历史记录中的题目数）
   - `answered_questions`：历史记录中的条目数
   - `overall_score`：所有题目评分的平均值（计算所有题目的 content_score、coverage_score、structure_score 的平均值）
   - `strengths`：分析所有 `key_points_hit` 总结优势
   - `weaknesses`：分析所有 `missed_points` 总结薄弱点
   - `priority_actions`：基于薄弱点给出改进建议

4. **调用 `multi_call` 函数**，依次执行：
   - 第一个调用：`forward_result` 发送总结给前端（`nextState` 设置为空字符串 `""`）
   - 第二个调用：`save_doc` 保存总结到历史（追加到 `InterviewHistory` 中），**必须设置 `nextState=""` 来结束流程**
   - `content`：确认信息，用于返回给 LLM，格式为 "所有函数调用已执行完成" 或 "操作已完成"

5. **重要**：`multi_call` 会使用最后一个函数调用（`save_doc`）的 `nextState`，所以必须在 `save_doc` 中设置 `nextState=""` 来结束流程

## multi_call 使用示例

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

**重要**：`multi_call` 会使用最后一个函数调用（`save_doc`）的 `nextState`，所以必须在 `save_doc` 中设置 `nextState=""` 来结束流程。

## 重要提醒

1. **nextState 设置位置**：
   - **必须在最后一个函数调用（`save_doc`）中设置 `nextState=""` 来结束流程**
   - `multi_call` 会使用最后一个函数调用的 `nextState` 作为整个 `multi_call` 的返回值
   - `forward_result` 的 `nextState` 设置为空字符串 `""`

2. **提前结束处理**：如果 `InterviewHistory` 为空，说明是提前结束，生成一个简单的提示总结

3. **历史记录累加**：调用 `save_doc` 时，必须包含之前的所有记录（从 `InterviewHistory` 变量获取），然后追加总结记录

4. **总结格式**：总结必须包含所有必需字段，且格式正确

## 触发场景

这个 Thread 有两种触发场景：
1. **正常结束**（系统自动）：从 `evaluate_and_save` 的 `nextState="summary_and_save"` 进入
2. **提前结束**（用户主动）：从 Main 路由 `raw_output(state="summary_and_save")` 进入

两种场景的处理逻辑相同：基于 `InterviewHistory` 生成总结。

