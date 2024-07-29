        self.assertEqual(treeFromTraversals(preorder, inorder), expected)

    def test_reject_traverals_of_different_length(self):
        preorder = ["a", "b"]
        inorder = ["b", "a", "r"]
        with self.assertRaisesWithMessage(ValueError):
            treeFromTraversals(preorder, inorder)

    def test_reject_inconsistent_traversals_of_same_length(self):
        preorder = ["x", "y", "z"]
        inorder = ["a", "b", "c"]
        with self.assertRaisesWithMessage(ValueError):
            treeFromTraversals(preorder, inorder)

    def test_reject_traverals_with_repeated_items(self):
        preorder = ["a", "b", "a"]
        inorder = ["b", "a", "a"]
        with self.assertRaisesWithMessage(ValueError):
            treeFromTraversals(preorder, inorder)

    # Utility functions
    def setUp(self):
        try:
            self.assertRaisesRegex
        except AttributeError:
            self.assertRaisesRegex = self.assertRaisesRegexp

    def assertRaisesWithMessage(self, exception):
        return self.assertRaisesRegex(exception, r".+")


if __name__ == '__main__':
    unittest.main()
