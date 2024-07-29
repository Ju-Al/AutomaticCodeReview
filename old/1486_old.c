	 * ascending order
	 */
	{
		opmphmRatio = 1;
		for (size_t w = 2; w < opmphmGetWidth (maxNsimple); ++w)
		{
			size_t n = getPower (w, OPMPHMNOEIOP);
			void * datap = elektraMalloc (sizeof (void *) * n);
			OpmphmOrder * order = elektraMalloc (sizeof (OpmphmOrder) * n);
			exit_if_fail (datap, "malloc");
			exit_if_fail (order, "malloc");
			// prepare
			OpmphmInit init;
			init.getString = test_opmphm_getString;
			init.seed = 0;
			init.data = datap;
			Opmphm opmphm;
			// fill
			for (size_t i = 0; i < n; ++i)
			{
				getOrderedPair (&order[i], w, i);
			}
			// check
			OpmphmOrder ** sortOrder = opmphmInit (&opmphm, &init, order, n);
			succeed_if (sortOrder, "not duplicate data marked as duplicate");
			for (size_t i = 0; i < n - 1; ++i)
			{
				succeed_if (getNumber (sortOrder[i], w) < getNumber (sortOrder[i + 1], w), "sort incorrect");
			}

			elektraFree (sortOrder);
			elektraFree (order);
			elektraFree (datap);
		}
	}
	/* test w, from 1 to full.
	 * ascending order
	 */
	{
		for (size_t w = 2; w < opmphmGetWidth (maxNcomplex); ++w)
		{
			size_t power = getPower (w, OPMPHMNOEIOP);
			void * datap = elektraMalloc (sizeof (void *) * power);
			OpmphmOrder * order = elektraMalloc (sizeof (OpmphmOrder) * power);
			exit_if_fail (datap, "malloc");
			exit_if_fail (order, "malloc");
			for (size_t n = 1; n < power; ++n)
			{
				opmphmRatio = (double)power / n;
				// prepare
				OpmphmInit init;
				init.getString = test_opmphm_getString;
				init.seed = 0;
				init.data = datap;
				Opmphm opmphm;
				// fill
				for (size_t i = 0; i < n; ++i)
				{
					getOrderedPair (&order[i], w, i);
				}
				// check
				OpmphmOrder ** sortOrder = opmphmInit (&opmphm, &init, order, n);
				succeed_if (sortOrder, "not duplicate data marked as duplicate");
				for (size_t i = 0; i < n - 1; ++i)
				{
					succeed_if (getNumber (sortOrder[i], w) < getNumber (sortOrder[i + 1], w), "sort incorrect");
				}
				elektraFree (sortOrder);
			}
			elektraFree (order);
			elektraFree (datap);
		}
	}
	/* test w, always full. w^OPMPHMNOEIOP elements in each run.
	 * descending order
	 */
	{
		opmphmRatio = 1;
		for (size_t w = 2; w < opmphmGetWidth (maxNsimple); ++w)
		{
			size_t n = getPower (w, OPMPHMNOEIOP);
			void * datap = elektraMalloc (sizeof (void *) * n);
			OpmphmOrder * order = elektraMalloc (sizeof (OpmphmOrder) * n);
			exit_if_fail (datap, "malloc");
			exit_if_fail (order, "malloc");
			// prepare
			OpmphmInit init;
			init.getString = test_opmphm_getString;
			init.seed = 0;
			init.data = datap;
			Opmphm opmphm;
			// fill
			for (size_t i = 0; i < n; ++i)
			{
				getOrderedPair (&order[i], w, n - 1 - i);
			}
			// check
			OpmphmOrder ** sortOrder = opmphmInit (&opmphm, &init, order, n);
			succeed_if (sortOrder, "not duplicate data marked as duplicate");
			for (size_t i = 0; i < n - 1; ++i)
			{
				succeed_if (getNumber (sortOrder[i], w) < getNumber (sortOrder[i + 1], w), "sort incorrect");
			}

			elektraFree (sortOrder);
			elektraFree (order);
			elektraFree (datap);
		}
	}
	/* test w, from 1 to full.
	 * descending order
	 */
	{
		for (size_t w = 2; w < opmphmGetWidth (maxNcomplex); ++w)
		{
			size_t power = getPower (w, OPMPHMNOEIOP);
			void * datap = elektraMalloc (sizeof (void *) * power);
			OpmphmOrder * order = elektraMalloc (sizeof (OpmphmOrder) * power);
			exit_if_fail (datap, "malloc");
			exit_if_fail (order, "malloc");
			for (size_t n = 1; n < power; ++n)
			{
				opmphmRatio = (double)power / n;
				// prepare
				OpmphmInit init;
				init.getString = test_opmphm_getString;
				init.seed = 0;
				init.data = datap;
				Opmphm opmphm;
				// fill
				for (size_t i = 0; i < n; ++i)
				{
					getOrderedPair (&order[i], w, n - 1 - i);
				}
				// check
				OpmphmOrder ** sortOrder = opmphmInit (&opmphm, &init, order, n);
				succeed_if (sortOrder, "not duplicate data marked as duplicate");
				for (size_t i = 0; i < n - 1; ++i)
				{
					succeed_if (getNumber (sortOrder[i], w) < getNumber (sortOrder[i + 1], w), "sort incorrect");
				}
				elektraFree (sortOrder);
			}
			elektraFree (order);
			elektraFree (datap);
		}
	}
}


