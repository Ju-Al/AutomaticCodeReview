
	mockMap.DeleteErr = nil
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(BeEmpty())
}

// TestCachingMap_CleanUp verifies cleaning up of a whole map.
func TestCachingMap_CleanUp(t *testing.T) {
	mockMap, cm := setupCachingMapTest(t)
	_ = mockMap.Update([]byte{1, 2}, []byte{1, 2, 3, 4})
	_ = mockMap.Update([]byte{1, 3}, []byte{1, 2, 4, 4})

	err := cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(BeEmpty())
}

// TestCachingMap_ApplyAll mainline test using separate Apply calls for adds and deletes.
func TestCachingMap_SplitUpdateAndDelete(t *testing.T) {
	mockMap, cm := setupCachingMapTest(t)
	mockMap.Contents = map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 4}),
		string([]byte{1, 3}): string([]byte{1, 2, 4, 4}),
	}

	cm.SetDesiredState([]byte{1, 1}, []byte{1, 2, 4, 3}) // Same value for existing key.
	cm.SetDesiredState([]byte{1, 2}, []byte{1, 2, 3, 6}) // New value for existing key.
	cm.SetDesiredState([]byte{1, 4}, []byte{1, 2, 3, 5}) // New K/V
	// Shouldn't do anything until we hit apply.
	Expect(mockMap.OpCount()).To(Equal(0))

	err := cm.ApplyUpdatesOnly()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(Equal(map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}), // No change
		string([]byte{1, 2}): string([]byte{1, 2, 3, 6}), // Updated
		string([]byte{1, 3}): string([]byte{1, 2, 4, 4}), // Not desired but should be left alone
		string([]byte{1, 4}): string([]byte{1, 2, 3, 5}), // Added
	}))
	// Two updates and an iteration to load the map initially.
	Expect(mockMap.UpdateCount).To(Equal(2))
	Expect(mockMap.DeleteCount).To(Equal(0))
	Expect(mockMap.GetCount).To(Equal(0))
	Expect(mockMap.IterCount).To(Equal(1))

	err = cm.ApplyDeletionsOnly()
	Expect(err).NotTo(HaveOccurred())

	Expect(mockMap.Contents).To(Equal(map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 6}),
		string([]byte{1, 4}): string([]byte{1, 2, 3, 5}),
	}))
	// No new updates or iterations but should get one extra deletion.
	Expect(mockMap.UpdateCount).To(Equal(2))
	Expect(mockMap.GetCount).To(Equal(0))
	Expect(mockMap.DeleteCount).To(Equal(1))
	Expect(mockMap.IterCount).To(Equal(1))

	// Doing an extra apply should make no changes.
	preApplyOpCount := mockMap.OpCount()
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.OpCount()).To(Equal(preApplyOpCount))
}

// TestCachingMap_ApplyAll mainline test using ApplyAll() to update the dataplane.
func TestCachingMap_ApplyAll(t *testing.T) {
	mockMap, cm := setupCachingMapTest(t)
	mockMap.Contents = map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 4}),
		string([]byte{1, 3}): string([]byte{1, 2, 4, 4}),
	}

	cm.SetDesiredState([]byte{1, 1}, []byte{1, 2, 4, 3}) // Same value for existing key.
	cm.SetDesiredState([]byte{1, 2}, []byte{1, 2, 3, 6}) // New value for existing key.
	cm.SetDesiredState([]byte{1, 4}, []byte{1, 2, 3, 5}) // New K/V
	// Shouldn't do anything until we hit apply.
	Expect(mockMap.OpCount()).To(Equal(0))

	err := cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(Equal(map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 6}),
		string([]byte{1, 4}): string([]byte{1, 2, 3, 5}),
	}))
	// Two updates and an iteration to load the map initially.
	Expect(mockMap.UpdateCount).To(Equal(2))
	Expect(mockMap.DeleteCount).To(Equal(1))
	Expect(mockMap.GetCount).To(Equal(0))
	Expect(mockMap.IterCount).To(Equal(1))

	// Doing an extra apply should make no changes.
	preApplyOpCount := mockMap.OpCount()
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.OpCount()).To(Equal(preApplyOpCount))

	// Finish with a DeleteAll()
	cm.DeleteAll()
	Expect(mockMap.OpCount()).To(Equal(preApplyOpCount)) // No immediate change
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(BeEmpty())
	Expect(mockMap.DeleteCount).To(Equal(4))

	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(BeEmpty())
	Expect(mockMap.DeleteCount).To(Equal(4))
}

