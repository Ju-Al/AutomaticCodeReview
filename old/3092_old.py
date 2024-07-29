            headers['x-amz-security-token'] = self.credentials.token
        string_to_sign = self.canonical_string(method,
                                               split,
                                               headers,
                                               auth_path=auth_path)
        return string_to_sign
