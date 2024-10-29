package hypertext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer_Index(t *testing.T) {
	t.Run("it returns a name", func(t *testing.T) {
		server := Server{}
		ctx := context.Background()
		data := server.Index(ctx)
		require.NotZero(t, data.Name)
	})
}