// TestCachingMap_DeleteBeforeLoad does some set and delete calls before loading from
// the dataplane.
func TestCachingMap_DeleteBeforeLoad(t *testing.T) {
	mockMap, cm := setupCachingMapTest(t)
	mockMap.Contents = map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 4}),
		string([]byte{1, 3}): string([]byte{1, 2, 4, 4}),
	}

	cm.SetDesiredState([]byte{1, 1}, []byte{1, 2, 4, 3}) // Same value for existing key.
	cm.SetDesiredState([]byte{1, 2}, []byte{1, 2, 3, 6}) // New value for existing key.
	cm.SetDesiredState([]byte{1, 4}, []byte{1, 2, 3, 5}) // New K/V
	cm.DeleteDesiredState([]byte{1, 2})                  // Changed my mind.
	cm.DeleteDesiredState([]byte{1, 4})                  // Changed my mind.
	cm.DeleteDesiredState([]byte{1, 8})                  // Delete of non-existent key is a no-op.
	// Shouldn't do anything until we hit apply.
	Expect(mockMap.OpCount()).To(Equal(0))

	err := cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(Equal(map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
	}))
	// Just the two deletes.
	Expect(mockMap.UpdateCount).To(Equal(0))
	Expect(mockMap.DeleteCount).To(Equal(2))
	Expect(mockMap.GetCount).To(Equal(0))
	Expect(mockMap.IterCount).To(Equal(1))

	// Doing an extra apply should make no changes.
	preApplyOpCount := mockMap.OpCount()
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.OpCount()).To(Equal(preApplyOpCount))
}

// TestCachingMap_PreLoad verifies calling LoadCacheFromDataplane before setting values.
func TestCachingMap_PreLoad(t *testing.T) {
	mockMap, cm := setupCachingMapTest(t)
	mockMap.Contents = map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 4}),
		string([]byte{1, 3}): string([]byte{1, 2, 4, 4}),
	}
	err := cm.LoadCacheFromDataplane()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.IterCount).To(Equal(1))
	Expect(mockMap.OpCount()).To(Equal(1))

	// Check we can query the cache.
	Expect(cm.GetDataplaneCache([]byte{1, 1})).To(Equal([]byte{1, 2, 4, 3}))
	seenValues := make(map[string]string)
	cm.IterDataplaneCache(func(k, v []byte) {
		seenValues[string(k)] = string(v)
	})
	Expect(seenValues).To(Equal(mockMap.Contents))

	cm.SetDesiredState([]byte{1, 1}, []byte{1, 2, 4, 3}) // Same value for existing key.
	cm.SetDesiredState([]byte{1, 2}, []byte{1, 2, 3, 6}) // New value for existing key.
	cm.SetDesiredState([]byte{1, 4}, []byte{1, 2, 3, 5}) // New K/V
	cm.DeleteDesiredState([]byte{1, 8})                  // Delete of non-existent key is a no-op.

	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(Equal(map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 6}),
		string([]byte{1, 4}): string([]byte{1, 2, 3, 5}),
	}))
	// Two updates and an iteration to load the map initially.
	Expect(mockMap.UpdateCount).To(Equal(2))
	Expect(mockMap.DeleteCount).To(Equal(1))
	Expect(mockMap.GetCount).To(Equal(0))
	Expect(mockMap.IterCount).To(Equal(1))

	// Doing an extra apply should make no changes.
	preApplyOpCount := mockMap.OpCount()
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.OpCount()).To(Equal(preApplyOpCount))
}

