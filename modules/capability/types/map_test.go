package types_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/ibc-go/modules/capability/types"
)

func TestCapabilityMap_GetOrSet(t *testing.T) {
	m := types.NewCapabilityMap()

	_, ok := m.Get(1)
	require.False(t, ok)

	first := types.NewCapability(1)
	require.Same(t, first, m.GetOrSet(1, first))

	// A later creation at the same index adopts the existing pointer.
	second := types.NewCapability(1)
	require.Same(t, first, m.GetOrSet(1, second))

	got, ok := m.Get(1)
	require.True(t, ok)
	require.Same(t, first, got)
}

func TestCapabilityMap_SetDelete(t *testing.T) {
	m := types.NewCapabilityMap()

	first := types.NewCapability(1)
	m.Set(1, first)

	// Set overwrites unconditionally; used when rebuilding from committed state.
	second := types.NewCapability(1)
	m.Set(1, second)
	got, ok := m.Get(1)
	require.True(t, ok)
	require.Same(t, second, got)

	m.Delete(1)
	_, ok = m.Get(1)
	require.False(t, ok)

	m.Delete(1) // deleting a missing index is a no-op
}

// Run with -race: concurrent creators of the same index must all end up with
// one pointer and must not trip the map's concurrent-access detector.
func TestCapabilityMap_ConcurrentGetOrSet(t *testing.T) {
	m := types.NewCapabilityMap()

	const goroutines = 32
	results := make([]*types.Capability, goroutines)

	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = m.GetOrSet(7, types.NewCapability(7))
			m.Get(7)
		}(i)
	}
	wg.Wait()

	winner, ok := m.Get(7)
	require.True(t, ok)
	for _, cap := range results {
		require.Same(t, winner, cap)
	}
}
