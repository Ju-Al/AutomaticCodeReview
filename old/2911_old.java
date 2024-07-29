/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *
 *  http://aws.amazon.com/apache2.0
 *
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package software.amazon.awssdk.metrics;

import software.amazon.awssdk.annotations.SdkPublicApi;
import software.amazon.awssdk.utils.Logger;

/**
 * An implementation of {@link MetricPublisher} that writes all published metrics to the logs at the INFO level under the
 * {@code software.amazon.awssdk.metrics.LoggingMetricPublisher} namespace.
 */
@SdkPublicApi
public final class LoggingMetricPublisher implements MetricPublisher {
    private static final Logger LOGGER = Logger.loggerFor(LoggingMetricPublisher.class);

    private LoggingMetricPublisher() {
    }

    public static LoggingMetricPublisher create() {
        return new LoggingMetricPublisher();
    }

    @Override
    public void publish(MetricCollection metricCollection) {
        LOGGER.info(() -> "Metrics published: " + metricCollection);
        metrics.forEach(m -> {
            kvps.add(String.format("%s=%s", m.metric().name(), m.value()));
        });

        int maxLen = getMaxLen(kvps);

        // MetricCollection name
        LOGGER.log(logLevel, () -> String.format("[%s]%s %s",
                                                 guid,
                                                 repeat(" ", indent),
                                                 metrics.name()));

        // Open box
        LOGGER.log(logLevel, () -> String.format("[%s]%s ┌%s┐",
                                                 guid,
                                                 repeat(" ", indent),
                                                 repeat("─", maxLen + 2)));

        // Metric key-value-pairs
        kvps.forEach(kvp -> LOGGER.log(logLevel, () -> {
            return String.format("[%s]%s │ %s │",
                                 guid,
                                 repeat(" ", indent),
                                 pad(kvp, maxLen));
        }));

        // Close box
        LOGGER.log(logLevel, () -> String.format("[%s]%s └%s┘",
                                                 guid,
                                                 repeat(" ", indent),
                                                 repeat("─", maxLen + 2)));

        // Recursively repeat for any children
        metrics.children().forEach(child -> logPretty(guid, child, indent + PRETTY_INDENT_SIZE));
    }

    private static int getMaxLen(List<String> strings) {
        int maxLen = 0;
        for (String str : strings) {
            maxLen = Math.max(maxLen, str.length());
        }
        return maxLen;
    }

    private static String pad(String str, int length) {
        return str + repeat(" ", length - str.length());
    }

    @Override
    public void close() {
    }
}
