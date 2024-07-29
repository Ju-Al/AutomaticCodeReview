        }

        [Theory]
        [InlineData("application/json1")]
        [InlineData("applicaiton/xml")]
        public async Task WhenVersionsEndpointIsCalled_GivenInvalidAcceptHeaderIsProvided_ThenServerShouldReturnUnsupportedMediaType(string acceptHeaderValue)
        {
            HttpRequestMessage request = GenerateOperationVersionsRequest(acceptHeaderValue);
            HttpResponseMessage response = await _client.SendAsync(request);

            Assert.Equal(HttpStatusCode.UnsupportedMediaType, response.StatusCode);
        }

        private HttpRequestMessage GenerateOperationVersionsRequest(
            string acceptHeader,
            string path = "$versions")
        {
            var request = new HttpRequestMessage
            {
                Method = HttpMethod.Get,
            };

            request.Headers.Add(HeaderNames.Accept, acceptHeader);
            request.RequestUri = new Uri(_client.BaseAddress, path);

            return request;
        }
    }
}
