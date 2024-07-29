            throw new IllegalArgumentException("tags' length is odd, gauge need even.");
        }
        return INSTANCE.gaugesMap.computeIfAbsent(getKey(name, tags), s -> {
            List<Tag> tagList = new ArrayList<>();
            for (int i = 0; i < tags.length;) {
                tagList.add(new ImmutableTag(tags[i++], tags[i++]));
            }
            return Metrics.gauge(name, tagList, new AtomicLong());
        });
    }
    
    /**
     * exporter for Metrics.
     */
    private static List<Metrics> exporter() {
        return null;
    }
    
    /**
     * create key for metrics.
     */
    private static String getKey(String name, String... k) {
        StringBuilder sb = new StringBuilder(name + SPILT);
        for (String s : k) {
            sb.append(s).append(SPILT);
        }
        return sb.toString();
    }
    
}
