def test_unique_int_1d(values):
    array = np.array(values, dtype=int)
    assert_equal(unique_int_1d(array), np.unique(array))
