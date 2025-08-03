package simpleexample

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"ahs/internal/service"
)

var (
	openaiAPIKey  string
	openaiModel   string
	openaiBaseURL string

	input string
)

type SimpleProcessor struct{}

func (p *SimpleProcessor) Process(ctx context.Context, input string) (string, error) {
	s, err := gen()
	if err != nil {
		return "", err
	}
	return s.Content, nil
}
func (p *SimpleProcessor) ProcessStream(
	ctx context.Context, input string, callback service.StreamCallback) error {
	//TODO:
	return nil
}

func init() {
	openaiAPIKey = ""
	openaiModel = ""
	openaiBaseURL = ""

	input = ""
}

func newChatModel(ctx context.Context) model.ToolCallingChatModel {
	var cm model.ToolCallingChatModel
	var err error

	cm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  openaiAPIKey,
		BaseURL: openaiBaseURL,
		Model:   openaiModel,
	})
	if err != nil {
		log.Fatal(err)
	}
	return cm
}

func gen() (*schema.Message, error) {
	ctx := context.Background()
	messages := []*schema.Message{{Role: schema.User, Content: input}}
	cm := newChatModel(ctx)
	ret, err := cm.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("Generate failed, err=%v", err)
	}
	return ret, err
}
