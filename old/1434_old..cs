        public async Task<string> GetTokenAsync(string registryServer, CancellationToken cancellationToken)
#pragma warning restore CS1998 // Async method lacks 'await' operators and will run synchronously
        {
            var containerRegistryInfo = _dataConvertConfiguration.ContainerRegistries
                .FirstOrDefault(registry => string.Equals(registry.ContainerRegistryServer, registryServer, StringComparison.OrdinalIgnoreCase));
            if (containerRegistryInfo == null)
            {
                throw new ContainerRegistryNotRegisteredException(string.Format(Resources.ContainerRegistryNotRegistered, registryServer));
            }

            if (string.IsNullOrEmpty(containerRegistryInfo.ContainerRegistryServer)
                || string.IsNullOrEmpty(containerRegistryInfo.ContainerRegistryUsername)
                || string.IsNullOrEmpty(containerRegistryInfo.ContainerRegistryPassword))
            {
                throw new ContainerRegistryNotAuthorizedException(string.Format(Resources.ContainerRegistryNotAuthorized, containerRegistryInfo.ContainerRegistryServer));
            }

            return string.Format("Basic {0}", GenerateBasicToken(containerRegistryInfo.ContainerRegistryUsername, containerRegistryInfo.ContainerRegistryPassword));
        }

        private static string GenerateBasicToken(string username, string password)
        {
            var input = $"{username}:{password}";
            var inputBytes = Encoding.UTF8.GetBytes(input);
            return Convert.ToBase64String(inputBytes);
        }
    }
}
