package ragtool

import (
	"ahs/internal/handler"
	"ahs/internal/service/rag"
	"context"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// __测试目标__
// 1) 参数注入测试：验证 WithUnmarshalArguments 自定义反序列化逻辑会从上下文注入 user_id / archive_id。
// 2) 工具接口测试：验证工具符合 Eino InvokableTool 接口规范（Info(ctx)、InvokableRun）。
// 3) 错误处理测试：验证 WrapToolWithErrorHandler 将错误转换为字符串结果并不返回 error。

func TestMemoryQuery_Interface_Info(t *testing.T) {
	// 获取工具并断言接口
	qTool, err := GetMemoryQueryTool()
	require.NoError(t, err)
	require.NotNil(t, qTool)
	_, ok := any(qTool).(tool.InvokableTool)
	assert.True(t, ok, "should implement InvokableTool")

	// 校验 Info(ctx) 返回
	ti, err := qTool.Info(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ti)
	assert.Equal(t, "memory_query", ti.Name)
	assert.Contains(t, ti.Desc, "查询记忆")
}

func TestMemoryQuery_WithUnmarshalArguments_Injection(t *testing.T) {
	// 确保默认 Manager 初始化
	m := rag.Default()

	// 预置多租户数据
	tenantA := rag.Tenant{UserID: "u_inject", ArchiveID: "a_inject"}
	tenantB := rag.Tenant{UserID: "u_other", ArchiveID: "a_other"}

	now := time.Now()
	// 仅 tenantA 有两条；tenantB 有一条
	items := []rag.MemoryItem{
		{Tenant: tenantA, Content: "A-1", Tags: []string{"t1"}, Kind: rag.KindNote, CreatedAt: now},
		{Tenant: tenantA, Content: "A-2", Tags: []string{"t2"}, Kind: rag.KindFact, CreatedAt: now},
		{Tenant: tenantB, Content: "B-1", Tags: []string{"t1"}, Kind: rag.KindNote, CreatedAt: now},
	}
	for _, it := range items {
		require.NoError(t, m.Save(context.Background(), it, rag.SaveOptions{ToMemory: true}))
	}

	// 构造包含租户信息的上下文（模拟 HTTP 原始 body 注入）
	reqBody := map[string]any{
		"user_id":    tenantA.UserID,
		"archive_id": tenantA.ArchiveID,
	}
	bodyBytes, _ := sonic.Marshal(reqBody)
	ctx := handler.WithRequestBody(context.Background(), bodyBytes)

	// agent 传参不包含租户信息，由 WithUnmarshalArguments 注入
	args := map[string]any{
		"query": "", // 仅按租户过滤
		"top_k": 10,
	}
	argsStr, _ := sonic.MarshalString(args)

	qTool, err := GetMemoryQueryTool()
	require.NoError(t, err)

	// 通过 InvokableRun 触发自定义解包与注入
	result, err := qTool.InvokableRun(ctx, argsStr)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var out MemoryQueryOutput
	require.NoError(t, sonic.UnmarshalString(result, &out))
	assert.True(t, out.Success)
	assert.Equal(t, 2, out.Count)
	// 返回的内容仅应来自 tenantA
	for _, it := range out.Items {
		assert.Contains(t, []string{"A-1", "A-2"}, it.Content)
	}
}

func TestMemoryQuery_ErrorHandling_UnmarshalFail(t *testing.T) {
	qTool, err := GetMemoryQueryTool()
	require.NoError(t, err)

	// 构造带租户的上下文，但参数是非法 JSON，触发 WithUnmarshalArguments 的解析错误
	reqBody := map[string]any{"user_id": "u1", "archive_id": "a1"}
	bodyBytes, _ := sonic.Marshal(reqBody)
	ctx := handler.WithRequestBody(context.Background(), bodyBytes)

	// 非法 JSON
	result, err := qTool.InvokableRun(ctx, "{invalid json}")
	// 由于 WrapToolWithErrorHandler，err 应为 nil，结果包含友好消息
	assert.NoError(t, err)
	assert.Contains(t, result, "记忆查询失败")
}

func TestMemoryQuery_ErrorHandling_MissingTenant(t *testing.T) {
	qTool, err := GetMemoryQueryTool()
	require.NoError(t, err)

	// 无请求体的上下文，parseTenantFromContext 失败
	args := map[string]any{"query": "hello"}
	argsStr, _ := sonic.MarshalString(args)

	result, err := qTool.InvokableRun(context.Background(), argsStr)
	assert.NoError(t, err)
	assert.Contains(t, result, "记忆查询失败")
}

func TestMemorySave_Interface_Info_And_ErrorHandling(t *testing.T) {
	sTool, err := GetMemorySaveTool()
	require.NoError(t, err)
	require.NotNil(t, sTool)

	// 接口与 Info(ctx)
	_, ok := any(sTool).(tool.InvokableTool)
	assert.True(t, ok)
	ti, err := sTool.Info(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ti)
	assert.Equal(t, "memory_save", ti.Name)
	assert.Contains(t, ti.Desc, "保存记忆")

	// 错误处理：缺少上下文租户，参数虽合法但应在注入阶段失败
	agentArgs := map[string]any{
		"content":     "test",
		"tags":        []string{"t"},
		"kind":        "note",
		"ttl_seconds": 0,
	}
	argsStr, _ := sonic.MarshalString(agentArgs)
	result, err := sTool.InvokableRun(context.Background(), argsStr)
	assert.NoError(t, err)
	assert.Contains(t, result, "记忆保存失败")
}

func TestMemorySave_WithUnmarshalArguments_Injection(t *testing.T) {
	// 构造包含租户信息的上下文
	reqBody := map[string]any{"user_id": "u_save", "archive_id": "a_save"}
	bodyBytes, _ := sonic.Marshal(reqBody)
	ctx := handler.WithRequestBody(context.Background(), bodyBytes)

	sTool, err := GetMemorySaveTool()
	require.NoError(t, err)

	// agent 仅提供业务字段
	agentArgs := map[string]any{
		"content":     "需要被保存的记忆",
		"tags":        []string{"k1", "k2"},
		"kind":        "note",
		"ttl_seconds": 0,
	}
	argsStr, _ := sonic.MarshalString(agentArgs)

	// 执行保存
	result, err := sTool.InvokableRun(ctx, argsStr)
	require.NoError(t, err)
	var out MemorySaveOutput
	require.NoError(t, sonic.UnmarshalString(result, &out))
	assert.True(t, out.Success)
	assert.Contains(t, out.Message, "记忆保存成功")
}
