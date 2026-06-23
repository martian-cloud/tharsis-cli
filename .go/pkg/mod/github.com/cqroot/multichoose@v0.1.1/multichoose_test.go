package multichoose_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cqroot/multichoose"
)

func TestMultiChoose(t *testing.T) {
	mc := multichoose.New(100)
	require.Equal(t, 100, mc.Length())

	for i := 0; i < mc.Length(); i++ {
		require.Equal(t, false, mc.IsSelected(i))
	}

	for i := 0; i < mc.Length(); i++ {
		mc.Select(i)
		require.Equal(t, true, mc.IsSelected(i))
	}

	for i := 0; i < mc.Length(); i++ {
		mc.Deselect(i)
		require.Equal(t, false, mc.IsSelected(i))
	}

	mc.SetLimit(50)

	for i := 0; i < 50; i++ {
		mc.Select(i)
		require.Equal(t, true, mc.IsSelected(i))
	}
	for i := 50; i < mc.Length(); i++ {
		mc.Select(i)
		require.Equal(t, false, mc.IsSelected(i))
	}

	for i := 0; i < mc.Length(); i++ {
		mc.Deselect(i)
		require.Equal(t, false, mc.IsSelected(i))
	}
}

func TestMultiChooseToggle(t *testing.T) {
	mc := multichoose.New(100)

	for i := 0; i < mc.Length(); i++ {
		mc.Toggle(i)
		require.Equal(t, true, mc.IsSelected(i))
		mc.Toggle(i)
		require.Equal(t, false, mc.IsSelected(i))
	}
}
