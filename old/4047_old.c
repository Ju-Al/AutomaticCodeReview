
		// Feed key name into sha_256_write() without NULL terminator (hence -1).
		// This makes it easier to compare expected results with other sha256 tools.
		sha_256_write (&sha_256, keyName (currentKey), keyGetNameSize (currentKey) - 1);
		// Note: The value of the key itself is not relevant / part of specification. Only the key's name + its metadata!

		KeySet * currentMetaKeys = keyMeta (currentKey);
		// Feed name + values from meta keys into sha_256_write().
		for (elektraCursor metaIt = 0; metaIt < ksGetSize (currentMetaKeys); metaIt++)
		{
			Key * currentMetaKey = ksAtCursor (currentMetaKeys, metaIt);
			sha_256_write (&sha_256, keyName (currentMetaKey), keyGetNameSize (currentMetaKey) - 1);
			sha_256_write (&sha_256, keyString (currentMetaKey), keyGetValueSize (currentMetaKey) - 1);
		}
	}

	sha_256_close (&sha_256);
	hash_to_string (hash_string, hash);

	ksDel (dupKs);
	ksDel (cutKs);

	return true;
}

/**
 * Convert hash array to hex string
 * Copied from https://github.com/amosnier/sha-2/blob/f0d7baf076207b943649e68946049059f018c10b/test.c
 * @param string A string array of length 65 (64 characters + \0)
 * @param hash The input hash array
 */
static void hash_to_string (char string[65], const uint8_t hash[32])
{
	size_t i;
	for (i = 0; i < 32; i++)
	{
		string += sprintf (string, "%02x", hash[i]);
	}
}
