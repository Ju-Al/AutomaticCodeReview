
    def test_invalid_proj_dim(self):
        with self.assertRaisesRegexp(ValueError, r"is not supported yet"):
            x0 = np.random.random((12, 5))
            x = np.random.random((12, 5))
            layer = PolynomialCrossing(projection_dim=6)
            layer([x0, x])


if __name__ == "__main__":
    tf.test.main()
