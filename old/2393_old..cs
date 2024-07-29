
        public static Regex Get(string pattern, RegexOptions options)
        {
            if (Compiled)
            {
                return _cache.GetOrAdd((pattern, options), k => new Regex(k.pattern, k.options | RegexOptions.Compiled));
            }
            else
            {
                return _cache.GetOrAdd((pattern, options), k => new Regex(k.pattern, k.options));
            }
        }

        public static Regex Get(string pattern) => Get(pattern, default);

        public static bool IsMatch(string text, string pattern)
        {
            return Get(pattern).IsMatch(text);
        }
    }
}