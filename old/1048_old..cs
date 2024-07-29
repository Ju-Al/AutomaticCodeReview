        {
            _ScopeGroups = scopeGroups;
        }

        protected override Task HandleRequirementAsync(
            AuthorizationHandlerContext context, MultipleScopeGroupsRequirement requirement)
        {
            if (_ScopeGroups.Any(scopeGroup =>
                scopeGroup.All(s => context.User.HasClaim(OpenIddictConstants.Claims.Scope, s))))
            {
                context.Succeed(requirement);
            }

            return Task.CompletedTask;
        }
    }
}
