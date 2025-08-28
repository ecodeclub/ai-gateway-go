## functionCall说明

### 抽象说明
参考type.go下面的FunctionCall的定义


### 几个特殊的functionCall

### emit_json

这个functionCall，主要是提取用户的结构化的数据。要求这些数据需要放在 data字段里

### invoke_llm

这个functionCall，目的是进一步发起一个LLM调用。接收一个InvocationConfig.ID,以及一个叫做gjson_expr
的参数进行二次渲染。
