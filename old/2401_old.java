
        JobConfig cfg = new JobConfig().setName("mysql-monitor");
        JetInstance jet = Jet.bootstrappedInstance();
        jet.newJobIfAbsent(pipeline, cfg).join();
    }

}