static void test_fail ()
{
	/* test w, always full. w^OPMPHMNOEIOP elements in each run.
	 * ascending order
	 * only 1 duplicate, should be detected during partition
	 */
	{
		opmphmRatio = 1;
		for (size_t w = 2; w < opmphmGetWidth (maxNsimple); ++w)
		{
			size_t n = getPower (w, OPMPHMNOEIOP);
			void * datap = elektraMalloc (sizeof (void *) * n);
			OpmphmOrder * order = elektraMalloc (sizeof (OpmphmOrder) * n);
			exit_if_fail (datap, "malloc");
			exit_if_fail (order, "malloc");
			// prepare
			OpmphmInit init;
			init.getString = test_opmphm_getString;
			init.seed = 0;
			init.data = datap;
			Opmphm opmphm;
			// fill
			for (size_t i = 0; i < n; ++i)
			{
				getOrderedPair (&order[i], w, i);
			}
			order[n * (3 / 4)].h[OPMPHMNOEIOP / 2] = (order[n * (3 / 4)].h[OPMPHMNOEIOP / 2] + 1) % w;
			// check
			OpmphmOrder ** sortOrder = opmphmInit (&opmphm, &init, order, n);
			//~ succeed_if (!sortOrder, "duplicate data marked as ok");

			elektraFree (sortOrder);
			elektraFree (order);
			elektraFree (datap);
		}
	}
	/* test w, from 1 to full-1.
	 * ascending order
	 * use the last for a duplicate
	 */
	{
		for (size_t w = 2; w < opmphmGetWidth (maxNcomplex); ++w)
		{
			size_t power = getPower (w, OPMPHMNOEIOP);
			void * datap = elektraMalloc (sizeof (void *) * power);
			OpmphmOrder * order = elektraMalloc (sizeof (OpmphmOrder) * power);
			exit_if_fail (datap, "malloc");
			exit_if_fail (order, "malloc");
			for (size_t n = 2; n < power; ++n)
			{
				opmphmRatio = (double)power / n;
				// prepare
				OpmphmInit init;
				init.getString = test_opmphm_getString;
				init.seed = 0;
				init.data = datap;
				Opmphm opmphm;
				// fill
				for (size_t i = 0; i < n; ++i)
				{
					getOrderedPair (&order[i], w, i);
				}
				for (unsigned int t = 0; t < OPMPHMNOEIOP; ++t)
				{
					order[n - 1].h[t] = order[0].h[t];
				}
				// check
				OpmphmOrder ** sortOrder = opmphmInit (&opmphm, &init, order, n);
				succeed_if (!sortOrder, "duplicate data marked as ok");
				elektraFree (sortOrder);
			}
			elektraFree (order);
			elektraFree (datap);
		}
	}
}


int main (int argc, char ** argv)
{
	printf ("OPMPHM INIT      TESTS\n");
	printf ("==================\n\n");

	init (argc, argv);

	test_success ();
	test_fail ();

	printf ("\ntest_key RESULTS: %d test(s) done. %d error(s).\n", nbTest, nbError);

	return nbError;
}
