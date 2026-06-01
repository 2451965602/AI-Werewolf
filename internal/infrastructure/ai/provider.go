package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"ai-werewolf-go/internal/domain"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

var firstNumber = regexp.MustCompile(`\d+`)

type EinoProvider struct {
	model model.BaseChatModel
}

func NewEinoProvider(model model.BaseChatModel) *EinoProvider {
	return &EinoProvider{model: model}
}

func (p *EinoProvider) Speak(player domain.Player, view domain.DecisionContext) (string, error) {
	message, err := p.generate(fmt.Sprintf("你是%d号%s，请用一句话发言。", player.ID, player.Name), view)
	if err != nil {
		return "", err
	}
	return message.Content, nil
}

func (p *EinoProvider) VoteTarget(player domain.Player, view domain.DecisionContext) (int, error) {
	message, err := p.generate(fmt.Sprintf("你是%d号%s，请只输出要投票的玩家编号。", player.ID, player.Name), view)
	if err != nil {
		return 0, err
	}
	return parseTargetID(message.Content)
}

func (p *EinoProvider) WerewolfTarget(player domain.Player, view domain.DecisionContext) (int, error) {
	message, err := p.generate(fmt.Sprintf("你是%d号狼人%s，请只输出今晚击杀的玩家编号。", player.ID, player.Name), view)
	if err != nil {
		return 0, err
	}
	return parseTargetID(message.Content)
}

func (p *EinoProvider) generate(instruction string, view domain.DecisionContext) (*schema.Message, error) {
	if p.model == nil {
		return nil, fmt.Errorf("eino chat model is nil")
	}
	messages := []*schema.Message{
		schema.SystemMessage("你是狼人杀 AI 玩家。只能基于可见信息回答，不得编造隐藏身份。"),
		schema.UserMessage(fmt.Sprintf("当前第%d轮，阶段%s。%s", view.Round, view.Phase, instruction)),
	}
	return p.model.Generate(context.Background(), messages)
}

func parseTargetID(content string) (int, error) {
	match := firstNumber.FindString(content)
	if match == "" {
		return 0, fmt.Errorf("model response does not contain target id: %q", content)
	}
	target, err := strconv.Atoi(match)
	if err != nil {
		return 0, fmt.Errorf("parse target id: %w", err)
	}
	return target, nil
}

type FallbackProvider struct{}

func (FallbackProvider) Speak(player domain.Player, _ domain.DecisionContext) (string, error) {
	return "大家好，我是" + player.Name + "。", nil
}

func (FallbackProvider) VoteTarget(_ domain.Player, context domain.DecisionContext) (int, error) {
	for _, player := range context.Players {
		if player.Alive {
			return player.ID, nil
		}
	}
	return 0, nil
}

func (FallbackProvider) WerewolfTarget(_ domain.Player, context domain.DecisionContext) (int, error) {
	for _, player := range context.Players {
		if player.Alive && player.Team != domain.TeamWolf {
			return player.ID, nil
		}
	}
	return 0, nil
}
