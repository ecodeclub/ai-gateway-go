# 重构方案 V2：跳过 Main Thread，直接使用 state 参数

## 一、重构目标

**核心目标**：去掉 Main Thread 的路由逻辑，前端直接传递 `state` 参数，减少一次 LLM 调用，提升响应速度。

**性能提升**：预计从点击"开始面试"到获取第一题的时间从 30 秒减少到 20 秒左右（节省 Main Thread 的 LLM 调用时间）。

## 二、架构变化

### 2.1 Thread 数量变化

**重构前**：5 个 Thread（Main + 4 个业务 Thread）
```
Main Thread (cfgID_main) - 命令路由器
  ├─ "get_question" (cfgID_get_question)
  ├─ "evaluate_and_save" (cfgID_evaluate_save)
  ├─ "summary_and_save" (cfgID_summary_save)
  └─ "send_to_user" (cfgID_send)
```

**重构后**：4 个 Thread（去掉 Main）
```
Threads:
  ├─ "get_question" (cfgID_get_question)
  ├─ "evaluate_and_save" (cfgID_evaluate_save)
  ├─ "summary_and_save" (cfgID_summary_save)
  └─ "send_to_user" (cfgID_send)
```

### 2.2 前端命令到 state 映射

| 前端命令 | state 参数 | 说明 |
|---------|-----------|------|
| "开始面试" | `"get_question"` | 获取第一题 |
| "评价答案" | `"evaluate_and_save"` | 评价答案并保存历史 |
| "结束面试" | `"summary_and_save"` | 生成总结并保存 |

### 2.3 流程变化对比

#### 重构前（使用 Main Thread）
```
前端发送: "开始面试"
  ↓
Main Thread → LLM 推理 → raw_output(state="get_question") [~6秒]
  ↓
Thread["get_question"] → kbase_rag → nextState="send_to_user" [~11秒]
  ↓
Thread["send_to_user"] → forward_result → nextState="" [~3秒]
  ↓
完成，返回题目给前端
总耗时：~20秒
```

#### 重构后（直接使用 state）
```
前端发送: state="get_question", input="开始面试"
  ↓
Thread["get_question"] → kbase_rag → nextState="send_to_user" [~11秒]
  ↓
Thread["send_to_user"] → forward_result → nextState="" [~3秒]
  ↓
完成，返回题目给前端
总耗时：~14秒（节省 ~6秒）
```

## 三、详细实施步骤

### 步骤1：修改前端代码 (`interview.html`)

#### 1.1 修改 `streamRequest` 函数，添加 `state` 参数

```javascript
// 修改前：
async function streamRequest(userInput, shouldUpdateJudgeBox = true) {
    const response = await fetch(`${INTERVIEW_API}/api/interview/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_sn: chatSn,
            input: userInput,
            uid: 123
        })
    });
    // ...
}

