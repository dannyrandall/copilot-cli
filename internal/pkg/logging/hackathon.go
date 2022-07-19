package logging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/aws/cloudwatchlogs"
	"github.com/aws/copilot-cli/internal/pkg/ui/logview"
)

func (s *WorkloadClient) Query(query string) logview.QueryResult {
	logEventsOpts := cloudwatchlogs.LogEventsOpts{
		LogGroup: s.logGroupName,
		Limit:    aws.Int64(10000),
	}

	events, err := s.eventsGetter.LogEventsQuery(logEventsOpts, query)
	if err != nil {
		panic(err)
	}

	logs := make(logview.QueryResult, len(events.Events))
	for i := range events.Events {
		logs[i] = logview.Log{
			Log: events.Events[i].Message,
		}
	}

	return logs
}
