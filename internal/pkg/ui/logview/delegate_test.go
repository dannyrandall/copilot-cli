package logview

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	t.Run("", func(t *testing.T) {
		l := Log{
			Log:       "haha",
			Timestamp: time.Now(),
		}
		fmt.Println(l.FilterValue())

	})

	t.Run("", func(t *testing.T) {
		r := FilterLog("Get", []string{
			"{\"time\":\"2022-07-20T17:17:19.660198-07:03\",\"log\":\"Geteeeeeee\"}",
			"{\"time\":\"2022-07-20T17:17:19.660198-07:01\",\"log\":\"Get SSS\"}",
			"{\"time\":\"2022-07-20T17:17:19.660198-07:02\",\"log\":\"haha\"}",
			"{\"time\":\"2022-07-20T17:17:19.660198-07:02\",\"log\":\"hahaGettttt\"}",
		})
		fmt.Println(r)
		require.Equal(t, true, false)
	})
}
