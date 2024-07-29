            + '}';
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        ParameterizedMetricKey that = (ParameterizedMetricKey) o;

        if (!key.equals(that.key)) {
            return false;
        }
        return version.equals(that.version);
    }

    @Override
    public int hashCode() {
        int result = key.hashCode();
        result = 31 * result + version.hashCode();
        return result;
    }
}
