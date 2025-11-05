# 完整重构计划：拆分命令与 Thread 系统提示词优化

## 一、总体目标

1. **拆分前端命令**：将复合命令拆分为原子命令，比如："开始面试"、"评价答案"、"结束面试"。
2. **Main 作为路由器**：使用 `raw_output` 函数做命令路由
3. **Thread 合并优化**：将相似逻辑的 Thread 合并，减少复杂度
4. **使用 `multi_call`**：合并多个函数调用到一个 Thread 中
5. **性能优化**：减少 LLM 推理时间，提升响应速度

## 二、最终架构设计

### 2.1 Thread 映射关系（5个 Thread）

```
Main Thread (cfgID_main)
  ├─ 命令路由器：使用 raw_output 做路由
  
Threads:
  ├─ "get_question" (cfgID_get_question) - 获取题目（合并 start_interview 和 get_next_question）
  ├─ "evaluate_and_save" (cfgID_evaluate_save) - 评价答案并保存历史（合并 evaluate_answer 和 save_history，使用 multi_call）
  ├─ "summary_and_save" (cfgID_summary_save) - 生成总结并保存（合并 generate_summary 和 save_summary，使用 multi_call）
  └─ "send_to_user" (cfgID_send) - 发送题目给用户
```

### 2.2 前端命令定义

| 前端命令 | 触发场景 | 说明 |
|---------|---------|------|
| "开始面试" | 用户点击"开始面试"按钮 | 获取第一题 |
| "评价答案" | 用户回答题目后点击"提交" | 统一命令，后端自动判断流程 |
| "结束面试" | 用户点击"结束面试"按钮 | 提前结束面试，生成总结 |

**重要原则**：
- 前端根据 `remaining_questions` 显示面试进度提示，但**不决定发送的命令**
- 所有评价场景统一发送 "评价答案" 命令
- 后端根据 `remaining_questions` 自动判断下一步流程

### 2.3 Main 路由规则

```markdown
## 命令映射规则（精确匹配）

- 用户输入第一行完全等于 "开始面试" → 调用 `raw_output(state="get_question")`
- 用户输入第一行完全等于 "评价答案" → 调用 `raw_output(state="evaluate_and_save")`
- 用户输入第一行完全等于 "结束面试" → 调用 `raw_output(state="summary_and_save")`
```

### 2.4 完整流程示例

#### 流程1：开始面试
```
前端发送: "开始面试"
  ↓
Main Thread → raw_output(state="get_question")
  ↓
Thread["get_question"]
  ├─ 已问题目ID列表为空 → varName="Question_1"
  ├─ 调用 kbase_rag → nextState="send_to_user"
  ↓
Thread["send_to_user"]
  ├─ 调用 forward_result(题目) → nextState=""
  ↓
完成，返回题目给前端
```

#### 流程2：评价答案并获取下一题（非最后一题）
```
前端发送: "评价答案\n用户的回答：xxx"
  ↓
Main Thread → raw_output(state="evaluate_and_save")
  ↓
Thread["evaluate_and_save"]
  ├─ 生成评价数据
  ├─ 调用 multi_call([
  │     {name: "forward_result", arguments: {评价数据}},
  │     {name: "save_doc", arguments: {历史记录}}
  │   ])
  ├─ 判断: remaining_questions > 1（还有题目）
  ├─ nextState="get_question"
  ↓
Thread["get_question"]
  ├─ 已问题目ID列表不为空 → 查找最大N → varName="Question_{N+1}"
  ├─ 调用 kbase_rag → nextState="send_to_user"
  ↓
Thread["send_to_user"]
  ├─ 调用 forward_result(题目) → nextState=""
  ↓
完成，返回题目给前端
```

#### 流程3：评价答案并生成总结（最后一题，正常结束）
```
前端发送: "评价答案\n用户的回答：xxx"
  ↓
Main Thread → raw_output(state="evaluate_and_save")
  ↓
Thread["evaluate_and_save"]
  ├─ 生成评价数据
  ├─ 调用 multi_call([
  │     {name: "forward_result", arguments: {评价数据}},
  │     {name: "save_doc", arguments: {历史记录}}
  │   ])
  ├─ 判断: remaining_questions == 1（最后一题）
  ├─ nextState="summary_and_save"
  ↓
Thread["summary_and_save"]
  ├─ 生成总结 JSON
  ├─ 调用 multi_call([
  │     {name: "forward_result", arguments: {总结}},
  │     {name: "save_doc", arguments: {总结记录}}
  │   ])
  ├─ nextState=""
  ↓
完成，返回总结给前端
```

