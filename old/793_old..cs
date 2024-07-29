        }

        [HttpIntegrationFixtureArgumentSets(DataStore.All, Format.Xml)]
        [Theory]
        [InlineData("Observation/$validate", "<Observation xmlns=\"http://hl7.org/fhir\"><status value=\"final\"/><code><coding><system value=\"system\"/><code value=\"code\"/></coding></code></Observation>")]
        public async void GivenAValidateRequestInXML_WhenTheResourceIsValid_ThenAnOkMessageIsReturned(string path, string payload)
        {
            OperationOutcome outcome = await _client.ValidateAsync(path, payload, true);

            Assert.Single(outcome.Issue);
            CheckOperationOutcomeIssue(
                outcome.Issue[0],
                OperationOutcome.IssueSeverity.Information,
                OperationOutcome.IssueType.Informational,
                "All OK");
        }

        [Theory]
        [InlineData(
            "Patient/$validate",
            "{\"resourceType\":\"Patient\",\"name\":\"test, one\"}",
            "Type checking the data: Since type HumanName is not a primitive, it cannot have a value (at Resource.name[0])")]
        public async void GivenAValidateRequest_WhenTheResourceIsInvalid_ThenADetailedErrorIsReturned(string path, string payload, string expectedIssue)
        {
            OperationOutcome outcome = await _client.ValidateAsync(path, payload);

            Assert.Single(outcome.Issue);
            CheckOperationOutcomeIssue(
                    outcome.Issue[0],
                    OperationOutcome.IssueSeverity.Error,
                    OperationOutcome.IssueType.Invalid,
                    expectedIssue);
        }

        [Theory]
        [InlineData(
            "Observation/$validate",
            "{\"resourceType\":\"Observation\",\"code\":{\"coding\":[{\"system\":\"system\",\"code\":\"code\"}]}}",
            "Element with min. cardinality 1 cannot be null",
            "Observation.StatusElement")]
        [InlineData(
            "Observation/$validate",
            "{\"resourceType\":\"Patient\",\"name\":{\"family\":\"test\",\"given\":\"one\"}}",
            "Resource type in the URL must match resourceType in the resource.",
            "TypeName")]
        public async void GivenAValidateRequest_WhenTheResourceIsInvalid_ThenADetailedErrorWithLocationsIsReturned(string path, string payload, string expectedIssue, string location)
        {
            OperationOutcome outcome = await _client.ValidateAsync(path, payload);

            Assert.Single(outcome.Issue);
            CheckOperationOutcomeIssue(
                    outcome.Issue[0],
                    OperationOutcome.IssueSeverity.Error,
                    OperationOutcome.IssueType.Invalid,
                    expectedIssue,
                    location);
        }

        [Fact]
        public async void GivenAValidateByIdRequest_WhenTheResourceIsValid_ThenAnOkMessageIsReturned()
        {
            var payload = "{\"resourceType\": \"Patient\", \"id\": \"123\"}";

            OperationOutcome outcome = await _client.ValidateAsync("Patient/123/$validate", payload);

            Assert.Single(outcome.Issue);
            CheckOperationOutcomeIssue(
                outcome.Issue[0],
                OperationOutcome.IssueSeverity.Information,
                OperationOutcome.IssueType.Informational,
                "All OK");
        }

        [Fact]
        public async void GivenAValidateByIdRequest_WhenTheResourceIdDoesNotMatch_ThenADetailedErrorIsReturned()
        {
            var payload = "{\"resourceType\": \"Patient\", \"id\": \"456\"}";

            OperationOutcome outcome = await _client.ValidateAsync("Patient/123/$validate", payload);

            Assert.Single(outcome.Issue);
            CheckOperationOutcomeIssue(
                outcome.Issue[0],
                OperationOutcome.IssueSeverity.Error,
                OperationOutcome.IssueType.Invalid,
                "Id in the URL must match id in the resource.",
                "Patient.id");
        }

        [Fact]
        public async void GivenAValidateByIdRequest_WhenTheResourceIdIsMissing_ThenADetailedErrorIsReturned()
        {
            var payload = "{\"resourceType\": \"Patient\"}";

            OperationOutcome outcome = await _client.ValidateAsync("Patient/123/$validate", payload);

            Assert.Single(outcome.Issue);
            CheckOperationOutcomeIssue(
                outcome.Issue[0],
                OperationOutcome.IssueSeverity.Error,
                OperationOutcome.IssueType.Invalid,
                "Id must be specified in the resource.",
                "Patient.id");
        }

        private void CheckOperationOutcomeIssue(
            OperationOutcome.IssueComponent issue,
            OperationOutcome.IssueSeverity? expectedSeverity,
            OperationOutcome.IssueType? expectedCode,
            string expectedMessage,
            string expectedLocation = null)
        {
            // Check expected outcome
            Assert.Equal(expectedSeverity, issue.Severity);
            Assert.Equal(expectedCode, issue.Code);
            Assert.Equal(expectedMessage, issue.Diagnostics);

            if (expectedLocation != null)
            {
                Assert.Single(issue.LocationElement);
                Assert.Equal(expectedLocation, issue.LocationElement[0].ToString());
            }
        }

        private string ExtractFromJson(string json, string property, bool isArray = false)
        {
            var propertyWithQuotes = property + "\":" + (isArray ? "[" : string.Empty) + "\"";
            var start = json.IndexOf(propertyWithQuotes) + propertyWithQuotes.Length;
            var end = json.IndexOf("\"", start);

            return json.Substring(start, end - start);
        }
    }
}