import unittest

from retree import treeFromTraversals


# Tests adapted from `problem-specifications//canonical-data.json` @ v1.0.0

class TwoFerTest(unittest.TestCase):
    def test_empty_tree(self):
        preorder = []
        inorder = []
        expected = {}
        self.assertEqual(treeFromTraversals(preorder, inorder), expected)

    def test_tree_with_one_item(self):
        preorder = ["a"]
        inorder = ["a"]
        expected = {"v": "a", "l": {}, "r": {}}
        self.assertEqual(treeFromTraversals(preorder, inorder), expected)

    def test_tree_with_many_items(self):
        preorder = ["a", "i", "x", "f", "r"]
        inorder = ["i", "a", "f", "x", "r"]
        expected = {"v": "a", "l": {"v": "i", "l": {}, "r": {}},
                    "r": {"v": "x", "l": {"v": "f", "l": {}, "r": {}},
                          "r": {"v": "r", "l": {}, "r": {}}}}
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