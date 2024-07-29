        return new SingularityExecutorLogrotateAdditionalFile(value, Optional.<String>absent(), Optional.<String>absent());
    }

    @JsonCreator
    public SingularityExecutorLogrotateAdditionalFile(@JsonProperty("filename") String filename,
        @JsonProperty("extension") Optional<String> extension,
        @JsonProperty("dateformat") Optional<String> dateformat) {
        this.filename = filename;
        this.extension = extension;
        this.dateformat = dateformat;
    }

    public String getFilename() {
        return filename;
    }

    public Optional<String> getExtension() {
        return extension;
    }

    public Optional<String> getDateformat() {
        return dateformat;
    }
}
