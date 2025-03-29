# 网关

## 接口

### 流式接口
目前来说，当下的大模型在执行任务的时候，所需的时间都是非常长的，所以为了保障用户体验，接口都被设计为流式接口。也就是大模型一边返回数据，系统也一边返回数据给用户。

流式接口

### 同步调用

### 异步调用

### 同步批量调用

### 异步批量调用

## 异步请求处理

### 异步请求重试机制

### 获得异步请求结果
AI 网关需要考虑同步提供：
- 同步轮询接口
- 回调机制


## 多服务商接入

### 异构 API 兼容
正常来说，不同的业务方可能需要使用不同的大模型，而这些大模型分散在不同的服务商上，所以需要支持不同的云服务商提供的大模型接口。

例如说，AI 网关应该提供调用百度千帆、阿里云百炼的功能。


# 管理后台

## Prompt 管理
如果要求每个业务方自己去管理自己的 Prompt 是一个很麻烦的事情。所以在这里我们将提供统一的管理功能。

Prompt 管理包括：
- Prompt 本身的增删改查；
- Prompt 的版本控制、发布、回滚；
- Prompt 的审核功能。但是目前来说审核功能并非必须；
- Prompt 环境控制功能。也就是说不管是开发环境、还是测试环境等都是通过同一个管理后台来进行管理的；

目前暂时不支持组织结构，所以所有的 Prompt 都被划归到个人名下。


### Prompt 创建
用户在最开始的时候需要创建一个 Prompt，并且填入基本的信息。这些信息包括：
- name: 这个 prompt 的关键点
- description: 描述这个 Prompt 是用来干什么的

当 Prompt 创建之后，就可以开始第一个版本的创建。

> 具体的分析流程

### 需求文档
1. 创建阶段

场景：运营人员需要为电商客服机器人创建一个新Prompt

操作流程：

（1）填写prompt的基本信息：
   - name: prompt 的名称
   - description: 描述这个 Prompt 是用来干什么的
   - category: 分类
   - biz: 业务方
   - content： prompt 模板，如下：
```
您好！关于{商品名称}的信息如下：  
价格：{价格}元  
库存：{库存状态}  
支持服务：{服务列表}  
需要了解更多吗？  
```
 
（2）点击保存或提交

系统行为：

（1）生成首个版本号，如`v1.0.0`，记录状态，如草稿`draft`、待审核`review`等

（2）记录创建人、业务、时间戳和环境标签（开发环境`dev`）

2. 审核阶段

场景：产品经理/主管审核草稿后要求优化Prompt

操作流程：

（1）在版本列表中找到 `v1.0.0` 点击【编辑】

（2）修改 prompt 内容

（3）填写修改说明

（4）提交保存

系统行为：

（1）生成新版本`v1.1.0`，保留旧版本`v1.0.0`

（2）更新最后修改人和时间戳

3. 发布流程

场景：将修改后的Prompt部署到测试环境验证

操作流程：

（1）在版本 `v1.1.0` 点击【发布】

（2）选择目标环境：测试环境

（3）系统触发自动检查：
   - 语法校验
   - 变量完整性检查（确保{价格}等变量存在对应数据源）

（4）确认发布

系统行为：

（1）生成正式版本号 `v1.1.0`，修改状态为已发布`release`

（2）记录发布时间和环境标记（测试环境`test`）

（3）同步更新测试环境的 Prompt

4. 多环境推广

场景：测试通过后需上线至生产环境

操作流程：

（1）在已发布的测试环境版本 `v1.1.0` 点击【跨环境发布】

（2）选择目标环境：生产环境

（3）输入说明："正式上线新版友好话术"

（4）进入上线审批流程

系统行为： 审批通过后，生产环境版本更新为 `v1.1.0`

5. 异常回滚

场景：用户反馈新Prompt中的表情符号导致部分设备显示乱码
紧急操作：

（1）在生产环境版本历史中选中稳定版本 `v1.0.0`

（1）点击【紧急回滚】并选择回滚范围：仅生产环境

（3）填写事故原因："特殊符号兼容性问题"

系统行为：

