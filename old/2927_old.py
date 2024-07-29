        match = self.api_re.search(js)
        api_url = self.api_url.format(match[1])
        stream_url = self.session.http.get(api_url, schema=self.api_schema)
        return HLSStream.parse_variant_playlist(self.session, stream_url)


__plugin__ = NBCNews