#### 流程4：提前结束面试
```
前端发送: "结束面试"
  ↓
Main Thread → raw_output(state="summary_and_save")
  ↓
Thread["summary_and_save"]
  ├─ 从 InterviewHistory 读取历史记录
  ├─ 生成总结 JSON（基于已问问题的评价）
  ├─ 调用 multi_call([
  │     {name: "forward_result", arguments: {总结}},
  │     {name: "save_doc", arguments: {总结记录}}
  │   ])
  ├─ nextState=""
  ↓
完成，返回总结给前端
```

## 三、关键设计要点

### 3.1 `remaining_questions` 逻辑

**重要**：`remaining_questions` 是 ES 返回的剩余题数（**包含当前题**）

- `remaining_questions == 1` → 这是最后一题
- `remaining_questions > 1` → 还有更多题目

### 3.2 前端 `remaining_questions` 使用原则

- **用于显示进度**：前端可以根据 `remaining_questions` 显示面试进度提示
- **不决定命令**：前端**不根据** `remaining_questions` 决定发送什么命令
- **统一命令**：所有评价场景统一发送 "评价答案" 命令

### 3.3 结束面试的两种场景

1. **正常结束**（系统自动）：
   - 用户回答最后一题 → 发送 "评价答案"
   - 后端 `evaluate_and_save` 判断 `remaining_questions == 1`
   - 自动进入 `summary_and_save` → 生成总结

2. **提前结束**（用户主动）：
   - 用户点击 "结束面试" 按钮 → 发送 "结束面试"
   - Main 路由到 `summary_and_save` → 生成总结

## 四、详细实施步骤

### 步骤1：创建系统提示词文件（5个文件）

1. **`system_prompt_main.md`** - Main 路由器（约 30 行）
   - 命令路由逻辑
   - 精确匹配规则

2. **`system_prompt_get_question.md`** - 获取题目（约 35 行）
   - 判断 varName（第一题 vs 下一题）
   - 调用 kbase_rag

3. **`system_prompt_evaluate_save.md`** - 评价答案并保存历史（约 50 行）
   - 生成评价数据
   - 使用 multi_call 调用 forward_result 和 save_doc
   - 根据 remaining_questions 判断 nextState

4. **`system_prompt_summary_save.md`** - 生成总结并保存（约 45 行）
   - 生成总结 JSON
   - 使用 multi_call 调用 forward_result 和 save_doc

5. **`system_prompt_send.md`** - 发送题目（约 25 行）
   - 格式化题目数据
   - 调用 forward_result

### 步骤2：创建用户提示词文件（5个文件）

1. **`user_prompt_main.md`** - Main 路由器（约 10 行）
2. **`user_prompt_get_question.md`** - 获取题目（约 15 行）
3. **`user_prompt_evaluate_save.md`** - 评价答案并保存历史（约 20 行）
4. **`user_prompt_summary_save.md`** - 生成总结并保存（约 20 行）
5. **`user_prompt_send.md`** - 发送题目（约 15 行）

### 步骤3：修改前端代码 (`interview.html`)

#### 3.1 修改命令发送逻辑

```javascript
// 修改前：
if (isLastQuestion) {
    command = "评价答案并生成面试总结";
} else {
    command = "评价答案并获取下一题";
}

// 修改后：
// 统一发送 "评价答案"，不根据 isLastQuestion 判断
const userInput = `评价答案\n\n用户的回答：${transcript}`;
await streamRequest(userInput);
```

#### 3.2 保留进度提示（可选）

```javascript
// 保留 isLastQuestion 用于UI展示
if (q.remaining_questions === 1) {
    isLastQuestion = true;
    // 可以显示提示："这是最后一题"
} else {
    isLastQuestion = false;
}
```

#### 3.3 修改结束面试按钮

```javascript
els.btnEnd.onclick = async () => {
    // 发送 "结束面试" 命令（提前结束）
    await streamRequest("结束面试");
};
```

### 步骤4：修改测试代码 (`interview_test.go`)

#### 4.1 创建 5 个 `InvocationConfig`

```go
cfgID_main := 100001        // Main 路由器
cfgID_get_question := 100002 // 获取题目
cfgID_evaluate_save := 100003 // 评价答案并保存
cfgID_summary_save := 100004  // 生成总结并保存
cfgID_send := 100005          // 发送题目
```

#### 4.2 创建 5 个 `InvocationConfigVersion`

每个版本对应：
- 系统提示词（从对应的 `system_prompt_*.md` 文件读取）
- 用户提示词（从对应的 `user_prompt_*.md` 文件读取）
- 对应的函数列表

#### 4.3 修改 `BizOrchestration` 配置

```go
biz.Config = domain.BizConfig{
    Orchestration: domain.Orchestration{
        Main: domain.NewThread(cfgID_main),
        Threads: map[string]*domain.Thread{
            "get_question":      domain.NewThread(cfgID_get_question),
            "evaluate_and_save": domain.NewThread(cfgID_evaluate_save),
            "summary_and_save":  domain.NewThread(cfgID_summary_save),
            "send_to_user":      domain.NewThread(cfgID_send),
        },
    },
}
```

