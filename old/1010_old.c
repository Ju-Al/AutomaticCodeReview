			json_object_set_new (response, getFieldName (FIELD_SCOPE_KEY), escapeName (tag, FIELD_SCOPE_KIND_LONG));
	}

	if (isFieldEnabled (FIELD_TYPE_REF) &&
	    tag->extensionFields.typeRef [0] != NULL  &&
	    tag->extensionFields.typeRef [1] != NULL)
		json_object_set_new (response, getFieldName (FIELD_TYPE_REF), escapeName (tag, FIELD_TYPE_REF));

	if (isFieldEnabled (FIELD_FILE_SCOPE) &&  tag->isFileScope)
		json_object_set_new (response, getFieldName (FIELD_FILE_SCOPE), json_boolean(1));

	if (isFieldEnabled (FIELD_INHERITANCE) &&
			tag->extensionFields.inheritance != NULL)
		json_object_set_new (response, getFieldName (FIELD_INHERITANCE), escapeName (tag, FIELD_INHERITANCE));

	if (isFieldEnabled (FIELD_ACCESS) &&  tag->extensionFields.access != NULL)
		json_object_set_new (response, getFieldName (FIELD_ACCESS), json_string (tag->extensionFields.access));

	if (isFieldEnabled (FIELD_IMPLEMENTATION) &&
			tag->extensionFields.implementation != NULL)
		json_object_set_new (response, getFieldName (FIELD_IMPLEMENTATION), json_string (tag->extensionFields.implementation));

	if (isFieldEnabled (FIELD_SIGNATURE) &&
			tag->extensionFields.signature != NULL)
		json_object_set_new (response, getFieldName (FIELD_SIGNATURE), escapeName (tag, FIELD_SIGNATURE));

	if (isFieldEnabled (FIELD_ROLE) && tag->extensionFields.roleIndex != ROLE_INDEX_DEFINITION)
		json_object_set_new (response, getFieldName (FIELD_ROLE), escapeName (tag, FIELD_ROLE));

	if (isFieldEnabled (FIELD_EXTRA))
	{
		json_t *value = escapeName (tag, FIELD_EXTRA);
		if (value)
			json_object_set_new (response, getFieldName (FIELD_EXTRA), value);
	}

#ifdef HAVE_LIBXML
	if (isFieldEnabled(FIELD_XPATH))
	{
		json_t *value = escapeName (tag, FIELD_XPATH);
		if (value)
			json_object_set_new (response, getFieldName (FIELD_XPATH), value);
	}
#endif
}

extern int writeJsonEntry (MIO * mio, const tagEntryInfo *const tag, void *data __unused__)
{
	json_t *response = json_pack ("{ss ss ss ss}",
		"_type", "tag",
		"name", tag->name,
		"path", tag->sourceFileName,
		"pattern", tag->pattern
	);

	if (includeExtensionFlags ())
		addExtensionFields (response, tag);

	addParserFields (response, tag);

	char *buf = json_dumps (response, 0);
	int length = mio_printf (mio, "%s\n", buf);

	free (buf);
	json_decref (response);

	return length;
}
