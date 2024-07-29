        TestCase.assertThrows(() => realm.create('ObjectA', { name: 'Boom' }));
    },

    testAliasWhenReadingProperties() {
        const realm = getRealm();
        addTestObjects(realm);

        // Internal names are not visible as properties, only aliases are.
        let obj = realm.objects("ObjectA")[0];
        TestCase.assertEqual(obj.name, undefined);
        TestCase.assertEqual(obj.otherName, 'Foo');
    },

    testAliasInQueries() {
        const realm = getRealm();
        addTestObjects(realm);

        // Queries also use aliases
        let results = realm.objects("ObjectA").filtered("otherName = 'Foo'");
        TestCase.assertEqual(results.length, 1);

        // Querying on internal names are still allowed
        results = realm.objects("ObjectA").filtered("name = 'Foo'");
        TestCase.assertEqual(results.length, 1);
    },

};