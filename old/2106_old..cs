			{
				Assert.Ignore(Dialect.GetType().Name + " does not support scalar sub-queries");
			}

			using (var session = OpenSession())
			using (var transaction = session.BeginTransaction())
			{
				// Simple syntax check
				session.CreateQuery(query).List();
				transaction.Commit();
			}
		}
	}
}
