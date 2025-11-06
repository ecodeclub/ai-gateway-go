# 前端驱动面试流程重构计划

## 一、问题分析

### 当前问题
1. **LLM 需要判断 `remaining_questions`**：在 `evaluate_and_save` Thread 中，LLM 需要从 `QuestionOutput_N` 变量中读取 `remaining_questions` 并判断 `nextState`，增加了 LLM 的推理负担。
2. **前端已经知道 `remaining_questions`**：前端在收到题目 JSON 时已经获取了 `remaining_questions`，并设置了 `isLastQuestion` 标志。
3. **重复判断**：前端和 LLM 都在判断是否最后一题，存在重复逻辑。

### 优化方案
**前端驱动流程**：前端根据 `isLastQuestion` 自动决定下一步操作，LLM 的 `nextState` 固定为 `""`，简化 LLM 的职责。

## 二、状态流转图

### 2.1 当前流程（LLM 驱动）

```
┌─────────────────┐
│  前端：开始面试  │
└────────┬────────┘
         │ state="get_question"
         ▼
┌─────────────────────┐
│ Thread: get_question │
│ - kbase_rag         │
│ - nextState="send_to_user" │
└────────┬────────────┘
         ▼
┌─────────────────────┐
│ Thread: send_to_user│
│ - forward_result    │
│ - nextState=""      │
└────────┬────────────┘
         ▼
┌─────────────────┐
│ 前端：收到题目   │
│ remaining_questions│
│ isLastQuestion  │
└────────┬────────┘
         │ 用户回答
         ▼
┌─────────────────────┐
│ Thread: evaluate_and_save │
│ - multi_call:       │
│   1. forward_result │
│   2. save_doc       │
│ - LLM判断remaining_questions │
│ - nextState="get_question"  │
│   或 "summary_and_save"     │
└────────┬────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌─────────────────┐
│get_question│ │summary_and_save│
└─────────┘ └─────────────────┘
```

### 2.2 优化后流程（前端驱动）

```
┌─────────────────┐
│  前端：开始面试  │
└────────┬────────┘
         │ 发送: state="get_question", input="开始面试"
         ▼
┌─────────────────────┐
│ Gateway 路由        │
│ → Thread: get_question │
└────────┬────────────┘
         ▼
┌─────────────────────┐
│ LLM: get_question   │
│ - kbase_rag         │
│ - nextState="send_to_user" │
└────────┬────────────┘
         ▼
┌─────────────────────┐
│ LLM: send_to_user   │
│ - forward_result    │
│ - nextState=""      │
└────────┬────────────┘
         ▼
┌─────────────────┐
│ 前端：收到题目   │
│ {type:"question",│
│  remaining_questions:5,│
│  question_id:6}  │
│ → 设置isLastQuestion│
└────────┬────────┘
         │ 用户回答："xxx"
         │ 前端发送: state="evaluate_and_save"
         │          input="评价答案\n\n用户的回答：xxx"
         ▼
┌─────────────────────┐
│ Gateway 路由        │
│ → Thread: evaluate_and_save │
└────────┬────────────┘
         ▼
┌─────────────────────┐
│ LLM: evaluate_and_save │
│ - multi_call:       │
│   1. forward_result │
│      (发送评价JSON) │
│   2. save_doc       │
│      (保存历史记录) │
│ - nextState="" (固定)│
│ → 流程结束          │
└────────┬────────────┘
         ▼
┌─────────────────┐
│ 前端：收到评价   │
│ {type:"evaluation",│
│  question_id:6,  │
│  scores:{...},   │
│  evaluation:{...}}│
│ → 显示评价       │
│ → 根据isLastQuestion│
│   自动发送下一步 │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌─────────────────┐
│isLastQuestion=false│ │isLastQuestion=true│
│发送: state="get_question"│ │发送: state="summary_and_save"│
│      input="获取下一题"│ │      input="结束面试"│
└─────────┘ └─────────────────┘
         │         │
         └────┬────┘
              ▼
         (继续循环或结束)
```

## 三、前端 - Gateway - LLM 交互逻辑

### 3.1 完整交互流程

#### 流程1：开始面试
```
前端 → Gateway → LLM
  │      │        │
  │      │        ├─ Thread: get_question
  │      │        │  - kbase_rag (获取题目)
  │      │        │  - nextState="send_to_user"
  │      │        │
  │      │        ├─ Thread: send_to_user
  │      │        │  - forward_result (发送题目JSON)
  │      │        │  - nextState=""
  │      │        │
  │      │        └─ 返回题目JSON给前端
  │      │
  │      └─ 路由到对应Thread，返回SSE流
  │
  └─ 收到题目JSON，设置isLastQuestion
```