（1）自动将生产环境恢复至 `v1.0.0`

（2）生成新版本 `v1.1.1` 记录回滚操作，标记状态为`rollback`，记录回滚前的版本信息

（3）向相关开发人员发送告警通知

5. 生命周期可视化
场景：用户想要查看 一个 prompt 的全生命周期。

操作：点击查看某个 prompt 的生命周期

系统行为：展示 prompt 的生命周期


#### 表结构设计
表1：prompt（Prompt元数据表）

| 字段名         | 类型           | 注释     | 约束       |
|-------------|--------------|--------|----------|
| id          | bigint       | 主键     |          |
| biz         | varchar(64)  | 唯一业务标识 | NOT NILL |
| name        | varchar(128) | 显示名称   | NOT NULL |
| category	   | varchar(64)  | 分类路径   |
| description | text         | 描述     |
| creator     | varchar(64)  | 创建人	   | NOT NULL |
| create_time | datetime     | 创建时间   | NOT NULL |
| update_time | datetime     | 创建时间   | NOT NULL |


```索引
INDEX idx_key (biz)
INDEX idx_creator(creator)
```

SQL
```
# 按业务查询
SELECT * FROM table WHERE biz = $biz

# 查询个人名下的prompt
SELECT * FROM table WHERE creator = $creator
```

表2：prompt_version（版本内容表）

| 字段名         | 类型           | 注释                                     | 约束          |
|-------------|--------------|----------------------------------------|-------------|
| id          | bigint       | 主键                                     |             |
| prompt_id   | bigint	      | 关联prompt.id                            | FK,NOT NULL |
| version     | varchar(32)  | 语义化版本号(如v1.2.3)	                       | NOT NULL    |
| parent_id   | bigint       | 上一个版本，关联 prompt_version.id             | FK,NOT NULL |
| variables   | json         | 提取的变量列表                                |             |
| state       | tinyint      | 1=draft 2=review 3=release 4- rollback | NOT NUL     |
| modify_by   | varchar(64)  | 最后修改人                                  | NOT NULL    |
| remark      | varchar(255) | 发布说明                                   |             |
| create_time | datetime     | 创建时间	                                  | NOT NULL    |
| update_time | datetime     | 创建时间                                   | NOT NULL    |

索引
```
INDEX idx_prompt_id_vresion_state(prompt_id,version,state)
```

SQL
```
# 查询 prompt 的版本和状态
SELECT version, state FOMR table WHRER prompt_id = $id

# 查询 prompt 的上一个版本
SELECT * FROM table WHRER id = $parent_id
```

表3：env_publish（环境发布记录表）

| 字段名          | 类型           | 注释                          | 约束           |
|--------------|--------------|-----------------------------|--------------|
| id           | bigint       | 主键                          |              |
| prompt_id    | bigint       | 关联prompt.id                 | 	FK,NOT NULL |
| version_id   | bigint       | 当前生效版本，关联prompt_version.id	 | FK,NOT NULL  |
| env          | varchar(16)  | 环境标识(prod/test等)            | 	NOT NULL    |
| publish_time | datetime     | 发布时间                        | 	NOT NULL    |
| operator     | varchar(64)  | 操作人                         | 	NOT NULL    |
| remark       | varchar(255) | 发布说明                        |              |
| create_time  | datetime     | 创建时间	                       | NOT NULL     |
| update_time  | datetime     | 创建时间                        | NOT NULL     |

索引
```
INDEX idx_env(env)
```

SQL
```
# 查看当前环境的prompt
SEKECT prompt_id,version_id FROM table WHERE env = $env ORDER BY id DESC LIMIT 1; 
```

表4：approval_flow（审批流水表）


## 业务接入管理

## 服务商管理

### 服务商

### 大模型

### 服务商密钥

## 用户管理

需要支持用户的注册、登录功能。

> 后期考虑通过管理员来开通具体权限。

## 组织管理

## 权限控制

# 非功能性
> 直接让 DeepSeek 生成

# 名词
## 开发环境
开发环境默认的情况下是 dev, test, production，业务方在启动这个平台的时候，可以修改配置文件支持更加多的环境。
