            {
                throw new InvalidOperationException("Cannot create an instance of DefaultMetadataModuleResolver with conventions parameter having null value.");
            }

            this.conventions = conventions;
            this.catalog = catalog;
        }

        /// <summary>
        /// Resolves a metadata module instance based on the provided information.
        /// </summary>
        /// <param name="moduleType">The type of the <see cref="INancyModule"/>.</param>
        /// <returns>An <see cref="IMetadataModule"/> instance if one could be found, otherwise <see langword="null"/>.</returns>
        public IMetadataModule GetMetadataModule(Type moduleType)
        {
            var metadataModuleTypes = this.catalog.GetMetadataModuleTypes();

            foreach (var convention in this.conventions)
            {
                var metadataModuleType = SafeInvokeConvention(convention, moduleType, metadataModuleTypes);

                if (metadataModuleType != null)
                {
                    return this.catalog.GetMetadataModule(metadataModuleType);
                }
            }

            return null;
        }

        private static Type SafeInvokeConvention(Func<Type, IEnumerable<Type>, Type> convention, Type moduleType, IEnumerable<Type> metadataModuleTypes)
        {
            try
            {
                return convention.Invoke(moduleType, metadataModuleTypes);
            }
            catch
            {
                return null;
            }
        }
    }
}