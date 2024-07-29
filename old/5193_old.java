        final int udpPort = pushReceiver.getUdpPort();
        String json = JacksonUtils.toJson(pack);
        final byte[] bytes = IoUtils.tryCompress(json, "UTF-8");
        final DatagramSocket datagramSocket = new DatagramSocket();
        datagramSocket.send(new DatagramPacket(bytes, bytes.length, InetAddress.getByName("localhost"), udpPort));
        byte[] buffer = new byte[20480];
        final DatagramPacket datagramPacket = new DatagramPacket(buffer, buffer.length);
        datagramSocket.receive(datagramPacket);
        final byte[] data = datagramPacket.getData();
        String res = new String(data, "UTF-8");
        return res.trim();
    }
    
    @Test
    public void testTestRunWithDump() throws InterruptedException, IOException {
        ServiceInfoHolder holder = Mockito.mock(ServiceInfoHolder.class);
        final PushReceiver pushReceiver = new PushReceiver(holder);
        final ExecutorService executorService = Executors.newFixedThreadPool(1);
        executorService.submit(new Runnable() {
            @Override
            public void run() {
                pushReceiver.run();
            }
        });
        TimeUnit.MILLISECONDS.sleep(10);
        
        PushReceiver.PushPacket pack1 = new PushReceiver.PushPacket();
        pack1.type = "dump";
        pack1.data = "pack1";
        pack1.lastRefTime = 1;
        final String res1 = udpClientRun(pack1, pushReceiver);
        Assert.assertEquals("{\"type\": \"dump-ack\", \"lastRefTime\": \"1\", \"data\":\"{}\"}", res1);
        verify(holder, times(1)).getServiceInfoMap();
    
    }
    
    @Test
    public void testTestRunWithUnknown() throws InterruptedException, IOException {
        ServiceInfoHolder holder = Mockito.mock(ServiceInfoHolder.class);
        final PushReceiver pushReceiver = new PushReceiver(holder);
        final ExecutorService executorService = Executors.newFixedThreadPool(1);
        executorService.submit(new Runnable() {
            @Override
            public void run() {
                pushReceiver.run();
            }
        });
        TimeUnit.MILLISECONDS.sleep(10);
        
        PushReceiver.PushPacket pack1 = new PushReceiver.PushPacket();
        pack1.type = "unknown";
        pack1.data = "pack1";
        pack1.lastRefTime = 1;
        final String res1 = udpClientRun(pack1, pushReceiver);
        Assert.assertEquals("{\"type\": \"unknown-ack\", \"lastRefTime\":\"1\", \"data\":\"\"}", res1);
    }
    
    @Test
    public void testShutdown() throws NacosException, NoSuchFieldException, IllegalAccessException {
        ServiceInfoHolder holder = Mockito.mock(ServiceInfoHolder.class);
        final PushReceiver pushReceiver = new PushReceiver(holder);
        
        pushReceiver.shutdown();
        
        final Field closed = PushReceiver.class.getDeclaredField("closed");
        closed.setAccessible(true);
        final boolean o = (boolean) closed.get(pushReceiver);
        Assert.assertTrue(o);
        
    }
    
    @Test
    public void testGetUdpPort() {
        ServiceInfoHolder holder = Mockito.mock(ServiceInfoHolder.class);
        final PushReceiver pushReceiver = new PushReceiver(holder);
        final int udpPort = pushReceiver.getUdpPort();
        System.out.println("udpPort = " + udpPort);
        Assert.assertTrue(udpPort > 0);
    }
}