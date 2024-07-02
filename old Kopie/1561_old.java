/*
        private final Map<Long, Map<String, Long>> metrics = new HashMap<>();

        InternalJobMetricsPublisher(@Nonnull JobExecutionService jobExecutionService) {
                metrics.computeIfAbsent(executionId, x -> new HashMap<>())
                        .put(name, value);
            publishLong(name, JobMetricsUtil.toLongMetricValue(value));
            jobExecutionService.updateMetrics(System.currentTimeMillis(), metrics);
            metrics.clear();
 * Copyright (c) 2008-2019, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.hazelcast.jet.impl.metrics;

import com.hazelcast.config.Config;
import com.hazelcast.core.Member;
import com.hazelcast.internal.diagnostics.Diagnostics;
import com.hazelcast.internal.metrics.ProbeLevel;
import com.hazelcast.internal.metrics.renderers.ProbeRenderer;
import com.hazelcast.jet.config.MetricsConfig;
import com.hazelcast.jet.impl.JobExecutionService;
import com.hazelcast.jet.impl.JobMetricsUtil;
import com.hazelcast.jet.impl.LiveOperationRegistry;
import com.hazelcast.jet.impl.metrics.jmx.JmxPublisher;
import com.hazelcast.jet.impl.metrics.management.ConcurrentArrayRingbuffer;
import com.hazelcast.jet.impl.metrics.management.ConcurrentArrayRingbuffer.RingbufferSlice;
import com.hazelcast.logging.ILogger;
import com.hazelcast.spi.LiveOperations;
import com.hazelcast.spi.LiveOperationsTracker;
import com.hazelcast.spi.NodeEngine;
import com.hazelcast.spi.impl.NodeEngineImpl;

import javax.annotation.Nonnull;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;

import static com.hazelcast.jet.Util.entry;
import static com.hazelcast.jet.impl.util.ExceptionUtil.withTryCatch;
import static java.util.stream.Collectors.joining;

/**
 * A service to render metrics at regular intervals and store them in a
 * ringbuffer from which the clients can read.
 */
public class JetMetricsService implements LiveOperationsTracker {

    private final NodeEngineImpl nodeEngine;
    private final ILogger logger;
    private final LiveOperationRegistry liveOperationRegistry;
    // Holds futures for pending read metrics operations
    private final ConcurrentMap<CompletableFuture<RingbufferSlice<Map.Entry<Long, byte[]>>>, Long>
            pendingReads = new ConcurrentHashMap<>();

    /**
     * Ringbuffer which stores a bounded history of metrics. For each round of collection,
     * the metrics are compressed into a blob and stored along with the timestamp,
     * with the format (timestamp, byte[])
     */
    private ConcurrentArrayRingbuffer<Map.Entry<Long, byte[]>> metricsJournal;
    private MetricsConfig config;
    private volatile ScheduledFuture<?> scheduledFuture;

    private List<MetricsPublisher> publishers;

    public JetMetricsService(NodeEngine nodeEngine) {
        this.nodeEngine = (NodeEngineImpl) nodeEngine;
        this.logger = nodeEngine.getLogger(getClass());
        this.liveOperationRegistry = new LiveOperationRegistry();
    }

    // apply MetricsConfig to HZ properties
    // these properties need to be set here so that the metrics get applied from startup
    public static void applyMetricsConfig(Config hzConfig, MetricsConfig metricsConfig) {
        if (metricsConfig.isEnabled()) {
            hzConfig.setProperty(Diagnostics.METRICS_LEVEL.getName(), ProbeLevel.INFO.name());
            if (metricsConfig.isMetricsForDataStructuresEnabled()) {
                hzConfig.setProperty(Diagnostics.METRICS_DISTRIBUTED_DATASTRUCTURES.getName(), "true");
            }
        }
    }

