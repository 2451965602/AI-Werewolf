package ai

import (
	"context"
	"encoding/json"
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

func (p *EinoProvider) SeerTarget(player domain.Player, view domain.DecisionContext) (int, error) {
	message, err := p.generate(fmt.Sprintf("你是%d号预言家%s，请只输出今晚查验的玩家编号。", player.ID, player.Name), view)
	if err != nil {
		return 0, err
	}
	return parseTargetID(message.Content)
}

func (p *EinoProvider) WitchAction(player domain.Player, view domain.DecisionContext) (domain.WitchAction, error) {
	message, err := p.generate(fmt.Sprintf("你是%d号女巫%s，请输出动作：heal <编号>、poison <编号> 或 none。", player.ID, player.Name), view)
	if err != nil {
		return domain.WitchAction{}, err
	}
	if regexp.MustCompile(`(?i)none`).FindString(message.Content) != "" {
		return domain.WitchAction{Type: domain.WitchActionNone}, nil
	}
	target, err := parseTargetID(message.Content)
	if err != nil {
		return domain.WitchAction{}, err
	}
	if regexp.MustCompile(`(?i)poison`).FindString(message.Content) != "" {
		return domain.WitchAction{Type: domain.WitchActionPoison, TargetID: target}, nil
	}
	return domain.WitchAction{Type: domain.WitchActionHeal, TargetID: target}, nil
}

func (p *EinoProvider) generate(instruction string, view domain.DecisionContext) (*schema.Message, error) {
	if p.model == nil {
		return nil, fmt.Errorf("eino chat model is nil")
	}
	visibleContext, err := json.Marshal(view)
	if err != nil {
		return nil, fmt.Errorf("marshal visible decision context: %w", err)
	}
	messages := []*schema.Message{
		schema.SystemMessage("你是狼人杀 AI 玩家。只能基于可见信息回答，不得编造隐藏身份。"),
		schema.UserMessage(fmt.Sprintf("可见上下文JSON：%s\n%s", string(visibleContext), instruction)),
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

func (FallbackProvider) SeerTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	for _, candidate := range context.Players {
		if candidate.Alive && candidate.ID != player.ID {
			return candidate.ID, nil
		}
	}
	return 0, nil
}

func (FallbackProvider) WitchAction(domain.Player, domain.DecisionContext) (domain.WitchAction, error) {
	return domain.WitchAction{Type: domain.WitchActionNone}, nil
}
