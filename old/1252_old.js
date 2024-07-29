
Inputs:
- REGION (into command line below)
- BUCKET_NAME (in code)
- ZIP_FILE_NAME (in code)

Running the code:
    node lambda-function-setup.js REGION ACCESS_KEY_ID
*/

// snippet-start:[lambda.JavaScript.v3.LambdaFunctionSetUp]
// Load the Lambda client
const { LambdaClient, CreateFunctionCommand } = require('@aws-sdk/client-lambda');
// Instantiate a Lambda client
const region = process.argv[2];
const lambda = new LambdaClient(region);

var params = {
  Code: { /* required */
    S3Bucket: 'BUCKET_NAME',
    S3Key: 'ZIP_FILE_NAME'
  },
  FunctionName: 'slotpull', /* required */
  Handler: 'index.handler', /* required */
  Role: 'arn:aws:iam::650138640062:role/v3-lambda-tutorial-lambda-role', /* required */
  Runtime: 'nodejs12.x', /* required */
  Description: 'Slot machine game results generator',
};
lambda.send(new CreateFunctionCommand(params)).then(
  data => { console.log(data) }, // successful response
  err => {console.log(err)} // an error occurred
);

// snippet-end:[lambda.JavaScript.v3.LambdaFunctionSetUp]
