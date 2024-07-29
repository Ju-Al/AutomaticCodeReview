    {
        final Duration localNode = availabilityETAOnLocalNode( instant );
        if ( !localNode.isZero() )
        {
            return localNode;
        }

        // Check all other cluster nodes.
        final Collection<Duration> objects = CacheFactory.doSynchronousClusterTask( new GetArchiveWriteETATask( instant, id ), false );
        final Duration maxDuration = objects.stream().max( Comparator.naturalOrder() ).orElse( Duration.ZERO );
        return maxDuration;
    }

    /**
     * Returns an estimation on how long it takes for all data that arrived before a certain instant will have become
     * available in the data store. When data is immediately available, 'zero', is returned;
     *
     * This method is intended to be used to determine if it's safe to construct an answer (based on database
     * content) to a request for archived data. Such response should only be generated after all data that was
     * queued before the request arrived has been written to the database.
     *
     * This method performs a check on the local cluster-node only, unlike {@link #availabilityETA(Instant)}.
     *
     * @param instant The timestamp of the data that should be available (cannot be null).
     * @return A period of time, zero when the requested data is already available.
     */
    public Duration availabilityETAOnLocalNode( final Instant instant )
    {
        if ( instant == null )
        {
            throw new IllegalArgumentException( "Argument 'instant' cannot be null." );
        }

        final Instant now = Instant.now();

        // If the date of interest is in the future, data might still arrive.
        if ( instant.isAfter( now ) )
        {
            final Duration result = Duration.between( now, instant ).plus( gracePeriod );
            Log.debug( "The timestamp that's requested ({}) is in the future. It's unpredictable if more data will become available. Data writes cannot have finished until the requested timestamp plus the grace period, which is in {}", instant, result );
            return result;
        }

        // If the last message that's processed is newer than the instant that we're after, all of the
        // data that is of interest must have been written.
        if ( lastProcessed != null && lastProcessed.isAfter( instant ) )
        {
            Log.debug( "Creation date of last logged data ({}) is younger than the timestamp that's requested ({}). Therefor, all data must have already been written.", lastProcessed, instant );
            return Duration.ZERO;
        }

        // If the date of interest is not in the future, the queue is empty, and
        // no data is currently being worked on, then all data has been written.
        if ( queue.isEmpty() && workQueue.isEmpty() )
        {
            Log.debug( "The timestamp that's requested ({}) is not in the future. All data must have already been received. There's no data queued or being worked on. Therefor, all data must have already been written.", instant );
            return Duration.ZERO;
        }

        if ( queue.isEmpty() ) {
            Log.trace( "Cannot determine with certainty if all data that arrived before {} has been written. The queue of pending writes is empty. Unless new data becomes available, the next write should occur within {}", instant, gracePeriod );
            return gracePeriod;
        } else {
            Log.trace( "Cannot determine with certainty if all data that arrived before {} has been written. The queue of pending writes contains data, which can be an indication of high load. A write should have occurred within {}", instant, maxPurgeInterval );
            return maxPurgeInterval;
        }
    }

    public int getMaxWorkQueueSize()
    {
        return maxWorkQueueSize;
    }

    public void setMaxWorkQueueSize( final int maxWorkQueueSize )
    {
        this.maxWorkQueueSize = maxWorkQueueSize;
    }

    public Duration getMaxPurgeInterval()
    {
        return maxPurgeInterval;
    }

    public void setMaxPurgeInterval( final Duration maxPurgeInterval )
    {
        this.maxPurgeInterval = maxPurgeInterval;
    }

    public Duration getGracePeriod()
    {
        return gracePeriod;
    }

    public void setGracePeriod( final Duration gracePeriod )
    {
        this.gracePeriod = gracePeriod;
    }

    protected abstract void store( List<E> batch );
}
