                .hasCauseInstanceOf(QueryException.class)
                .hasMessageContaining("CROSS join not supported");
    }
}
