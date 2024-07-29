        mos.write(colIndex+"", new LongWritable(count), new Text(keyBytes), "part_sort/"+colIndex);
    }

    private void logAFewRows(String value) {
        if(count<10){
            logger.info("key:{}, temp dict num :{},colIndex:{},colName:{}", value, count, colIndex, colName);
        }
    }

    @Override
    protected void doCleanup(Context context) throws IOException, InterruptedException {
        Configuration conf = context.getConfiguration();

        String partition = conf.get(MRJobConfig.TASK_PARTITION);
        mos.write(colIndex + "", new LongWritable(count), new Text(partition), "reduce_stats/" + colIndex);
        mos.close();
        logger.info("Reduce partition num {} finish,this reduce done item count is {}" , partition, count);
    }

}
