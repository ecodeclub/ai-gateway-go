#!/bin/bash

# 检查API_KEY是否设置
if [ -z "$API_KEY" ]; then
    echo "❌ 错误: 未设置 API_KEY 环境变量"
    echo "请先执行: export API_KEY='你的百炼API密钥'"
    exit 1
fi

echo "✅ API_KEY已设置"
echo ""

# 测试音频URL（使用您的COS音频文件）
AUDIO_URL="https://webook-1314583317.cos.ap-nanjing.myqcloud.com/audio-temp/1761557566_audio.webm"

echo "📤 步骤1: 提交语音识别任务（paraformer-v2）..."
echo "音频URL: $AUDIO_URL"
echo ""

# 提交任务
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "https://dashscope.aliyuncs.com/api/v1/services/audio/asr/transcription" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "X-DashScope-Async: enable" \
  -d "{
    \"model\": \"paraformer-v2\",
    \"input\": {
      \"file_urls\": [\"${AUDIO_URL}\"]
    }
  }")

# 分离HTTP状态码和响应体
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
RESPONSE_BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE:/d')

echo "HTTP状态码: $HTTP_CODE"
echo "响应体:"
echo "$RESPONSE_BODY" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE_BODY"
echo ""

if [ "$HTTP_CODE" != "200" ]; then
    echo "❌ 提交任务失败，HTTP状态码: $HTTP_CODE"
    exit 1
fi

# 提取 task_id
TASK_ID=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('output', {}).get('task_id', ''))" 2>/dev/null)

if [ -z "$TASK_ID" ]; then
    echo "❌ 提交任务失败，无法获取task_id"
    echo "完整响应: $RESPONSE_BODY"
    exit 1
fi

echo "✅ 任务已提交，task_id: $TASK_ID"
echo ""
echo "📊 步骤2: 轮询任务状态（每2秒一次，最多10次）..."
echo ""

# 轮询任务状态
for i in {1..10}; do
    sleep 2
    
    STATUS_RESPONSE=$(curl -s -X GET "https://dashscope.aliyuncs.com/api/v1/tasks/${TASK_ID}" \
      -H "Authorization: Bearer ${API_KEY}")
    
    TASK_STATUS=$(echo "$STATUS_RESPONSE" | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('output', {}).get('task_status', ''))" 2>/dev/null)
    
    echo "[$i/10] 任务状态: $TASK_STATUS"
    
    if [ "$TASK_STATUS" = "SUCCEEDED" ]; then
        echo ""
        echo "🎉 任务成功！"
        echo ""
        echo "完整响应:"
        echo "$STATUS_RESPONSE" | python3 -m json.tool
        echo ""
        
        # 提取转写结果（可能需要下载transcription_url）
        TRANSCRIPTION_URL=$(echo "$STATUS_RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
results = data.get('output', {}).get('results', [])
if results and len(results) > 0:
    print(results[0].get('transcription_url', ''))
" 2>/dev/null)
        
        if [ -n "$TRANSCRIPTION_URL" ]; then
            echo "🔄 下载转写结果: $TRANSCRIPTION_URL"
            TRANSCRIPTION=$(curl -s "$TRANSCRIPTION_URL" | python3 -c "
import sys, json
data = json.load(sys.stdin)
transcripts = data.get('transcripts', [])
if transcripts and len(transcripts) > 0:
    print(transcripts[0].get('text', ''))
" 2>/dev/null)
        else
            # 尝试直接获取 transcription 字段
            TRANSCRIPTION=$(echo "$STATUS_RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
results = data.get('output', {}).get('results', [])
if results and len(results) > 0:
    print(results[0].get('transcription', ''))
" 2>/dev/null)
        fi
        
        if [ -n "$TRANSCRIPTION" ]; then
            echo "📝 转写结果: $TRANSCRIPTION"
        fi
        
        exit 0
    elif [ "$TASK_STATUS" = "FAILED" ]; then
        echo ""
        echo "❌ 任务失败"
        echo "$STATUS_RESPONSE" | python3 -m json.tool
        exit 1
    fi
done

echo ""
echo "⏰ 超时：任务仍在运行（状态: $TASK_STATUS）"
echo "您可以手动继续查询: curl -H 'Authorization: Bearer \$API_KEY' https://dashscope.aliyuncs.com/api/v1/tasks/${TASK_ID}"
