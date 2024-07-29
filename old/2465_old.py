

class Rule(object):
    """Rule properties from the rule definition file. Also finds violations."""

    def __init__(self, rule_index, rule_name, permissions, res):
        """Initialize.

        Args:
            rule_index (int): The index of the rule.
            rule_name (str): Name of the rule.
            permissions (int): Expected permissions of the role.
            res (dict): Parent resource of the role that should obey the rule.
        """
        self.rule_name = rule_name
        self.rule_index = rule_index
        self.permissions = permissions[:]
        self.res_types = res[:]

        for res_item in self.res_types:
            if 'type' not in res_item:
                raise audit_errors.InvalidRulesSchemaError(
                    'Lack of resource:type in rule {}'.format(rule_index))
            if 'resource_ids' not in res_item:
                raise audit_errors.InvalidRulesSchemaError(
                    'Lack of resource:resource_ids in rule {}'.format(
                        rule_index))

            if '*' in res_item['resource_ids']:
                res_item = ['*']

    def generate_violation(self, role):
        """Generate a violation.

        Args:
            role (Role): The role that triggers the violation.
        Returns:
            RuleViolation: The violation.
        """
        permissions = role.get_permissions()
        permissions_str = json.dumps(permissions)

        return RuleViolation(
            resource_name=role.name,
            resource_id=role.id,
            resource_type=role.type,
            full_name=role.full_name,
            rule_name=self.rule_name,
            rule_index=self.rule_index,
            violation_type=VIOLATION_TYPE,
            violation_data=permissions_str,
            resource_data=role.data,
        )

    def find_violations(self, res):
        """Get a generator for violations.

        Args:
            res (Resource): A class derived from Resource.
        Returns:
            Generator: All violations of the resource breaking the rule.

        Raises:
            ValueError: Raised if the resource type is bucket.
        """

        if res.type == 'role':
            return self.find_violations_in_role(res)
        raise ValueError(
            'only role is supported.'
        )

    def find_violations_in_role(self, role):
        """Get a generator for violations.

        Args:
            role (role): Find violation from the role.
        Returns:
            RuleViolation: All violations of the role breaking the rule.
        """
        resource_ancestors = (relationship.find_ancestors(
            role, role.full_name))

        violations = itertools.chain()
        for related_resources in resource_ancestors:
            violations = itertools.chain(
                violations,
                self.find_violations_by_ancestor(related_resources, role))
        return violations

    def find_violations_by_ancestor(self, ancestor, role):
        """Get a generator on a given ancestor of the role.

        Args:
            role (role): Role to find violation from.
            ancestor (Resource): Ancestor of the role or the role itself.
        Yields:
            RuleViolation: All violations of the role breaking the rule.
        """
        for res in self.res_types:
            if ancestor.type != res['type']:
                continue
            if '*' in res['resource_ids']:
                if set(role.get_permissions()) != set(self.permissions):
                    yield self.generate_violation(role)
            else:
                for res_id in res['resource_ids']:
                    if res_id == ancestor.id:
                        if set(role.get_permissions()) != set(self.permissions):
                            yield self.generate_violation(role)
