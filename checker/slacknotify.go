package checker

import (
	"fmt"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type SlackNotifier struct {
	api     *slack.Client
	channel string
	logger  *zap.SugaredLogger
}

func NewSlackNotifier(token string, channel string, logger *zap.SugaredLogger) *SlackNotifier {
	return &SlackNotifier{
		api:     slack.New(token),
		channel: channel,
		logger:  logger,
	}
}

func (s *SlackNotifier) NotifyError(chall Challenge, result TestResult, stdout string, errlog string) error {
	args := slack.PostMessageParameters{
		Username:  "TSGCTF Status",
		IconEmoji: ":fire:",
		Markdown:  true,
	}
	stdout = fmt.Sprintf("```\n%s\n```", stdout)
	errlog = fmt.Sprintf("```\n%s\n```", errlog)
	msg := fmt.Sprintf("Status check failed for `%s`\n"+"Result: `%s`\n"+"Asignee: <@%s>\nSTDOUT:\n%sSTDERR:\n%s", chall.Name, result.ToMessage(), chall.Assignee, stdout, errlog)

	_, _, err := s.api.PostMessage(s.channel, slack.MsgOptionText(msg, false), slack.MsgOptionPostMessageParameters(args))
	if err != nil {
		s.logger.Errorf("Failed to send slack message: %s", err)
		return err
	}

	return nil
}
