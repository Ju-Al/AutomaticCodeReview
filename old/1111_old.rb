        end

        path ||= opts['path']

        URI.parse(File.join(url, "v1", path))
      end

      # Auth token to vault server
      def token(opts)
        if (auth = opts['auth'] || config['auth'])
          request_token(auth, opts)
        else
          ENV['VAULT_TOKEN']
        end
      end

      def request(verb, uri, data, header = {})
        # Add on any header options
        header = DEFAULT_HEADER.merge(header)

        # Create the HTTP request
        http = Net::HTTP.new(uri.host, uri.port)
        request = Net::HTTP.const_get(verb).new(uri.request_uri, header)

        # Attach any data
        request.body = data if data

        # Send the request
        begin
          response = http.request(request)
        rescue StandardError => e
          raise Bolt::Error.new(
            "Failed to connect to #{uri}: #{e.message}",
            'CONNECT_ERROR'
          )
        end

        case response
        when Net::HTTPOK
          JSON.parse(response.body)
        else
          raise VaultHTTPError, response
        end
      end

      def parse_response(response, opts)
        data = case opts['version']
               when 2
                 response['data']['data']
               else
                 response['data']
               end

        if opts['field']
          unless data[opts['field']]
            raise Bolt::ValidationError, "Unknown secrets field: #{opts['field']}"
          end
          data[opts['field']]
        else
          data
        end
      end

      # Request a token from Vault using one of the auth methods
      def request_token(auth, opts)
        case auth['method']
        when 'token'
          auth_token(auth)
        when 'userpass'
          auth_userpass(auth, opts)
        else
          raise Bolt::ValidationError, "Unknown auth method: #{auth['method']}"
        end
      end

      def validate_auth(auth, required_keys)
        required_keys.each do |key|
          next if auth[key]
          raise Bolt::ValidationError, "Expected key in #{auth['method']} auth method: #{key}"
        end
      end

      # Authenticate with Vault using the 'Token' auth method
      def auth_token(auth)
        validate_auth(auth, %w[token])
        auth['token']
      end

      # Authenticate with Vault using the 'Userpass' auth method
      def auth_userpass(auth, opts)
        validate_auth(auth, %w[user pass])
        path = "auth/userpass/login/#{auth['user']}"
        uri = uri(opts, path)
        data = { "password" => auth['pass'] }.to_json

        request(:Post, uri, data)['auth']['client_token']
      end
    end
  end
end
