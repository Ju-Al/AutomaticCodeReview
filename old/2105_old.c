	def->kindTable  = HaskellKinds;
	def->kindCount  = ARRAY_SIZE(HaskellKinds);
	def->extensions = extensions;
	def->parser     = findLiterateHaskellTags;
	return def;
}
