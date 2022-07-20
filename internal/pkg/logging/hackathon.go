package logging

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/aws/cloudwatchlogs"
	"github.com/aws/copilot-cli/internal/pkg/ui/logview"
)

func (s *WorkloadClient) Query(query string) []logview.Log {
	logEventsOpts := cloudwatchlogs.LogEventsOpts{
		LogGroup: s.logGroupName,
		Limit:    aws.Int64(10000),
	}

	events, err := s.eventsGetter.LogEventsQuery(logEventsOpts, query)
	if err != nil {
		panic(err)
	}

	logs := make([]logview.Log, len(events.Events))
	for i := range events.Events {
		// TODO convert timestamp
		logs[i] = logview.Log{
			Log: events.Events[i].Message,
		}
	}

	return logs
}

func (s *WorkloadClient) StreamLogs(opts WriteLogEventsOpts, done chan struct{}) (chan logview.Log, chan error) {
	logEventsOpts := cloudwatchlogs.LogEventsOpts{
		LogGroup:            s.logGroupName,
		Limit:               opts.limit(),
		StartTime:           opts.startTime(s.now),
		EndTime:             opts.EndTime,
		StreamLastEventTime: nil,
		LogStreamLimit:      opts.LogStreamLimit,
	}

	errs := make(chan error)
	logs := make(chan logview.Log)

	go func() {
		defer close(errs)
		defer close(logs)

		for {
			events, err := s.eventsGetter.LogEvents(logEventsOpts)
			if err != nil {
				select {
				case errs <- fmt.Errorf("get log events for log group %s: %w", s.logGroupName, err):
				case <-done:
				}
				return
			}

			for i := range events.Events {
				// TODO convert timestamp
				select {
				case logs <- logview.Log{
					Log: events.Events[i].Message,
				}:
				case <-done:
					return
				}
			}

			select {
			case <-time.After(cloudwatchlogs.SleepDuration):
			case <-done:
				return
			}
			logEventsOpts.StreamLastEventTime = events.StreamLastEventTime
		}
	}()

	return logs, errs
}
