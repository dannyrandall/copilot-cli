package cloudwatchlogs

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// LogEvents returns an array of Cloudwatch Logs events.
func (c *CloudWatchLogs) LogEventsQuery(opts LogEventsOpts, query string) (*LogEventsOutput, error) {
	var events []*Event
	in := initGetLogEventsInput(opts)
	realIn := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  in.LogGroupName,
		StartTime:     in.StartTime,
		EndTime:       in.EndTime,
		Limit:         in.Limit,
		FilterPattern: aws.String(query),
	}

	for {
		resp, err := c.client.FilterLogEvents(realIn)
		if err != nil {
			return nil, fmt.Errorf("get log events of %s/*: %w", opts.LogGroup, err)
		}

		for _, event := range resp.Events {
			log := &Event{
				IngestionTime: aws.Int64Value(event.IngestionTime),
				Message:       aws.StringValue(event.Message),
				Timestamp:     aws.Int64Value(event.Timestamp),
			}
			events = append(events, log)
		}

		if resp.NextToken == nil {
			break
		}

		realIn.NextToken = resp.NextToken
	}

	sort.SliceStable(events, func(i, j int) bool { return events[i].Timestamp < events[j].Timestamp })
	limit := int(aws.Int64Value(in.Limit))
	if limit != 0 {
		return &LogEventsOutput{
			Events: truncateEvents(limit, events),
		}, nil
	}
	return &LogEventsOutput{
		Events: events,
	}, nil
}
