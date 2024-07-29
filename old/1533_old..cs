		{
			using (ISession session = OpenSession())
			using (session.BeginTransaction())
			{
				IEnumerable<Entity> inMemory = entities.Where(predicate).ToList();
				IEnumerable<Entity> inSession = session.Query<Entity>().Where(predicate).ToList();

				Assume.That(inMemory.Any());

				CollectionAssert.AreEquivalent(inMemory, inSession);
			}
		}
	}
}