// TestCachingMap_Resync verifies handling of a dataplane reload while there are pending
// changes.  Pending changes should be dropped if the reload finds that they've already
// been made.
func TestCachingMap_Resync(t *testing.T) {
	mockMap, cm := setupCachingMapTest(t)
	mockMap.Contents = map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 4}),
		string([]byte{1, 3}): string([]byte{1, 2, 4, 4}),
	}
	err := cm.LoadCacheFromDataplane()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.IterCount).To(Equal(1))
	Expect(mockMap.OpCount()).To(Equal(1))

	cm.SetDesiredState([]byte{1, 1}, []byte{1, 2, 4, 3}) // Same value for existing key.
	cm.SetDesiredState([]byte{1, 2}, []byte{1, 2, 3, 6}) // New value for existing key.
	cm.SetDesiredState([]byte{1, 4}, []byte{1, 2, 3, 5}) // New K/V

	// At this point we've got some updates and a deletion queued up. Change the contents
	// of the map:
	// - Remove the key that was already correct.
	// - Remove the key that we were about to delete.
	// - Correct the value of the other key.
	mockMap.Contents = map[string]string{
		string([]byte{1, 2}): string([]byte{1, 2, 3, 6}),
	}

	err = cm.LoadCacheFromDataplane()
	Expect(err).NotTo(HaveOccurred())

	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.Contents).To(Equal(map[string]string{
		string([]byte{1, 1}): string([]byte{1, 2, 4, 3}),
		string([]byte{1, 2}): string([]byte{1, 2, 3, 6}),
		string([]byte{1, 4}): string([]byte{1, 2, 3, 5}),
	}))
	// Two updates and an iteration to load the map initially.
	Expect(mockMap.UpdateCount).To(Equal(2))
	Expect(mockMap.DeleteCount).To(Equal(0))
	Expect(mockMap.GetCount).To(Equal(0))
	Expect(mockMap.IterCount).To(Equal(2))

	// Doing an extra apply should make no changes.
	preApplyOpCount := mockMap.OpCount()
	err = cm.ApplyAllChanges()
	Expect(err).NotTo(HaveOccurred())
	Expect(mockMap.OpCount()).To(Equal(preApplyOpCount))
}

func setupCachingMapTest(t *testing.T) (*mock.Map, *CachingMap) {
	RegisterTestingT(t)
	mockMap := mock.NewMockMap(cachingMapParams)
	cm := New(cachingMapParams, mockMap)
	return mockMap, cm
}

func TestByteArrayToByteArrayMap(t *testing.T) {
	RegisterTestingT(t)

	m := NewByteArrayToByteArrayMap(2, 4)

	Expect(m.Get([]byte{1, 2})).To(BeNil(), "New map should not contain a value")
	m.Set([]byte{1, 2}, []byte{1, 2, 3, 4})
	Expect(m.Get([]byte{1, 2})).To(Equal([]byte{1, 2, 3, 4}), "Map should contain set value")
	m.Set([]byte{1, 2}, []byte{1, 2, 3, 5})
	Expect(m.Get([]byte{1, 2})).To(Equal([]byte{1, 2, 3, 5}), "Map should record updates")
	m.Set([]byte{3, 4}, []byte{1, 2, 3, 6})
	Expect(m.Get([]byte{3, 4})).To(Equal([]byte{1, 2, 3, 6}), "Map should record updates")

	seenValues := map[string][]byte{}
	m.Iter(func(k, v []byte) {
		Expect(k).To(HaveLen(2))
		Expect(v).To(HaveLen(4))
		seenValues[string(k)] = v
	})
	Expect(seenValues).To(Equal(map[string][]byte{
		string([]byte{1, 2}): {1, 2, 3, 5},
		string([]byte{3, 4}): {1, 2, 3, 6},
	}))

	m.Delete([]byte{1, 2})
	Expect(m.Get([]byte{1, 2})).To(BeNil(), "Deletion should remove the value")
}