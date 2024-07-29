
                // Explicitly allow all resource types. SQL tends to create far better query plans than when there is no filter on ResourceTypeId.

                var updatedResourceTableExpressions = new List<SearchParameterExpressionBase>(expression.ResourceTableExpressions.Count + 1);
                updatedResourceTableExpressions.AddRange(expression.ResourceTableExpressions);
                updatedResourceTableExpressions.Add(AllTypesExpression);

                return new SqlRootExpression(expression.SearchParamTableExpressions, updatedResourceTableExpressions);
            }

            // There is a primary key continuation token.
            // Now look at the _type restrictions to construct a PrimaryKeyRange
            // that has only the allowed types.

            var primaryKeyParameter = (SearchParameterExpression)expression.ResourceTableExpressions[primaryKeyValueIndex];

            (short? singleAllowedResourceTypeId, BitArray allowedTypes) = TypeConstraintVisitor.Instance.Visit(expression, _model);

            var existingPrimaryKeyBinaryExpression = (BinaryExpression)primaryKeyParameter.Expression;
            var existingPrimaryKeyValue = (PrimaryKeyValue)existingPrimaryKeyBinaryExpression.Value;

            SearchParameterExpression newSearchParameterExpression;
            if (singleAllowedResourceTypeId.HasValue)
            {
                // we'll keep the existing _type parameter and just need to add a ResourceSurrogateId expression
                newSearchParameterExpression = Expression.SearchParameter(
                    SqlSearchParameters.ResourceSurrogateIdParameter,
                    new BinaryExpression(existingPrimaryKeyBinaryExpression.BinaryOperator, SqlFieldName.ResourceSurrogateId, null, existingPrimaryKeyValue.ResourceSurrogateId));
            }
            else
            {
                // Intersect allowed types with the direction of primary key parameter
                // e.g. if >, then eliminate all types that are <=
                switch (existingPrimaryKeyBinaryExpression.BinaryOperator)
                {
                    case BinaryOperator.GreaterThan:
                        for (int i = existingPrimaryKeyValue.ResourceTypeId; i >= 0; i--)
                        {
                            allowedTypes[i] = false;
                        }

                        break;
                    case BinaryOperator.LessThan:
                        for (int i = existingPrimaryKeyValue.ResourceTypeId; i < allowedTypes.Length; i++)
                        {
                            allowedTypes[i] = false;
                        }

                        break;
                    default:
                        throw new InvalidOperationException($"Unexpected operator {existingPrimaryKeyBinaryExpression.BinaryOperator}");
                }

                newSearchParameterExpression = Expression.SearchParameter(
                    primaryKeyParameter.Parameter,
                    new BinaryExpression(
                        existingPrimaryKeyBinaryExpression.BinaryOperator,
                        existingPrimaryKeyBinaryExpression.FieldName,
                        null,
                        new PrimaryKeyRange(existingPrimaryKeyValue, allowedTypes)));
            }

            var newResourceTableExpressions = new List<SearchParameterExpressionBase>();
            for (var i = 0; i < expression.ResourceTableExpressions.Count; i++)
            {
                if (i == primaryKeyValueIndex || // eliminate the existing primaryKey expression
                    (singleAllowedResourceTypeId == null && // if there are many possible types, the PrimaryKeyRange expression will already be constrained to those types
                     expression.ResourceTableExpressions[i] is SearchParameterExpression searchParameterExpression &&
                     searchParameterExpression.Parameter.Name == SearchParameterNames.ResourceType))
                {
                    continue;
                }

                newResourceTableExpressions.Add(expression.ResourceTableExpressions[i]);
            }

            newResourceTableExpressions.Add(newSearchParameterExpression);
            return new SqlRootExpression(expression.SearchParamTableExpressions, newResourceTableExpressions);
        }
    }
}
