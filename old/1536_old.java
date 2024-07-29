    private static TransferUtility util;
    private static Context context = InstrumentationRegistry.getContext();

    private File file;
    private CountDownLatch started;
    private CountDownLatch paused;
    private CountDownLatch resumed;
    private CountDownLatch completed;

    /**
     * Creates and initializes all the test resources needed for these tests.
     */
    @BeforeClass
    public static void setUpBeforeClass() throws Exception {
        setUp();
        TransferNetworkLossHandler.getInstance(context);
        util = TransferUtility.builder()
                .context(context)
                .s3Client(s3)
                .build();

        try {
            s3.createBucket(bucketName);
            waitForBucketCreation(bucketName);
        } catch (final Exception e) {
            System.out.println("Error in creating the bucket. "
                    + "Please manually create the bucket " + bucketName);
        }
    }

    @AfterClass
    public static void tearDown() {
        try {
            deleteBucketAndAllContents(bucketName);
        } catch (final Exception e) {
            System.out.println("Error in deleting the bucket. "
                    + "Please manually delete the bucket " + bucketName);
            e.printStackTrace();
        }
    }

    @Before
    public void setUpLatches() {
        started = new CountDownLatch(1);
        paused = new CountDownLatch(1);
        resumed = new CountDownLatch(1);
        completed = new CountDownLatch(1);
    }

    @Test
    public void testSinglePartUploadPause() throws Exception {
        // Small (1KB) file upload
        file = getRandomTempFile("small", 1000L);
        testUploadPause();
    }

    @Test
    public void testMultiPartUploadPause() throws Exception {
        // Large (10MB) file upload
        file = getRandomSparseFile("large", 10L * 1024 * 1024);
        testUploadPause();
    }

    private void testUploadPause() throws Exception {
        // start transfer and wait for progress
        TransferObserver observer = util.upload(bucketName, file.getName(), file);
        observer.setTransferListener(new TestListener());
        started.await(100, TimeUnit.MILLISECONDS);

        // pause and wait
        util.pause(observer.getId());
        paused.await(100, TimeUnit.MILLISECONDS);
        Thread.sleep(1000); // throws if progress is made

        // resume if pause was properly executed
        util.resume(observer.getId());
        resumed.await(100, TimeUnit.MILLISECONDS);

        // cancel early to avoid having to wait for completion
        util.cancel(observer.getId());
        completed.await(100, TimeUnit.MILLISECONDS);
    }

    private final class TestListener implements TransferListener {
        @Override
        public void onStateChanged(int id, TransferState state) {
            switch (state) {
                case CANCELED:
                case COMPLETED:
                    completed.countDown();
                    break;
                case PAUSED:
                    paused.countDown();
                    break;
                case IN_PROGRESS:
                    if (paused.getCount() == 0) {
                        // Post-pause
                        resumed.countDown();
                    } else {
                        // Pre-pause
                        started.countDown();
                    }
                    break;
                case FAILED:
                    throw new RuntimeException("Failed transfer.");
            }
        }

        @Override
        public void onProgressChanged(int id, long bytesCurrent, long bytesTotal) {
            if (paused.getCount() == 0 && resumed.getCount() > 0) {
                throw new RuntimeException("Progress made even while paused.");
            }
        }

        @Override
        public void onError(int id, Exception ex) {
            throw new RuntimeException(ex);
        }
    }
}