#### 4.4 为每个 ConfigID 注册对应的函数

- **`cfgID_main`**: 只注册 `raw_output`
- **`cfgID_get_question`**: 注册 `kbase_rag`
- **`cfgID_evaluate_save`**: 注册 `multi_call`（multi_call 内部会调用 `forward_result` 和 `save_doc`）
- **`cfgID_summary_save`**: 注册 `multi_call`（multi_call 内部会调用 `forward_result` 和 `save_doc`）
- **`cfgID_send`**: 注册 `forward_result`

#### 4.5 创建 `multi_call` 函数的 JSON Schema

```json
{
  "name": "multi_call",
  "description": "依次执行多个函数调用。所有函数调用会按顺序执行，最后一个函数的 nextState 会作为整个 multi_call 的 nextState。",
  "strict": true,
  "parameters": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "calls": {
        "type": "array",
        "description": "要执行的函数调用列表，按顺序执行",
        "items": {
          "type": "object",
          "additionalProperties": false,
          "properties": {
            "name": {
              "type": "string",
              "enum": ["forward_result", "save_doc"],
              "description": "函数名。必须是 forward_result 或 save_doc 之一。"
            },
            "arguments": {
              "type": "object",
              "description": "函数参数对象。根据 name 字段的值，构造对应的参数结构。",
              "additionalProperties": true
            }
          },
          "required": ["name", "arguments"]
        },
        "minItems": 1
      },
      "nextState": {
        "type": "string",
        "enum": ["get_question", "send_to_user", "summary_and_save", ""],
        "description": "最终的下一个状态。如果最后一个函数调用已经设置了 nextState，这里可以不设置或设置为空字符串。系统会使用最后一个函数调用的 nextState。"
      }
    },
    "required": ["calls"]
  }
}
```

## 五、文件清单

### 5.1 新增文件（10个）

```
internal/test/interview/
├── system_prompt_main.md          # Main 路由器
├── system_prompt_get_question.md  # 获取题目
├── system_prompt_evaluate_save.md # 评价答案并保存历史
├── system_prompt_summary_save.md  # 生成总结并保存
├── system_prompt_send.md          # 发送题目
├── user_prompt_main.md            # Main 用户提示词
├── user_prompt_get_question.md    # 获取题目用户提示词
├── user_prompt_evaluate_save.md   # 评价答案并保存用户提示词
├── user_prompt_summary_save.md    # 生成总结并保存用户提示词
└── user_prompt_send.md            # 发送题目用户提示词
```

### 5.2 修改文件

1. **`internal/test/interview/interview_test.go`**
   - 创建 5 个 `InvocationConfig`
   - 创建 5 个 `InvocationConfigVersion`
   - 修改 `BizOrchestration` 配置
   - 为每个 ConfigID 注册对应的函数
   - 添加 `multi_call` 函数的 JSON Schema 定义

2. **`internal/test/interview/interview.html`**
   - 修改命令发送逻辑（统一发送 "评价答案"）
   - 保留 `remaining_questions` 用于进度提示（可选）

### 5.3 保留文件（暂不删除）

- `system_prompt_v2.md` - 作为参考
- `user_prompt_v2.md` - 作为参考

## 六、系统提示词设计要点

### 6.1 Main 路由器 (`system_prompt_main.md`)

```markdown
# 命令路由器

你的任务是根据用户输入，调用 `raw_output` 函数设置对应的状态。

## 命令映射规则（精确匹配）

- 用户输入第一行完全等于 "开始面试" → 调用 `raw_output(state="get_question")`
- 用户输入第一行完全等于 "评价答案" → 调用 `raw_output(state="evaluate_and_save")`
- 用户输入第一行完全等于 "结束面试" → 调用 `raw_output(state="summary_and_save")`

## 操作步骤

1. 读取用户输入的第一行（去除用户回答部分）
2. 精确匹配上面的命令（完全相等）
3. 调用 `raw_output` 函数，设置对应的 `state` 参数

**重要**：
- 必须精确匹配，不能使用模糊匹配
- 如果用户输入不匹配任何命令，调用 `raw_output(state="")` 表示未知命令
- 你不需要执行任何业务逻辑，只需要做路由决策
```

### 6.2 评价答案并保存 (`system_prompt_evaluate_save.md`)

