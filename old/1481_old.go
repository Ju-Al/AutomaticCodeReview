	for i := 0; i <= 255; i += 2 {
		a.SetBit(byte(i))
		b.SetBit(byte(256 - i))
	}
	require.Equal(t, a, b)

	for i := 0; i <= 255; i += 2 {
		require.NotZero(t, a.Bit(byte(i)))
	}

	for i := 1; i <= 255; i += 2 {
		require.Zero(t, a.Bit(byte(i)))
	}

	// clear the bits at different order, and testing that the bits were cleared correctly.
	for i := 0; i <= 255; i += 32 {
		a.ClearBit(byte(i))
		b.ClearBit(byte(256 - i))
	}
	require.Equal(t, a, b)

	for i := 0; i <= 255; i += 32 {
		require.Zero(t, a.Bit(byte(i)))
	}

	// clear all bits ( some would get cleared more than once )
	for i := 0; i <= 255; i += 2 {
		a.ClearBit(byte(i))
	}

	// check that the bitset is zero.
	require.True(t, a.IsZero())
}
