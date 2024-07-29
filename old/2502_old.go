		return nil, fmt.Errorf("exponent mapping requires scale <= 0")
	}
	if scale < MinScale {
		return nil, fmt.Errorf("scale too low")
	}
	return &prebuiltMappings[scale-MinScale], nil
}

// MapToIndex implements mapping.Mapping.
func (e *exponentMapping) MapToIndex(value float64) int32 {
	// Note: we can assume not a 0, Inf, or NaN; positive sign bit.

	// Note: bit-shifting does the right thing for negative
	// exponents, e.g., -1 >> 1 == -1.
	return getBase2(value) >> e.shift
}

func (e *exponentMapping) minIndex() int32 {
	return int32(MinNormalExponent) >> e.shift
}

func (e *exponentMapping) maxIndex() int32 {
	return int32(MaxNormalExponent) >> e.shift
}

// LowerBoundary implements mapping.Mapping.
func (e *exponentMapping) LowerBoundary(index int32) (float64, error) {
	if min := e.minIndex(); index < min {
		return 0, mapping.ErrUnderflow
	}

	if max := e.maxIndex(); index > max {
		return 0, mapping.ErrOverflow
	}

	unbiased := int64(index << e.shift)

	// Note: although the mapping function rounds subnormal values
	// up to the smallest normal value, there are still buckets
	// that may be filled that start at subnormal values.  The
	// following code handles this correctly.  It's equivalent to and
	// faster than math.Ldexp(1, int(unbiased)).
	if unbiased < int64(MinNormalExponent) {
		subnormal := uint64(1 << SignificandWidth)
		for unbiased < int64(MinNormalExponent) {
			unbiased++
			subnormal >>= 1
		}
		return math.Float64frombits(subnormal), nil
	}
	exponent := unbiased + ExponentBias

	bits := uint64(exponent << SignificandWidth)
	return math.Float64frombits(bits), nil
}

// Scale implements mapping.Mapping.
func (e *exponentMapping) Scale() int32 {
	return -int32(e.shift)
}