    public void init(NodeEngine nodeEngine, JobExecutionService jobExecutionService, MetricsConfig config) {
        this.config = config;

        this.publishers = getPublishers(nodeEngine, jobExecutionService);

        if (publishers.isEmpty()) {
            return;
        }

        logger.info("Configuring metrics collection, collection interval=" + config.getCollectionIntervalSeconds()
                + " seconds, retention=" + config.getRetentionSeconds() + " seconds, publishers="
                + publishers.stream().map(MetricsPublisher::name).collect(joining(", ", "[", "]")));

        ProbeRenderer renderer = new PublisherProbeRenderer();
        scheduledFuture = nodeEngine.getExecutionService().scheduleWithRepetition("MetricsPublisher", () -> {
            this.nodeEngine.getMetricsRegistry().render(renderer);
            for (MetricsPublisher publisher : publishers) {
                try {
                    publisher.whenComplete();
                } catch (Exception e) {
                    logger.severe("Error completing publication for publisher " + publisher, e);
                }
            }
        }, 1, config.getCollectionIntervalSeconds(), TimeUnit.SECONDS);
    }

    public boolean isEnabled() {
        return config.isEnabled();
    }

    public LiveOperationRegistry getLiveOperationRegistry() {
        return liveOperationRegistry;
    }

    @Override
    public void populate(LiveOperations liveOperations) {
        liveOperationRegistry.populate(liveOperations);
    }

    /**
     * Read metrics from the journal from the given sequence
     */
    public CompletableFuture<RingbufferSlice<Map.Entry<Long, byte[]>>> readMetrics(long startSequence) {
        if (!config.isEnabled()) {
            throw new IllegalArgumentException("Metrics collection is not enabled");
        }
        CompletableFuture<RingbufferSlice<Map.Entry<Long, byte[]>>> future = new CompletableFuture<>();
        future.whenComplete(withTryCatch(logger, (s, e) -> pendingReads.remove(future)));
        pendingReads.put(future, startSequence);

        tryCompleteRead(future, startSequence);

        return future;
    }

    private void tryCompleteRead(CompletableFuture<RingbufferSlice<Map.Entry<Long, byte[]>>> future, long sequence) {
        try {
            RingbufferSlice<Map.Entry<Long, byte[]>> slice = metricsJournal.copyFrom(sequence);
            if (!slice.isEmpty()) {
                future.complete(slice);
            }
        } catch (Exception e) {
            logger.severe("Error reading from metrics journal, sequence: " + sequence, e);
            future.completeExceptionally(e);
        }
    }

    public void reset() {
    }

    public void shutdown() {
        if (scheduledFuture != null) {
            scheduledFuture.cancel(false);
        }

        for (MetricsPublisher publisher : publishers) {
            try {
                publisher.shutdown();
            } catch (Exception e) {
                logger.warning("Error shutting down metrics publisher " + publisher.name(), e);
            }
        }
    }

    private List<MetricsPublisher> getPublishers(NodeEngine nodeEngine, JobExecutionService jobExecutionService) {
        List<MetricsPublisher> publishers = new ArrayList<>();
        if (config.isEnabled()) {
            ILogger logger = this.nodeEngine.getLogger(BlobPublisher.class);

            int journalSize = Math.max(
                    1, (int) Math.ceil((double) config.getRetentionSeconds() / config.getCollectionIntervalSeconds())
            );
            metricsJournal = new ConcurrentArrayRingbuffer<>(journalSize);
            publishers.add(
                    new BlobPublisher(
                            logger,
                            (blob, ts) -> {
                                metricsJournal.add(entry(ts, blob));
                                pendingReads.forEach(this::tryCompleteRead);
                            }
                    )
            );

            publishers.add(new InternalJobMetricsPublisher(jobExecutionService, nodeEngine.getLocalMember(), logger));
        }
        if (config.isJmxEnabled()) {
            publishers.add(new JmxPublisher(nodeEngine.getHazelcastInstance().getName(), "com.hazelcast"));
        }
        return publishers;
    }

    /**
     * Internal publisher which notifies the {@link JobExecutionService} about
     * the latest metric values.
     */
    public static class InternalJobMetricsPublisher implements MetricsPublisher {

