		this.proposals = proposals;
	}

	/**
	 * Represents a completion proposal for the DSL when using <i>e.g.</i> TAB completion in the Shell.
	 *
	 * @author Eric Bottard
	 */
	@JsonInclude(JsonInclude.Include.NON_NULL)
	public static class Proposal {

		private String text;

		private String explanation;

		private Proposal() {
			// No-arg constructor for Json serialization purposes
		}

		public Proposal(String text, String explanation) {
			this.text = text;
			this.explanation = explanation;
		}

		public String getText() {
			return text;
		}

		public String getExplanation() {
			return explanation;
		}
	}
}
