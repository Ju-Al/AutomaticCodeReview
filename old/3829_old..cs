                _propertySetter = propertySetterCache;
            }

            return propertySetterCache.Value?.SetPropertyValue(dbConnection, propertyValue) ?? false;
        }

        KeyValuePair<KeyValuePair<string, Type>, IPropertySetter> _propertySetter;
    }
}

#endif