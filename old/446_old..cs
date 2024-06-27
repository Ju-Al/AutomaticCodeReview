﻿using System.Collections.Concurrent;
        private static readonly ConcurrentDictionary<string, FractionExtractor> Instances = new ConcurrentDictionary<string, FractionExtractor>();
        public static FractionExtractor GetInstance(string placeholder = "")

            if (!Instances.ContainsKey(placeholder))
                var instance = new FractionExtractor();
                Instances.TryAdd(placeholder, instance);
            return Instances[placeholder];
        private FractionExtractor()
                        RegexOptions.IgnoreCase | RegexOptions.Singleline)
                    , "FracNum"
                        RegexOptions.IgnoreCase | RegexOptions.Singleline)
                    , "FracNum"
                        RegexOptions.IgnoreCase | RegexOptions.Singleline)
                    , "FracEng"
                        NumbersDefinitions.FractionNounWithArticleRegex, RegexOptions.IgnoreCase | RegexOptions.Singleline)
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Text.RegularExpressions;
using Microsoft.Recognizers.Definitions.English;

namespace Microsoft.Recognizers.Text.Number.English
{
    public class FractionExtractor : BaseNumberExtractor
    {
        internal sealed override ImmutableDictionary<Regex, string> Regexes { get; }

        protected sealed override NumberOptions Options { get; }

        protected sealed override string ExtractType { get; } = Constants.SYS_NUM_FRACTION; // "Fraction";

        private static readonly ConcurrentDictionary<(NumberOptions, string), FractionExtractor> Instances =
            new ConcurrentDictionary<(NumberOptions, string), FractionExtractor>();

        public static FractionExtractor GetInstance(NumberOptions options = NumberOptions.None, string placeholder = "")
        {
            var cacheKey = (options, placeholder);
            if (!Instances.ContainsKey(cacheKey))
            {
                var instance = new FractionExtractor(options);
                Instances.TryAdd(cacheKey, instance);
            }

            return Instances[cacheKey];
        }

        private FractionExtractor(NumberOptions options)
        {
            Options = options;

            var regexes = new Dictionary<Regex, string>
            {
                {
                    new Regex(NumbersDefinitions.FractionNotationWithSpacesRegex,
                        RegexOptions.IgnoreCase | RegexOptions.Singleline),
                    "FracNum"
                },
                {
                    new Regex(NumbersDefinitions.FractionNotationRegex,
                        RegexOptions.IgnoreCase | RegexOptions.Singleline),
                    "FracNum"
                },
                {
                    new Regex(
                        NumbersDefinitions.FractionNounRegex,
                        RegexOptions.IgnoreCase | RegexOptions.Singleline),
                    "FracEng"
                },
                {
                    new Regex(
                        NumbersDefinitions.FractionNounWithArticleRegex,
                        RegexOptions.IgnoreCase | RegexOptions.Singleline),
                    "FracEng"
                }
            };

            if ((Options & NumberOptions.PercentageMode) != 0)
            {
                regexes.Add(
                    new Regex(
                        NumbersDefinitions.FractionPrepositionWithinPercentModeRegex,
                        RegexOptions.IgnoreCase | RegexOptions.Singleline)
                    , "FracEng"
                );
            }
            else
            {
                regexes.Add(
                    new Regex(
                        NumbersDefinitions.FractionPrepositionRegex,
                        RegexOptions.IgnoreCase | RegexOptions.Singleline)
                    , "FracEng"
                );
            }

            Regexes = regexes.ToImmutableDictionary();
        }
    }
}