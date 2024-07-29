    }
    catch (const HootException& e)
    {
      exceptionMsg = e.what();
    }
    CPPUNIT_ASSERT(
      exceptionMsg.contains(
        "A *.implicitTagRules file must be the input to implicit tag rules derivation"));

    try
    {
      rulesDeriver.deriveRulesDatabase(
        inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules",
        outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runBasicTest-out.txt");
    }
    catch (const HootException& e)
    {
      exceptionMsg = e.what();
    }
    CPPUNIT_ASSERT(exceptionMsg.contains("Incorrect output specified"));
  }

  void runMinTagOccurrencePerWordTest()
  {
    QDir().mkpath(outDir());

    const QString input = inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules";
    const QString dbOutputFile =
      outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runMinTagOccurrencePerWordTest-out.sqlite";

    ImplicitTagRulesDatabaseDeriver rulesDeriver;
    rulesDeriver.setMinTagOccurrencesPerWord(4);
    rulesDeriver.setMinWordLength(1);
    rulesDeriver.setUseSchemaTagValuesForWordsOnly(true);
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setTagIgnoreFile("");
    rulesDeriver.setWordIgnoreFile("");
    rulesDeriver.deriveRulesDatabase(input, dbOutputFile);

    ImplicitTagRulesSqliteReader dbReader;
    dbReader.open(dbOutputFile);
    CPPUNIT_ASSERT_EQUAL(1L, dbReader.getRuleCount());
    dbReader.close();
  }

  void runMinWordLengthTest()
  {
    QDir().mkpath(outDir());

    const QString input = inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules";
    const QString dbOutputFile =
      outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runMinWordLengthTest-out.sqlite";

    ImplicitTagRulesDatabaseDeriver rulesDeriver;
    rulesDeriver.setMinTagOccurrencesPerWord(1);
    rulesDeriver.setMinWordLength(10);
    rulesDeriver.setUseSchemaTagValuesForWordsOnly(true);
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setTagIgnoreFile("");
    rulesDeriver.setWordIgnoreFile("");
    rulesDeriver.deriveRulesDatabase(input, dbOutputFile);

    ImplicitTagRulesSqliteReader dbReader;
    dbReader.open(dbOutputFile);
    CPPUNIT_ASSERT_EQUAL(51L, dbReader.getRuleCount());
    dbReader.close();
  }

  void runTagIgnoreTest()
  {
    QDir().mkpath(outDir());

    const QString input = inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules";
    const QString dbOutputFile =
      outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runTagIgnoreTest-out.sqlite";

    ImplicitTagRulesDatabaseDeriver rulesDeriver;
    rulesDeriver.setMinTagOccurrencesPerWord(1);
    rulesDeriver.setMinWordLength(1);
    rulesDeriver.setUseSchemaTagValuesForWordsOnly(true);
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setTagIgnoreFile(inDir() + "/ImplicitTagRulesDatabaseDeriverTest-tag-ignore-list");
    rulesDeriver.setWordIgnoreFile("");
    rulesDeriver.deriveRulesDatabase(input, dbOutputFile);

    ImplicitTagRulesSqliteReader dbReader;
    dbReader.open(dbOutputFile);
    CPPUNIT_ASSERT_EQUAL(57L, dbReader.getRuleCount());
    dbReader.close();
  }

  void runWordIgnoreTest()
  {
    QDir().mkpath(outDir());

    const QString input = inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules";
    const QString dbOutputFile =
      outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runWordIgnoreTest-out.sqlite";

    ImplicitTagRulesDatabaseDeriver rulesDeriver;
    rulesDeriver.setMinTagOccurrencesPerWord(1);
    rulesDeriver.setMinWordLength(1);
    rulesDeriver.setUseSchemaTagValuesForWordsOnly(true);
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setTagIgnoreFile("");
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setWordIgnoreFile(
      inDir() + "/ImplicitTagRulesDatabaseDeriverTest-word-ignore-list");
    rulesDeriver.deriveRulesDatabase(input, dbOutputFile);

    ImplicitTagRulesSqliteReader dbReader;
    dbReader.open(dbOutputFile);
    CPPUNIT_ASSERT_EQUAL(78L, dbReader.getRuleCount());
    dbReader.close();
  }

  void runCustomRuleTest()
  {
    QDir().mkpath(outDir());

    const QString input = inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules";
    const QString dbOutputFile =
      outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runCustomRuleTest-out.sqlite";

    ImplicitTagRulesDatabaseDeriver rulesDeriver;
    rulesDeriver.setMinTagOccurrencesPerWord(1);
    rulesDeriver.setMinWordLength(1);
    rulesDeriver.setUseSchemaTagValuesForWordsOnly(true);
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setTagIgnoreFile("");
    rulesDeriver.setCustomRuleFile(
      inDir() + "/ImplicitTagRulesDatabaseDeriverTest-custom-rules-list");
    rulesDeriver.setWordIgnoreFile("");
    rulesDeriver.deriveRulesDatabase(input, dbOutputFile);

    ImplicitTagRulesSqliteReader dbReader;
    dbReader.open(dbOutputFile);
    CPPUNIT_ASSERT_EQUAL(89L, dbReader.getRuleCount());
    dbReader.close();
  }

  void runSchemaValuesOnlyOffTest()
  {
    QDir().mkpath(outDir());

    const QString input = inDir() + "/ImplicitTagRulesDatabaseDeriverTest-input.implicitTagRules";
    const QString dbOutputFile =
      outDir() + "/ImplicitTagRulesDatabaseDeriverTest-runSchemaValuesOnlyOffTest-out.sqlite";

    ImplicitTagRulesDatabaseDeriver rulesDeriver;
    rulesDeriver.setMinTagOccurrencesPerWord(1);
    rulesDeriver.setMinWordLength(1);
    rulesDeriver.setUseSchemaTagValuesForWordsOnly(false);
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setTagIgnoreFile("");
    rulesDeriver.setCustomRuleFile("");
    rulesDeriver.setWordIgnoreFile("");
    rulesDeriver.deriveRulesDatabase(input, dbOutputFile);

    ImplicitTagRulesSqliteReader dbReader;
    dbReader.open(dbOutputFile);
    CPPUNIT_ASSERT_EQUAL(404L, dbReader.getRuleCount());
    dbReader.close();
  }
};

CPPUNIT_TEST_SUITE_NAMED_REGISTRATION(ImplicitTagRulesDatabaseDeriverTest, "quick");

}
