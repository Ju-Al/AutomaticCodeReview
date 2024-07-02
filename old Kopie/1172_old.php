<?php
/**
 * Copyright 2010-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
 *
 * This file is licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License. A copy of
 * the License is located at
 *
 * http://aws.amazon.com/apache2.0/
 *
 * This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *

/**
 * Describe Alarm History in CloudWatch
 *
 * This code expects that you have AWS credentials set up per:
 * https://docs.aws.amazon.com/sdk-for-php/v3/developer-guide/guide_credentials.html
 */
Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// snippet-start:[cloudwatch.php.describe_alarm_history.complete]
// snippet-start:[cloudwatch.php.describe_alarm_history.import]
require 'vendor/autoload.php';

use Aws\CloudWatch\CloudWatchClient; 
use Aws\Exception\AwsException;
// snippet-end:[cloudwatch.php.describe_alarm_history.import]

/* ////////////////////////////////////////////////////////////////////////////
 * Purpose: Provides history information for an existing Amazon CloudWatch 
 * alarm.
 *
 * Prerequisites: An existing CloudWatch alarm.
 * 
 * Inputs:
 * - $cloudWatchClient: An initialized AWS SDK for PHP SDK client 
 *   for CloudWatch.
 * - $alarmName: The name of the alarm to provide history information about.
 * 
 * Returns: Information about any alarm history if found; 
 * otherwise, the error message.
 * ///////////////////////////////////////////////////////////////////////// */

// snippet-start:[cloudwatch.php.describe_alarm_history.main]
function describeAlarmHistory($cloudWatchClient, $alarmName)
{
    try {
        $result = $cloudWatchClient->describeAlarmHistory([
            'AlarmName' => $alarmName
        ]);

        $message = '';

        if (isset($result['@metadata']['effectiveUri']))
        {
            $message .= 'Alarm history at the effective URI ' . 
                $result['@metadata']['effectiveUri'] . 
                ' for alarm ' . $alarmName . ":\n\n";

            if ((isset($result['AlarmHistoryItems'])) and 
                (count($result['AlarmHistoryItems']) > 0))
            {
                foreach ($result['AlarmHistoryItems'] as $alarm) {
                    $message .= $alarm['HistoryItemType'] . ' at ' . 
                        $alarm['Timestamp'] . "\n";
                }
            } else {
                $message .= 'No alarm history found for ' . 
                    $alarmName . '.';
            }
        } else {
            $message .= 'No such alarm or no history found for ' . 
                $alarmName . '.';
        }

        return $message;
    } catch (AwsException $e) {
        return 'Error: ' . $e->getAwsErrorMessage();
    }
}

function describeTheAlarmHistory()
{
    $alarmName = 'my-alarm';

    $cloudWatchClient = new CloudWatchClient([
        'profile' => 'default',
        'region' => 'us-east-1',
        'version' => '2010-08-01'
    ]);

    echo describeAlarmHistory($cloudWatchClient, $alarmName);
}

// Uncomment the following line to run this code in an AWS account.
// describeTheAlarmHistory();
// snippet-end:[cloudwatch.php.describe_alarm_history.main]
// snippet-end:[cloudwatch.php.describe_alarm_history.complete]
// snippet-sourcedescription:[DescribeAlarmHistory.php demonstrates how to retrieve history for the specified alarm in Amazon CloudWatch.]
// snippet-keyword:[PHP]
// snippet-sourcesyntax:[php]
// snippet-keyword:[AWS SDK for PHP v3]
// snippet-keyword:[Code Sample]
// snippet-keyword:[Amazon Cloudwatch]
// snippet-service:[cloudwatch]
// snippet-sourcetype:[full-example]
// snippet-sourcedate:[2020-05-05]
// snippet-sourceauthor:[pccornel (AWS)]