```markdown
# 评价答案并保存历史

你的任务是根据用户的回答，生成评价数据，发送给前端，并保存历史记录。

## 操作步骤

1. 从 User Prompt 中获取最新题目和用户回答
2. 生成评价数据（scores、evaluation）
3. 调用 `multi_call` 函数，依次执行：
   - 第一个调用：`forward_result` 发送评价给前端
   - 第二个调用：`save_doc` 保存历史记录
4. 根据 `remaining_questions` 设置 `nextState`：
   - 如果 `remaining_questions > 1`：还有题目，设置 `nextState="get_question"`
   - 如果 `remaining_questions == 1`：这是最后一题，设置 `nextState="summary_and_save"`

**重要**：
- `remaining_questions` 是 ES 返回的剩余题数（包含当前题）
- `remaining_questions == 1` 表示这是最后一题
- `remaining_questions > 1` 表示还有更多题目

## multi_call 使用示例

```json
{
  "calls": [
    {
      "name": "forward_result",
      "arguments": {
        "varName": "EvaluationOutput_N",
        "result": {
          "type": "evaluation",
          "question_id": 1,
          "scores": {...},
          "evaluation": {...}
        },
        "nextState": ""
      }
    },
    {
      "name": "save_doc",
      "arguments": {
        "varName": "InterviewHistory",
        "type": "json",
        "content": "[...历史记录JSON数组...]",
        "nextState": ""
      }
    }
  ],
  "nextState": "get_question"
}
```
```

### 6.3 生成总结并保存 (`system_prompt_summary_save.md`)

类似结构，但处理总结生成逻辑。

## 七、测试验证

### 7.1 测试用例

1. **开始面试流程**
   - 前端发送 "开始面试"
   - 验证：Main 路由到 `get_question` → `send_to_user` → 返回题目

2. **评价答案流程（非最后一题）**
   - 前端发送 "评价答案\n用户的回答：xxx"
   - 验证：Main 路由到 `evaluate_and_save` → `get_question` → `send_to_user` → 返回题目

3. **评价答案流程（最后一题，正常结束）**
   - 前端发送 "评价答案\n用户的回答：xxx"（remaining_questions == 1）
   - 验证：Main 路由到 `evaluate_and_save` → `summary_and_save` → 返回总结

4. **提前结束面试流程**
   - 前端发送 "结束面试"
   - 验证：Main 路由到 `summary_and_save` → 返回总结

### 7.2 性能验证

- 对比重构前后的响应时间
- 验证系统提示词大小是否显著减少
- 验证 LLM 推理时间是否减少

## 八、预期效果

1. **系统提示词大小**：从 323 行减少到每个约 25-50 行
2. **Thread 数量**：从 9 个减少到 5 个
3. **文件数量**：从 18 个减少到 10 个
4. **LLM 推理时间**：预计减少 40-60%
5. **代码可维护性**：每个 Thread 独立，易于维护和扩展
6. **前端逻辑简化**：统一发送 "评价答案" 命令，不根据状态判断

## 九、实施顺序

1. ✅ 创建系统提示词文件（5 个）
2. ✅ 创建用户提示词文件（5 个）
3. ✅ 修改测试代码（创建 5 个 ConfigID，配置 `multi_call`）
4. ✅ 修改前端代码（统一发送 "评价答案" 命令）
5. ✅ 运行测试验证
6. ✅ 性能对比测试

## 十、风险评估

### 10.1 潜在问题

1. **`multi_call` 函数调用失败**：如果某个函数调用失败，整个流程可能中断
   - **缓解**：在代码中添加错误处理，确保错误被正确传播

2. **`nextState` 设置错误**：如果 `multi_call` 的 `nextState` 设置错误，可能导致流程中断
   - **缓解**：在系统提示词中明确说明 `nextState` 的设置规则

3. **变量传递问题**：`multi_call` 中多个函数调用之间的变量传递
   - **缓解**：`Chat.Vars` 是全局的，应该可以共享

4. **`remaining_questions` 判断错误**：如果 LLM 判断错误，可能导致流程混乱
   - **缓解**：在系统提示词中明确说明判断规则，强调 `remaining_questions == 1` 表示最后一题

### 10.2 回滚方案

如果重构出现问题，可以：
1. 保留原有的 `system_prompt_v2.md` 和 `user_prompt_v2.md`
2. 快速回滚到使用单个 ConfigID 的版本

---

## 附录：关键设计决策

### A. 为什么使用 `multi_call`？

- **减少 Thread 数量**：将多个相关操作合并到一个 Thread 中
- **减少 LLM 往返次数**：一次调用完成多个操作
- **逻辑更清晰**：相关操作在一个地方处理

### B. 为什么前端不根据 `remaining_questions` 决定命令？

- **简化前端逻辑**：前端只需要发送命令，不需要理解业务逻辑
- **后端统一处理**：后端根据上下文自动判断流程，更可靠
- **解耦**：前端和后端解耦，易于维护

### C. 为什么 `remaining_questions == 1` 表示最后一题？

- **ES 返回逻辑**：ES 返回的 `remaining_questions` 包含当前题
- **判断规则**：如果剩余题数包含当前题，且值为 1，说明这是最后一题

---

**文档版本**：v1.0  
**最后更新**：2025-01-XX  
**作者**：AI Assistant

