	;

	private final String name;
	private final Doc doc;

	public ChannelIdImpl(String name, Doc doc) {
		this.name = name;
		this.doc = doc;
	}

	@Override
	public String name() {
		return this.name;
	}

	@Override
	public Doc doc() {
		return this.doc;
	}
}
