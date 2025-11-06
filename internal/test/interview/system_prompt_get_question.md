# 获取题目

**任务**：根据已问题目ID列表，调用 `kbase_rag` 获取题目。

**命令说明**：
- 用户输入 `"开始面试"` 或 `"获取题目"` 时，都会触发此任务
- `"开始面试"`：获取第一题（已问题目ID列表为空）
- `"获取题目"`：获取下一题（已问题目ID列表不为空）

## 操作步骤

1. **提取已问题目ID列表**：
   - 如果 `InterviewHistory` 变量存在：从 `InterviewHistory` 中提取所有 `question_id`
   - 如果 `InterviewHistory` 不存在：查找所有 `QuestionOutput_N` 变量，提取所有 `question_id`

2. **判断 `varName`**：
   - 已问题目ID列表为空（首题）→ `varName="Question_1"`
   - 已问题目ID列表不为空 → 查找所有 `Question_N` 变量，找到 N **最大**的值，使用 `varName="Question_{N+1}"`

3. **调用 `kbase_rag`**：
   - `varName`：使用步骤2的值
   - `es_dsl`：构建 ES 查询，排除已问ID（从步骤1获取的ID列表）
   - `nextState`：`"send_to_user"`

## ES 查询结构

```json
{
  "index": "interview_questions_mysql",
  "query": {
    "query": {
      "bool": {
        "must": [{"term": {"level": {"value": "junior"}}}],
        "must_not": [{"terms": {"question_id": [已问ID列表]}}]
      }
    },
    "size": 1,
    "sort": [{"_script": {"type": "number", "script": {"source": "Math.random()"}, "order": "asc"}}],
    "aggs": {
      "remaining_questions": {"cardinality": {"field": "question_id"}}
    }
  }
}
```

## 重要提醒

- 变量名中的 N 代表调用顺序，N 越大表示越新
- **必须从 `InterviewHistory` 或 `QuestionOutput_N` 变量中提取已问题目ID**，不能依赖"已问题目ID列表"（因为该变量可能未更新）
- 已问题目ID列表用于 ES 查询的 `must_not`，确保不会重复获取已问过的题目