// 修改后：
async function streamRequest(userInput, state, shouldUpdateJudgeBox = true) {
    const response = await fetch(`${INTERVIEW_API}/api/interview/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            chat_sn: chatSn,
            input: userInput,
            uid: 123,
            state: state  // 新增 state 参数
        })
    });
    // ...
}
```

#### 1.2 修改所有调用 `streamRequest` 的地方

```javascript
// 1. 开始面试
els.btnStartInterview.onclick = async () => {
    // ...
    await streamRequest("开始面试", "get_question", false);
    // ...
};

// 2. 评价答案（提交按钮）
els.btnSubmit.onclick = async () => {
    // ...
    const userInput = `评价答案\n\n用户的回答：${transcript}`;
    await streamRequest(userInput, "evaluate_and_save");
    // ...
};

// 3. 评价答案（不会按钮）
els.btnDontKnow.onclick = async () => {
    // ...
    const userInput = `评价答案\n\n用户的回答：${answer}`;
    await streamRequest(userInput, "evaluate_and_save");
    // ...
};

// 4. 结束面试
els.btnEnd.onclick = async () => {
    // ...
    await streamRequest("结束面试", "summary_and_save", false);
    // ...
};
```

### 步骤2：修改测试代码 (`interview_test.go`)

#### 2.1 删除 Main Thread 相关配置

**删除内容**：
- `cfgIDMain` 变量定义
- Main Thread 的 `InvocationConfig` 创建
- Main Thread 的 `InvocationConfigVersion` 创建
- `raw_output` 函数的注册
- `system_prompt_main.md` 和 `user_prompt_main.md` 的 embed

**修改 `BizOrchestration` 配置**：

```go
// 修改前：
biz.Config = domain.BizConfig{
    Orchestration: domain.Orchestration{
        Main: &domain.Thread{CfgID: cfgIDMain},
        Threads: map[string]*domain.Thread{
            "get_question":      {CfgID: cfgIDGetQuestion},
            "evaluate_and_save": {CfgID: cfgIDEvaluateSave},
            "summary_and_save":  {CfgID: cfgIDSummarySave},
            "send_to_user":      {CfgID: cfgIDSend},
        },
    },
}

// 修改后：
biz.Config = domain.BizConfig{
    Orchestration: domain.Orchestration{
        Main: nil,  // Main 不再使用，但保留字段以兼容结构体
        Threads: map[string]*domain.Thread{
            "get_question":      {CfgID: cfgIDGetQuestion},
            "evaluate_and_save": {CfgID: cfgIDEvaluateSave},
            "summary_and_save":  {CfgID: cfgIDSummarySave},
            "send_to_user":      {CfgID: cfgIDSend},
        },
    },
}
```

#### 2.2 修改 HTTP 代理，传递 state 参数

检查 `interview_test.go` 中的 HTTP 代理代码，确保 `state` 参数被正确传递到 gRPC 请求。

### 步骤3：Orchestrator 逻辑保持不变

**重要**：Orchestrator 的现有逻辑保持不变，它已经支持通过 `state` 参数跳过 Main Thread：
- 如果 `state != ""`，直接使用 `Threads[state]`，跳过 Main
- 如果 `state == ""`，使用 Main（但我们的业务中前端总是传递 state，所以不会走到 Main）

**不需要修改** `internal/service/orchestrator/orchestrator.go` 文件。

### 步骤4：删除 Main Thread 相关文件

**删除文件**：
- `system_prompt_main.md`
- `user_prompt_main.md`

**保留文件**（作为参考）：
- `system_prompt_v2.md`
- `user_prompt_v2.md`

## 四、完整流程示例（重构后）

### 流程1：开始面试
```
前端发送: state="get_question", input="开始面试"
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

### 流程2：评价答案并获取下一题（非最后一题）
```
前端发送: state="evaluate_and_save", input="评价答案\n\n用户的回答：xxx"
  ↓
Thread["evaluate_and_save"]
  ├─ 生成评价数据
  ├─ 调用 multi_call([forward_result, save_doc])
  ├─ 判断: remaining_questions > 1
  ├─ nextState="get_question"
  ↓
Thread["get_question"]
  ├─ 从 InterviewHistory 提取已问题目ID
  ├─ 调用 kbase_rag → nextState="send_to_user"
  ↓
Thread["send_to_user"]
  ├─ 调用 forward_result(题目) → nextState=""
  ↓
完成，返回题目给前端
```

### 流程3：评价答案并生成总结（最后一题）
```
前端发送: state="evaluate_and_save", input="评价答案\n\n用户的回答：xxx"
  ↓
Thread["evaluate_and_save"]
  ├─ 生成评价数据
  ├─ 调用 multi_call([forward_result, save_doc])
  ├─ 判断: remaining_questions == 1
  ├─ nextState="summary_and_save"
  ↓
Thread["summary_and_save"]
  ├─ 生成总结 JSON
  ├─ 调用 multi_call([forward_result, save_doc])
  ├─ nextState=""
  ↓
完成，返回总结给前端
```

### 流程4：提前结束面试
```
前端发送: state="summary_and_save", input="结束面试"
  ↓
Thread["summary_and_save"]
  ├─ 从 InterviewHistory 读取历史记录
  ├─ 生成总结 JSON
  ├─ 调用 multi_call([forward_result, save_doc])
  ├─ nextState=""
  ↓
完成，返回总结给前端
```

## 五、关键设计要点

### 5.1 state 参数是必需的

- **前端必须传递**：所有请求都必须包含 `state` 参数
- **后端验证**：如果 `state` 为空或无效，返回错误
- **无兜底逻辑**：不保留 Main Thread 作为兜底

### 5.2 前端状态映射

前端需要维护一个简单的映射表：
```javascript
const STATE_MAP = {
    "开始面试": "get_question",
    "评价答案": "evaluate_and_save",
    "结束面试": "summary_and_save"
};
```

### 5.3 Orchestration.Main 字段处理

- **保留字段**：`Orchestration.Main` 字段保留（为了兼容结构体定义）
- **设置为 nil**：在创建 `BizOrchestration` 时，`Main` 设置为 `nil`
- **Orchestrator 逻辑**：如果 `state != ""`，直接使用 `Threads[state]`，不再检查 `Main`

## 六、文件变更清单

### 6.1 需要修改的文件

1. **`internal/test/interview/interview.html`**
   - 修改 `streamRequest` 函数，添加 `state` 参数
   - 修改所有调用 `streamRequest` 的地方，传递对应的 `state`

2. **`internal/test/interview/interview_test.go`**
   - 删除 `cfgIDMain` 相关代码
   - 删除 Main Thread 的 `InvocationConfig` 和 `InvocationConfigVersion` 创建
   - 删除 `raw_output` 函数注册
   - 修改 `BizOrchestration` 配置，`Main` 设置为 `nil`
   - 删除 `system_prompt_main.md` 和 `user_prompt_main.md` 的 embed

3. **`internal/service/orchestrator/orchestrator.go`**（不需要修改）
   - 保持现有逻辑不变，已经支持通过 `state` 参数跳过 Main Thread

### 6.2 需要删除的文件

- `internal/test/interview/system_prompt_main.md`
- `internal/test/interview/user_prompt_main.md`

### 6.3 保留的文件（不变）

- `system_prompt_get_question.md`
- `system_prompt_evaluate_save.md`
- `system_prompt_summary_save.md`
- `system_prompt_send.md`
- `user_prompt_get_question.md`
- `user_prompt_evaluate_save.md`
- `user_prompt_summary_save.md`
- `user_prompt_send.md`

## 七、预期效果

### 7.1 性能提升

- **响应时间**：从 30 秒减少到 20 秒左右（节省 Main Thread 的 LLM 调用时间）
- **LLM 调用次数**：每次请求减少 1 次 LLM 调用

### 7.2 架构简化

- **Thread 数量**：从 5 个减少到 4 个
- **文件数量**：从 10 个减少到 8 个
- **代码复杂度**：减少 Main Thread 的路由逻辑

### 7.3 维护性提升

- **前端控制**：前端明确指定目标状态，逻辑更清晰
- **减少依赖**：不再依赖 LLM 进行路由决策
- **错误处理**：前端传递错误的 `state` 会立即报错，便于调试

## 八、风险评估

### 8.1 潜在问题

1. **前端必须传递 state**：如果前端忘记传递或传递错误，会导致请求失败
   - **缓解**：前端代码中明确定义 `STATE_MAP`，统一管理

2. **state 参数验证**：后端需要验证 `state` 是否有效
   - **缓解**：Orchestrator 中已经有验证逻辑，如果 `state` 不在 `Threads` 中，会返回错误

3. **向后兼容性**：如果其他业务还在使用 Main Thread，可能会有影响
   - **缓解**：这是 interview 测试代码的修改，不影响其他业务

### 8.2 回滚方案

如果出现问题，可以：
1. 恢复 Main Thread 的配置
2. 前端恢复不传递 `state` 参数的版本
3. 使用 git 回滚到重构前的版本

## 九、实施顺序

1. ✅ 修改前端代码（添加 `state` 参数）
2. ✅ 修改 HTTP 代理（传递 `state` 参数）
3. ✅ 修改测试代码（删除 Main Thread 配置）
4. ✅ 删除 Main Thread 相关文件
5. ✅ 运行测试验证
6. ✅ 性能对比测试

---

**文档版本**：v2.0  
**最后更新**：2025-01-XX  
**作者**：AI Assistant

