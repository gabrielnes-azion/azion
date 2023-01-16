package list

import (
	"github.com/aziontech/azion-cli/pkg/httpmock"
	"github.com/aziontech/azion-cli/pkg/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestList(t *testing.T) {
	t.Run("list page 1", func(t *testing.T) {
		mock := &httpmock.Registry{}

		mock.Register(
			httpmock.REST("GET", "edge_applications"),
			httpmock.JSONFromFile("./fixtures/applications.json"),
		)

		f, _, _ := testutils.NewFactory(mock)
		cmd := NewCmd(f)

		cmd.SetArgs([]string{"--page", "1"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)
	})
}
