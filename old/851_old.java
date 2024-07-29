    }

    @AfterTemplate
    void after(Collection<T> things) {
        assertThat(things).isNotEmpty();
    }
}
