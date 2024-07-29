        {
            var param = parameters?.Parameter.Find(p =>
                string.Equals(p.Name, paramName, StringComparison.OrdinalIgnoreCase));

            var paramValue = param?.Value?.ToString();
            if (string.IsNullOrEmpty(paramValue))
            {
                throw new RequestNotValidException(string.Format(Resources.DataConvertParameterValueNotValid, paramName));
            }

            return paramValue;
        }

        private static T ReadEnumParameter<T>(Parameters parameters, string paramName)
        {
            var param = parameters?.Parameter.Find(p =>
                string.Equals(p.Name, paramName, StringComparison.OrdinalIgnoreCase));

            object enumValue;
            if (!Enum.TryParse(typeof(T), param?.Value?.ToString(), ignoreCase: true, out enumValue))
            {
                throw new RequestNotValidException(string.Format(Resources.DataConvertParameterValueNotValid, paramName));
            }

            return (T)enumValue;
        }

        private static Dictionary<string, HashSet<string>> InitSupportedParams()
        {
            var postParams = new HashSet<string>()
            {
                JobRecordProperties.InputData,
                JobRecordProperties.InputDataType,
                JobRecordProperties.TemplateSetReference,
                JobRecordProperties.EntryPointTemplate,
            };

            var patchParams = new HashSet<string>()
            {
                JobRecordProperties.MaximumConcurrency,
                JobRecordProperties.Status,
            };

            var supportedParams = new Dictionary<string, HashSet<string>>();
            supportedParams.Add(HttpMethods.Post, postParams);

            return supportedParams;
        }

        private void CheckIfDataConvertIsEnabled()
        {
            if (!_config.Enabled)
            {
                throw new RequestNotValidException(string.Format(Resources.OperationNotEnabled, OperationsConstants.DataConvert));
            }
        }
    }
}
