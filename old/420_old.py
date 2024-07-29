    return 0.5 * x * (1 + tf.tanh(tf.sqrt(2 / math.pi) * \
             (x + 0.044715 * tf.pow(x, 3))))