#### 流程2：评价答案（非最后一题）
```
前端 → Gateway → LLM
  │      │        │
  │      │        ├─ Thread: evaluate_and_save
  │      │        │  - multi_call:
  │      │        │    1. forward_result (发送评价JSON)
  │      │        │    2. save_doc (保存历史记录)
  │      │        │  - nextState="" (固定，流程结束)
  │      │        │
  │      │        └─ 返回评价JSON给前端
  │      │
  │      └─ 路由到evaluate_and_save Thread，返回SSE流
  │
  └─ 收到评价JSON
     └─ 显示评价
     └─ 根据isLastQuestion=false，自动发送：
        state="get_question", input="获取下一题"
```

#### 流程3：评价答案（最后一题）
```
前端 → Gateway → LLM
  │      │        │
  │      │        ├─ Thread: evaluate_and_save
  │      │        │  - multi_call:
  │      │        │    1. forward_result (发送评价JSON)
  │      │        │    2. save_doc (保存历史记录)
  │      │        │  - nextState="" (固定，流程结束)
  │      │        │
  │      │        └─ 返回评价JSON给前端
  │      │
  │      └─ 路由到evaluate_and_save Thread，返回SSE流
  │
  └─ 收到评价JSON
     └─ 显示评价
     └─ 根据isLastQuestion=true，自动发送：
        state="summary_and_save", input="结束面试"
```

#### 流程4：获取下一题
```
前端 → Gateway → LLM
  │      │        │
  │      │        ├─ Thread: get_question
  │      │        │  - kbase_rag (获取题目，排除已问)
  │      │        │  - nextState="send_to_user"
  │      │        │
  │      │        ├─ Thread: send_to_user
  │      │        │  - forward_result (发送题目JSON)
  │      │        │  - nextState=""
  │      │        │
  │      │        └─ 返回题目JSON给前端
  │      │
  │      └─ 路由到get_question Thread，返回SSE流
  │
  └─ 收到题目JSON，更新isLastQuestion
```

#### 流程5：结束面试
```
前端 → Gateway → LLM
  │      │        │
  │      │        ├─ Thread: summary_and_save
  │      │        │  - multi_call:
  │      │        │    1. forward_result (发送总结JSON)
  │      │        │    2. save_doc (保存总结)
  │      │        │  - nextState="" (固定，流程结束)
  │      │        │
  │      │        └─ 返回总结JSON给前端
  │      │
  │      └─ 路由到summary_and_save Thread，返回SSE流
  │
  └─ 收到总结JSON，显示总结
```

### 3.2 关键交互点说明

#### 3.2.1 前端发送命令格式

**命令类型和状态映射**：
1. **开始面试**：
   - `state="get_question"`
   - `input="开始面试"`

2. **获取题目**（非第一题）：
   - `state="get_question"`
   - `input="获取下一题"`

3. **获取评价**：
   - `state="evaluate_and_save"`
   - `input="评价答案\n\n用户的回答：xxx"`

4. **结束面试**：
   - `state="summary_and_save"`
   - `input="结束面试"`

**完整流程**：
```
开始面试 (state=get_question, input="开始面试")
  ↓
获取题目 (返回题目JSON，前端设置isLastQuestion)
  ↓
获取评价 (state=evaluate_and_save, input="评价答案\n\n用户的回答：xxx")
  ↓
显示评价 (前端收到评价JSON)
  ↓
判断：如果是最后一题？
  ├─ 是 → 自动发送"结束面试" (state=summary_and_save, input="结束面试")
  └─ 否 → 自动发送"获取下一题" (state=get_question, input="获取下一题")
      ↓
      获取题目 (返回题目JSON，前端更新isLastQuestion)
      ↓
      获取评价 (state=evaluate_and_save, input="评价答案\n\n用户的回答：xxx")
      ↓
      (循环...)
```

#### 3.2.2 LLM 职责划分
- **`evaluate_and_save` Thread**：
  - ✅ 生成评价数据
  - ✅ 发送评价JSON给前端
  - ✅ 保存历史记录
  - ❌ **不再判断 `remaining_questions`**
  - ❌ **不再设置 `nextState` 为 `"get_question"` 或 `"summary_and_save"`**
  - ✅ **固定 `nextState=""`，流程结束**

#### 3.2.3 前端自动决策
- **收到评价JSON后**：
  - 如果 `isLastQuestion === false`：自动发送 `state="get_question"`, `input="获取下一题"`
  - 如果 `isLastQuestion === true`：自动发送 `state="summary_and_save"`, `input="结束面试"`

**重要说明**：
- **每个题目都要有评价**：不管是第一题还是最后一题，都需要经过"获取评价"步骤
- **"开始面试"和"获取下一题"使用相同的state**：都是 `state="get_question"`，但命令文本不同（`input="开始面试"` vs `input="获取下一题"`）
- **最后一题的评价后自动结束**：前端收到最后一题的评价后，自动发送"结束面试"命令，不再发送"获取下一题"

## 四、详细实现方案

### 4.1 前端修改

#### 4.1.1 `displayEvaluation` 函数
**位置**：`interview.html`

