	{

		[Test]
		public void LazyCollectionCanBeInitialized()
		{
			var collection = new LayzInitializationTestCollection(false);
			Assert.That(!NHibernateUtil.IsInitialized(collection));
			NHibernateUtil.Initialize(collection);
			Assert.That(NHibernateUtil.IsInitialized(collection));
		}

		[Test]
		public void PersistentCollectionCanBeInitialized()
		{
			var collection = new PersistentLayzInitializationTestCollection(false);
			Assert.That(!NHibernateUtil.IsInitialized(collection));
			NHibernateUtil.Initialize(collection);
			Assert.That(NHibernateUtil.IsInitialized(collection));
		}

	}
}
