  EXPECT_EQ(lifeCycleCostUsePriceEscalation.resource().get(), "Electricity");
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setResource("NaturalGas"));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.resource());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.resource().get(), "NaturalGas");
  EXPECT_FALSE(lifeCycleCostUsePriceEscalation.setResource("Natural Gas"));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.resource());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.resource().get(), "NaturalGas");

  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setEscalationStartMonth("January"));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.escalationStartMonth());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.escalationStartMonth().get(), "January");
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setEscalationStartMonth("February"));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.escalationStartMonth());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.escalationStartMonth().get(), "February");
  EXPECT_FALSE(lifeCycleCostUsePriceEscalation.setEscalationStartMonth("Febuary"));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.escalationStartMonth());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.escalationStartMonth().get(), "February");

  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setEscalationStartYear(2020));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.escalationStartYear());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.escalationStartYear().get(), 2020);
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setEscalationStartYear(2021));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.escalationStartYear());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.escalationStartYear().get(), 2021);
  EXPECT_FALSE(lifeCycleCostUsePriceEscalation.setEscalationStartYear(1899));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.escalationStartYear());
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.escalationStartYear().get(), 2021);

  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setYearEscalation(0, 100.0));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.yearEscalation(0));
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.yearEscalation(0).get(), 100.0);
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.numYears(), 1);
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setYearEscalation(0, 200.0));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.yearEscalation(0));
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.yearEscalation(0).get(), 200.0);
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.numYears(), 1);
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.setYearEscalation(1, 300.0));
  EXPECT_TRUE(lifeCycleCostUsePriceEscalation.yearEscalation(1));
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.yearEscalation(1).get(), 300.0);
  EXPECT_EQ(lifeCycleCostUsePriceEscalation.numYears(), 2);
}