**修改内容**：
- 在 `displayEvaluation` 函数中，根据 `isLastQuestion` 自动发送下一步命令
- 如果 `isLastQuestion === true`：自动发送 `streamRequest("结束面试", "summary_and_save", false)`
- 如果 `isLastQuestion === false`：自动发送 `streamRequest("获取下一题", "get_question", false)`

**代码示例**：
```javascript
function displayEvaluation(e) {
    // ... 渲染评分和评价详情 ...
    
    // 🚨 前端驱动：根据 isLastQuestion 自动发送下一步命令
    if (isLastQuestion) {
        // 最后一题：自动发送"结束面试"命令
        console.log('📝 最后一题，自动发送"结束面试"命令');
        streamRequest("结束面试", "summary_and_save", false).catch(err => {
            console.error('❌ 自动发送"结束面试"失败:', err);
        });
    } else {
        // 还有更多题：自动发送"获取下一题"命令
        console.log('📝 还有更多题，自动发送"获取下一题"命令');
        streamRequest("获取下一题", "get_question", false).catch(err => {
            console.error('❌ 自动发送"获取下一题"失败:', err);
        });
    }
}
```

### 4.2 后端修改

#### 4.2.1 `system_prompt_evaluate_save.md`
**修改内容**：
- **移除 `remaining_questions` 判断逻辑**：LLM 不再需要读取 `QuestionOutput_N` 变量并判断 `remaining_questions`
- **`save_doc` 的 `nextState` 固定为 `""`**：不再根据 `remaining_questions` 设置 `nextState`
- **`multi_call` 的 `nextState` 固定为 `""`**：流程在 `evaluate_and_save` Thread 结束后停止
- **添加说明**：前端会根据之前收到的题目 JSON 中的 `remaining_questions` 自动决定下一步操作

**修改后的关键部分**：
```markdown
3. **调用 `multi_call`**，依次执行：
   - `forward_result`：发送评价给前端（`nextState=""`）
   - `save_doc`：保存历史记录（`nextState=""`，固定为空字符串）
   - `content`：确认信息（如 "所有函数调用已执行完成"）

**重要**：
- `save_doc` 的 `nextState` 固定为 `""`，不再需要判断 `remaining_questions`
- `multi_call` 的 `nextState` 固定为 `""`，流程在 `evaluate_and_save` Thread 结束后停止
- 前端会根据之前收到的题目 JSON 中的 `remaining_questions` 自动决定下一步操作
- LLM 的职责是：生成评价 → 发送给前端 → 保存历史记录 → 结束
```

## 五、优势分析

### 5.1 简化 LLM 职责
- **移除判断逻辑**：LLM 不再需要读取 `QuestionOutput_N` 变量并判断 `remaining_questions`
- **固定 `nextState`**：所有 `nextState` 固定为 `""`，减少 LLM 的推理负担
- **更快的响应**：LLM 只需要生成评价和保存历史，响应时间更短

### 5.2 前端控制更精确
- **前端已有信息**：前端在收到题目 JSON 时已经知道 `remaining_questions` 和 `isLastQuestion`
- **避免重复判断**：不再需要 LLM 和前端都判断是否最后一题
- **更清晰的流程**：前端根据状态自动决定下一步，流程更清晰

### 5.3 减少错误
- **避免 LLM 误判**：不再依赖 LLM 正确读取和判断 `remaining_questions`
- **前端逻辑简单**：前端只需要根据 `isLastQuestion` 布尔值决定，逻辑简单可靠

## 六、潜在问题与解决方案

### 6.1 问题：前端和 LLM 状态不一致
**场景**：如果前端收到的题目 JSON 中的 `remaining_questions` 与 LLM 看到的不一致怎么办？

**解决方案**：
- 前端始终以最新收到的题目 JSON 中的 `remaining_questions` 为准
- LLM 不再判断 `remaining_questions`，避免不一致

### 6.2 问题：前端自动发送命令的时机
**场景**：前端在 `displayEvaluation` 中自动发送命令，是否会影响用户体验？

**解决方案**：
- 前端仍然显示评价内容，用户可以查看
- 自动发送命令是异步的，不会阻塞 UI
- 如果用户想提前结束，可以点击"结束面试"按钮（会覆盖自动发送的命令）

## 七、实施步骤

1. **讨论确认**：与用户确认方案，包括状态流转图和实现细节
2. **修改前端**：修改 `displayEvaluation` 函数，添加自动发送命令逻辑
3. **修改后端提示词**：修改 `system_prompt_evaluate_save.md`，移除判断逻辑
4. **测试验证**：
   - 测试非最后一题：验证前端自动发送"获取下一题"
   - 测试最后一题：验证前端自动发送"结束面试"
   - 测试提前结束：验证用户点击"结束面试"按钮仍然有效

## 八、待讨论问题

1. **前端自动发送的时机**：是否需要在 `displayEvaluation` 中立即发送，还是延迟一段时间？
2. **错误处理**：如果自动发送命令失败，是否需要重试或提示用户？
3. **用户体验**：自动发送命令是否会影响用户查看评价的时间？