        private final JobExecutionService jobExecutionService;
        private final String namePrefix;
        private final ILogger logger;
        private final PublisherProvider publisherProvider = new PublisherProvider();
        private final Map<Long, RawJobMetrics> jobMetrics = new HashMap<>();

        InternalJobMetricsPublisher(
                @Nonnull JobExecutionService jobExecutionService,
                @Nonnull Member member,
                @Nonnull ILogger logger
        ) {
            Objects.requireNonNull(jobExecutionService, "jobExecutionService");
            Objects.requireNonNull(member, "member");
            Objects.requireNonNull(logger, "logger");

            this.jobExecutionService = jobExecutionService;
            this.namePrefix = JobMetricsUtil.getMemberPrefix(member);
            this.logger = logger;
        }

        @Override
        public void publishLong(String name, long value) {
            Long executionId = JobMetricsUtil.getExecutionIdFromMetricDescriptor(name);
            if (executionId != null) {
                BlobPublisher blobPublisher = publisherProvider.get(executionId);
                String prefixedName = JobMetricsUtil.addPrefixToDescriptor(name, namePrefix);
                blobPublisher.publishLong(prefixedName, value);
            }
        }

        @Override
        public void publishDouble(String name, double value) {
            Long executionId = JobMetricsUtil.getExecutionIdFromMetricDescriptor(name);
            if (executionId != null) {
                BlobPublisher blobPublisher = publisherProvider.get(executionId);
                String prefixedName = JobMetricsUtil.addPrefixToDescriptor(name, namePrefix);
                blobPublisher.publishDouble(prefixedName, value);
            }
        }

        private BlobPublisher newBlobPublisher(Long executionId) {
            return new BlobPublisher(
                    logger,
                    (blob, ts) -> {
                        jobMetrics.put(executionId, RawJobMetrics.of(ts, blob));
                    }
            );
        }

        @Override
        public void whenComplete() {
            publisherProvider.complete();

            jobExecutionService.updateMetrics(jobMetrics);
            jobMetrics.clear();
        }

        @Override
        public String name() {
            return "Internal Publisher";
        }

        private class PublisherProvider {

            private final Map<Long, BlobPublisher> currentPublishers = new HashMap<>();
            private final Map<Long, BlobPublisher> previousPublishers = new HashMap<>();

            BlobPublisher get(Long executionId) {
                return currentPublishers.computeIfAbsent(
                        executionId,
                        x -> {
                            //check if we have a cached publisher
                            BlobPublisher blobPublisher = previousPublishers.remove(executionId);

                            //create a new one if we don't
                            blobPublisher = blobPublisher == null ? newBlobPublisher(executionId) : blobPublisher;

                            return blobPublisher;
                        }
                );
            }

            void complete() {
                currentPublishers.values().forEach(BlobPublisher::whenComplete);

                previousPublishers.clear();
                previousPublishers.putAll(currentPublishers);

                currentPublishers.clear();
            }

        }
    }

    /**
     * A probe renderer which renders the metrics to all the given publishers.
     */
    private class PublisherProbeRenderer implements ProbeRenderer {
        @Override
        public void renderLong(String name, long value) {
            for (MetricsPublisher publisher : publishers) {
                try {
                    publisher.publishLong(name, value);
                } catch (Exception e) {
                    logError(name, value, publisher, e);
                }
            }
        }

        @Override
        public void renderDouble(String name, double value) {
            for (MetricsPublisher publisher : publishers) {
                try {
                    publisher.publishDouble(name, value);
                } catch (Exception e) {
                    logError(name, value, publisher, e);
                }
            }
        }

        @Override
        public void renderException(String name, Exception e) {
            logger.warning("Error when rendering '" + name + '\'', e);
        }

        @Override
        public void renderNoValue(String name) {
            // noop
        }

        private void logError(String name, Object value, MetricsPublisher publisher, Exception e) {
            logger.fine("Error publishing metric to: " + publisher.name() + ", metric=" + name + ", value=" + value, e);
        }
    }
}